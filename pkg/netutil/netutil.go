// Package netutil provides network utility functions for NixGuard.
package netutil

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// ListInterfaces returns all network interfaces on the system.
func ListInterfaces() ([]InterfaceInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list interfaces: %w", err)
	}

	var result []InterfaceInfo
	for _, iface := range ifaces {
		info := InterfaceInfo{
			Name:    iface.Name,
			Index:   iface.Index,
			MTU:     iface.MTU,
			MAC:     iface.HardwareAddr.String(),
			Flags:   iface.Flags.String(),
			IsUp:    iface.Flags&net.FlagUp != 0,
			IsLoopback: iface.Flags&net.FlagLoopback != 0,
		}

		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				info.Addresses = append(info.Addresses, addr.String())
			}
		}

		result = append(result, info)
	}
	return result, nil
}

// InterfaceInfo holds network interface details.
type InterfaceInfo struct {
	Name       string   `json:"name"`
	Index      int      `json:"index"`
	MTU        int      `json:"mtu"`
	MAC        string   `json:"mac"`
	Flags      string   `json:"flags"`
	IsUp       bool     `json:"is_up"`
	IsLoopback bool     `json:"is_loopback"`
	Addresses  []string `json:"addresses"`
}

// GetDefaultGateway reads the default gateway from /proc/net/route.
func GetDefaultGateway() (string, string, error) {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return "", "", fmt.Errorf("read route table: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == "00000000" {
			// Convert hex gateway to IP
			gw := hexToIP(fields[2])
			return fields[0], gw, nil
		}
	}
	return "", "", fmt.Errorf("no default gateway found")
}

func hexToIP(hex string) string {
	if len(hex) != 8 {
		return ""
	}
	// Linux stores in little-endian
	return fmt.Sprintf("%d.%d.%d.%d",
		hexByte(hex[6:8]),
		hexByte(hex[4:6]),
		hexByte(hex[2:4]),
		hexByte(hex[0:2]),
	)
}

func hexByte(s string) uint8 {
	var b uint8
	fmt.Sscanf(s, "%x", &b)
	return b
}

// IsPrivateIP checks if an IP is in RFC1918 private range.
func IsPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7",
	}
	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// IsBogonIP checks if an IP is a bogon (should not appear on public internet).
func IsBogonIP(ip net.IP) bool {
	bogonRanges := []string{
		"0.0.0.0/8",
		"10.0.0.0/8",
		"100.64.0.0/10",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"192.168.0.0/16",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"224.0.0.0/4",
		"240.0.0.0/4",
	}
	for _, cidr := range bogonRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
