// Package iproute2 implements the network.NetworkEngine interface.
// Uses iproute2 commands (ip, bridge, ethtool) for Linux network management.
package iproute2

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nixguard/nixguard/internal/domain/network"
	"github.com/nixguard/nixguard/pkg/executor"
)

// Adapter implements network.NetworkEngine.
type Adapter struct {
	exec *executor.Safe
	log  *slog.Logger
}

func NewAdapter(exec *executor.Safe, log *slog.Logger) *Adapter {
	return &Adapter{exec: exec, log: log}
}

// ═══════════════════════════════════════════════════════════════
// Interface Operations
// ═══════════════════════════════════════════════════════════════

func (a *Adapter) GetInterfaceStatus(ctx context.Context, name string) (*network.InterfaceStatus, error) {
	// Use `ip -j link show <name>` for JSON output
	result, err := a.exec.Run(ctx, "ip", "-j", "-s", "link", "show", name)
	if err != nil {
		return nil, fmt.Errorf("ip link show %s: %w", name, err)
	}

	var links []ipLinkJSON
	if err := json.Unmarshal([]byte(result.Stdout), &links); err != nil {
		return nil, fmt.Errorf("parse ip link json: %w", err)
	}
	if len(links) == 0 {
		return nil, fmt.Errorf("interface %s not found", name)
	}

	link := links[0]
	status := &network.InterfaceStatus{
		OperState: link.OperState,
		Carrier:   link.Carrier == 1,
	}

	// Get addresses
	addrResult, err := a.exec.Run(ctx, "ip", "-j", "addr", "show", name)
	if err == nil {
		var addrs []ipAddrJSON
		if json.Unmarshal([]byte(addrResult.Stdout), &addrs) == nil && len(addrs) > 0 {
			for _, ai := range addrs[0].AddrInfo {
				status.Addresses = append(status.Addresses, fmt.Sprintf("%s/%d", ai.Local, ai.PrefixLen))
			}
		}
	}

	// Parse stats
	if link.Stats64 != nil {
		status.RxBytes = link.Stats64.RxBytes
		status.TxBytes = link.Stats64.TxBytes
		status.RxPackets = link.Stats64.RxPackets
		status.TxPackets = link.Stats64.TxPackets
		status.RxErrors = link.Stats64.RxErrors
		status.TxErrors = link.Stats64.TxErrors
		status.RxDropped = link.Stats64.RxDropped
		status.TxDropped = link.Stats64.TxDropped
	}

	// Get speed/duplex via ethtool
	ethResult, err := a.exec.Run(ctx, "ethtool", name)
	if err == nil {
		for _, line := range strings.Split(ethResult.Stdout, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Speed:") {
				status.Speed = strings.TrimSpace(strings.TrimPrefix(line, "Speed:"))
			} else if strings.HasPrefix(line, "Duplex:") {
				status.Duplex = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "Duplex:")))
			}
		}
	}

	return status, nil
}

func (a *Adapter) SetInterfaceUp(ctx context.Context, name string) error {
	_, err := a.exec.Run(ctx, "ip", "link", "set", name, "up")
	return err
}

func (a *Adapter) SetInterfaceDown(ctx context.Context, name string) error {
	_, err := a.exec.Run(ctx, "ip", "link", "set", name, "down")
	return err
}

func (a *Adapter) SetInterfaceAddress(ctx context.Context, name string, addr string) error {
	// Flush existing addresses first
	a.exec.Run(ctx, "ip", "addr", "flush", "dev", name)
	_, err := a.exec.Run(ctx, "ip", "addr", "add", addr, "dev", name)
	return err
}

func (a *Adapter) SetInterfaceMTU(ctx context.Context, name string, mtu int) error {
	_, err := a.exec.Run(ctx, "ip", "link", "set", name, "mtu", strconv.Itoa(mtu))
	return err
}

// ═══════════════════════════════════════════════════════════════
// VLAN Operations
// ═══════════════════════════════════════════════════════════════

func (a *Adapter) CreateVLAN(ctx context.Context, parent string, tag int) (string, error) {
	name := fmt.Sprintf("%s.%d", parent, tag)
	_, err := a.exec.Run(ctx, "ip", "link", "add", "link", parent, "name", name, "type", "vlan", "id", strconv.Itoa(tag))
	if err != nil {
		return "", fmt.Errorf("create vlan: %w", err)
	}
	a.exec.Run(ctx, "ip", "link", "set", name, "up")
	a.log.Info("VLAN created", slog.String("name", name), slog.Int("tag", tag))
	return name, nil
}

func (a *Adapter) DeleteVLAN(ctx context.Context, name string) error {
	_, err := a.exec.Run(ctx, "ip", "link", "del", name)
	return err
}

// ═══════════════════════════════════════════════════════════════
// Bond Operations
// ═══════════════════════════════════════════════════════════════

func (a *Adapter) CreateBond(ctx context.Context, name string, cfg network.BondConfig) error {
	mode := cfg.Mode
	if mode == "" {
		mode = "802.3ad"
	}
	_, err := a.exec.Run(ctx, "ip", "link", "add", name, "type", "bond", "mode", mode)
	if err != nil {
		return fmt.Errorf("create bond: %w", err)
	}

	for _, member := range cfg.Members {
		a.exec.Run(ctx, "ip", "link", "set", member, "down")
		if _, err := a.exec.Run(ctx, "ip", "link", "set", member, "master", name); err != nil {
			a.log.Warn("failed to add bond member", slog.String("member", member), slog.String("error", err.Error()))
		}
	}

	a.exec.Run(ctx, "ip", "link", "set", name, "up")
	a.log.Info("Bond created", slog.String("name", name), slog.String("mode", mode), slog.Int("members", len(cfg.Members)))
	return nil
}

func (a *Adapter) DeleteBond(ctx context.Context, name string) error {
	a.exec.Run(ctx, "ip", "link", "set", name, "down")
	_, err := a.exec.Run(ctx, "ip", "link", "del", name)
	return err
}

// ═══════════════════════════════════════════════════════════════
// Bridge Operations
// ═══════════════════════════════════════════════════════════════

func (a *Adapter) CreateBridge(ctx context.Context, name string, cfg network.BridgeConfig) error {
	_, err := a.exec.Run(ctx, "ip", "link", "add", name, "type", "bridge")
	if err != nil {
		return fmt.Errorf("create bridge: %w", err)
	}

	// STP
	stpVal := "0"
	if cfg.STP {
		stpVal = "1"
	}
	a.exec.Run(ctx, "ip", "link", "set", name, "type", "bridge", "stp_state", stpVal)

	for _, member := range cfg.Members {
		if _, err := a.exec.Run(ctx, "ip", "link", "set", member, "master", name); err != nil {
			a.log.Warn("failed to add bridge member", slog.String("member", member))
		}
	}

	a.exec.Run(ctx, "ip", "link", "set", name, "up")
	a.log.Info("Bridge created", slog.String("name", name), slog.Int("members", len(cfg.Members)))
	return nil
}

func (a *Adapter) DeleteBridge(ctx context.Context, name string) error {
	a.exec.Run(ctx, "ip", "link", "set", name, "down")
	_, err := a.exec.Run(ctx, "ip", "link", "del", name)
	return err
}

// ═══════════════════════════════════════════════════════════════
// Route Operations
// ═══════════════════════════════════════════════════════════════

func (a *Adapter) AddRoute(ctx context.Context, route network.Route) error {
	args := routeCommandArgs("add", route)
	_, err := a.exec.Run(ctx, "ip", args...)
	if err != nil {
		return fmt.Errorf("add route %s: %w", route.Destination, err)
	}
	a.log.Info("route added", slog.String("dest", route.Destination), slog.String("gw", route.Gateway))
	return nil
}

func (a *Adapter) DeleteRoute(ctx context.Context, route network.Route) error {
	args := routeCommandArgs("del", route)
	_, err := a.exec.Run(ctx, "ip", args...)
	return err
}

func (a *Adapter) ListRoutes(ctx context.Context, table int) ([]network.Route, error) {
	var out []network.Route
	for _, familyArgs := range [][]string{
		listRoutesArgs(false, table),
		listRoutesArgs(true, table),
	} {
		result, err := a.exec.Run(ctx, "ip", familyArgs...)
		if err != nil {
			return nil, err
		}

		var routes []ipRouteJSON
		if err := json.Unmarshal([]byte(result.Stdout), &routes); err != nil {
			return nil, fmt.Errorf("parse routes: %w", err)
		}

		for _, r := range routes {
			route := network.Route{
				Destination: r.Dst,
				Gateway:     r.Gateway,
				Interface:   r.Dev,
				Metric:      r.Metric,
				Type:        network.RouteDynamic,
			}
			if r.Protocol == "static" || r.Protocol == "boot" {
				route.Type = network.RouteStatic
			}
			out = append(out, route)
		}
	}
	return out, nil
}

// ═══════════════════════════════════════════════════════════════
// Gateway Monitoring
// ═══════════════════════════════════════════════════════════════

func (a *Adapter) PingGateway(ctx context.Context, addr string) (*network.GatewayStatus, error) {
	args := []string{"-c", "3", "-W", "2", "-q"}
	if strings.Contains(addr, ":") {
		args = append(args, "-6")
	}
	args = append(args, addr)

	result, err := a.exec.Run(ctx, "ping", args...)
	status := &network.GatewayStatus{
		State: "offline",
	}

	if err != nil {
		return status, nil
	}

	// Parse ping output
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		// "3 packets transmitted, 3 received, 0% packet loss, time 2003ms"
		if strings.Contains(line, "packet loss") {
			var sent, recv int
			var loss float64
			fmt.Sscanf(line, "%d packets transmitted, %d received, %f%% packet loss", &sent, &recv, &loss)
			status.PacketLoss = loss
			if recv > 0 {
				status.State = "online"
			}
		}
		// "rtt min/avg/max/mdev = 0.123/0.456/0.789/0.111 ms"
		if strings.Contains(line, "rtt min/avg/max") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				vals := strings.Split(strings.TrimSpace(parts[1]), "/")
				if len(vals) >= 2 {
					fmt.Sscanf(vals[1], "%f", &status.Latency)
				}
			}
		}
	}

	return status, nil
}

func (a *Adapter) CheckGatewayTCP(ctx context.Context, addr string, port int) (*network.GatewayStatus, error) {
	status := &network.GatewayStatus{State: "offline"}
	target := fmt.Sprintf("%s:%d", addr, port)

	start := time.Now()
	conn, err := net.DialTimeout("tcp", target, 5*time.Second)
	latency := time.Since(start)

	if err != nil {
		a.log.Debug("TCP check failed", slog.String("target", target), slog.String("error", err.Error()))
		return status, nil
	}
	conn.Close()

	status.State = "online"
	status.Latency = float64(latency.Milliseconds())
	status.PacketLoss = 0
	return status, nil
}

func (a *Adapter) CheckGatewayHTTP(ctx context.Context, url string) (*network.GatewayStatus, error) {
	status := &network.GatewayStatus{State: "offline"}

	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()
	resp, err := client.Get(url)
	latency := time.Since(start)

	if err != nil {
		a.log.Debug("HTTP check failed", slog.String("url", url), slog.String("error", err.Error()))
		return status, nil
	}
	resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		status.State = "online"
	}
	status.Latency = float64(latency.Milliseconds())
	status.PacketLoss = 0
	return status, nil
}

// ═══════════════════════════════════════════════════════════════
// Sysctl Operations
// ═══════════════════════════════════════════════════════════════

func (a *Adapter) SetSysctl(ctx context.Context, key, value string) error {
	_, err := a.exec.Run(ctx, "sysctl", "-w", fmt.Sprintf("%s=%s", key, value))
	return err
}

// ═══════════════════════════════════════════════════════════════
// Policy Routing
// ═══════════════════════════════════════════════════════════════

func (a *Adapter) AddPolicyRoute(ctx context.Context, mark uint32, gateway string, table int) error {
	// Add ip rule: packets with this mark use this routing table
	_, err := a.exec.Run(ctx, "ip", "rule", "add", "fwmark",
		fmt.Sprintf("0x%x", mark), "table", strconv.Itoa(table))
	if err != nil {
		return fmt.Errorf("add ip rule: %w", err)
	}

	// Add default route in that table via the specified gateway
	_, err = a.exec.Run(ctx, "ip", "route", "replace", "default",
		"via", gateway, "table", strconv.Itoa(table))
	if err != nil {
		return fmt.Errorf("add policy route: %w", err)
	}

	return nil
}

func (a *Adapter) DeletePolicyRoute(ctx context.Context, mark uint32, table int) error {
	a.exec.Run(ctx, "ip", "rule", "del", "fwmark",
		fmt.Sprintf("0x%x", mark), "table", strconv.Itoa(table))
	a.exec.Run(ctx, "ip", "route", "del", "default",
		"table", strconv.Itoa(table))
	return nil
}

func routeCommandArgs(action string, route network.Route) []string {
	if isIPv6Route(route) {
		return append(listRouteMutationArgs(true, action), routeMutationTail(route)...)
	}
	return append(listRouteMutationArgs(false, action), routeMutationTail(route)...)
}

func listRoutesArgs(ipv6 bool, table int) []string {
	args := []string{}
	if ipv6 {
		args = append(args, "-6")
	}
	args = append(args, "-j", "route", "show")
	if table > 0 && table != 254 {
		args = append(args, "table", strconv.Itoa(table))
	}
	return args
}

func listRouteMutationArgs(ipv6 bool, action string) []string {
	args := []string{}
	if ipv6 {
		args = append(args, "-6")
	}
	return append(args, "route", action)
}

func routeMutationTail(route network.Route) []string {
	args := []string{route.Destination}
	if route.Gateway != "" {
		args = append(args, "via", route.Gateway)
	}
	if route.Interface != "" {
		args = append(args, "dev", route.Interface)
	}
	if route.Metric > 0 {
		args = append(args, "metric", strconv.Itoa(route.Metric))
	}
	if route.Table > 0 && route.Table != 254 {
		args = append(args, "table", strconv.Itoa(route.Table))
	}
	return args
}

func isIPv6Route(route network.Route) bool {
	if strings.Contains(route.Destination, ":") || strings.Contains(route.Gateway, ":") {
		return true
	}
	return route.Destination == "default" && strings.Contains(route.Gateway, ":")
}

// ═══════════════════════════════════════════════════════════════
// Interface Discovery
// ═══════════════════════════════════════════════════════════════

// DiscoverInterfaces returns all system network interfaces with live status.
func (a *Adapter) DiscoverInterfaces(ctx context.Context) ([]network.Interface, error) {
	// Get all links
	result, err := a.exec.Run(ctx, "ip", "-j", "-s", "link", "show")
	if err != nil {
		return nil, fmt.Errorf("ip link show: %w", err)
	}

	var links []ipLinkJSON
	if err := json.Unmarshal([]byte(result.Stdout), &links); err != nil {
		return nil, fmt.Errorf("parse ip link json: %w", err)
	}

	// Get all addresses in one call
	addrResult, err := a.exec.Run(ctx, "ip", "-j", "addr", "show")
	addrMap := make(map[string][]string) // ifname → []addresses
	if err == nil {
		var allAddrs []struct {
			IfName   string `json:"ifname"`
			AddrInfo []struct {
				Local     string `json:"local"`
				PrefixLen int    `json:"prefixlen"`
			} `json:"addr_info"`
		}
		if json.Unmarshal([]byte(addrResult.Stdout), &allAddrs) == nil {
			for _, a := range allAddrs {
				for _, ai := range a.AddrInfo {
					addrMap[a.IfName] = append(addrMap[a.IfName], fmt.Sprintf("%s/%d", ai.Local, ai.PrefixLen))
				}
			}
		}
	}

	var ifaces []network.Interface
	for _, link := range links {
		ifType := detectInterfaceType(link.IfName, link.LinkType)

		status := network.InterfaceStatus{
			OperState: link.OperState,
			Carrier:   link.Carrier == 1,
			Addresses: addrMap[link.IfName],
		}
		if link.Stats64 != nil {
			status.RxBytes = link.Stats64.RxBytes
			status.TxBytes = link.Stats64.TxBytes
			status.RxPackets = link.Stats64.RxPackets
			status.TxPackets = link.Stats64.TxPackets
			status.RxErrors = link.Stats64.RxErrors
			status.TxErrors = link.Stats64.TxErrors
			status.RxDropped = link.Stats64.RxDropped
			status.TxDropped = link.Stats64.TxDropped
		}

		// Try ethtool for speed/duplex (best effort)
		if ethResult, err := a.exec.Run(ctx, "ethtool", link.IfName); err == nil {
			for _, line := range strings.Split(ethResult.Stdout, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "Speed:") {
					status.Speed = strings.TrimSpace(strings.TrimPrefix(line, "Speed:"))
				} else if strings.HasPrefix(line, "Duplex:") {
					status.Duplex = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "Duplex:")))
				}
			}
		}

		iface := network.Interface{
			ID:      link.IfName, // use name as ID for discovered interfaces
			Name:    link.IfName,
			Type:    ifType,
			Enabled: link.OperState == "UP" || link.OperState == "UNKNOWN",
			MTU:     link.MTU,
			MAC:     link.Address,
			Status:  status,
		}
		ifaces = append(ifaces, iface)
	}

	return ifaces, nil
}

func detectInterfaceType(name, linkType string) network.InterfaceType {
	switch {
	case name == "lo" || linkType == "loopback":
		return network.IfTypeLoopback
	case strings.HasPrefix(name, "br") || strings.HasPrefix(name, "docker") || linkType == "bridge":
		return network.IfTypeBridge
	case strings.HasPrefix(name, "bond"):
		return network.IfTypeBond
	case strings.Contains(name, "."):
		return network.IfTypeVLAN
	case strings.HasPrefix(name, "vxlan"):
		return network.IfTypeVXLAN
	case strings.HasPrefix(name, "wg"):
		return network.IfTypePhysical
	default:
		return network.IfTypePhysical
	}
}

// ═══════════════════════════════════════════════════════════════
// JSON parse types for ip -j output
// ═══════════════════════════════════════════════════════════════

type ipLinkJSON struct {
	IfName    string       `json:"ifname"`
	OperState string       `json:"operstate"`
	Carrier   int          `json:"carrier"`
	MTU       int          `json:"mtu"`
	Address   string       `json:"address"`
	LinkType  string       `json:"link_type"`
	Stats64   *linkStats64 `json:"stats64"`
}

type linkStats64 struct {
	RxBytes   uint64 `json:"rx_bytes"`
	TxBytes   uint64 `json:"tx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	TxPackets uint64 `json:"tx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	TxErrors  uint64 `json:"tx_errors"`
	RxDropped uint64 `json:"rx_dropped"`
	TxDropped uint64 `json:"tx_dropped"`
}

type ipAddrJSON struct {
	AddrInfo []struct {
		Local     string `json:"local"`
		PrefixLen int    `json:"prefixlen"`
		Family    string `json:"family"`
	} `json:"addr_info"`
}

type ipRouteJSON struct {
	Dst      string `json:"dst"`
	Gateway  string `json:"gateway"`
	Dev      string `json:"dev"`
	Protocol string `json:"protocol"`
	Metric   int    `json:"metric"`
	Scope    string `json:"scope"`
}

var _ network.NetworkEngine = (*Adapter)(nil)
