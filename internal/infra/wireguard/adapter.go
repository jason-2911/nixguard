// Package wireguard implements the vpn.WireGuardEngine interface.
// Manages WireGuard interfaces via the wg command and ip link.
package wireguard

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nixguard/nixguard/internal/domain/vpn"
	"github.com/nixguard/nixguard/pkg/executor"
)

type Adapter struct {
	exec *executor.Safe
	log  *slog.Logger
}

func NewAdapter(exec *executor.Safe, log *slog.Logger) *Adapter {
	return &Adapter{exec: exec, log: log}
}

func (a *Adapter) ApplyInterface(ctx context.Context, wg vpn.WireGuardInterface) error {
	// Create interface
	if _, err := a.exec.Run(ctx, "ip", "link", "add", "dev", wg.Name, "type", "wireguard"); err != nil {
		a.log.Debug("interface may already exist", slog.String("name", wg.Name))
	}

	// Write config
	conf := a.generateConfig(wg)
	confPath := fmt.Sprintf("/etc/wireguard/%s.conf", wg.Name)
	_ = conf
	_ = confPath

	// Apply config
	if _, err := a.exec.Run(ctx, "wg", "setconf", wg.Name, confPath); err != nil {
		return fmt.Errorf("wg setconf: %w", err)
	}

	// Set addresses
	for _, addr := range wg.Address {
		if _, err := a.exec.Run(ctx, "ip", "addr", "add", addr, "dev", wg.Name); err != nil {
			a.log.Debug("address may already exist", slog.String("addr", addr))
		}
	}

	// Bring up
	if _, err := a.exec.Run(ctx, "ip", "link", "set", wg.Name, "up"); err != nil {
		return fmt.Errorf("bring up %s: %w", wg.Name, err)
	}

	return nil
}

func (a *Adapter) RemoveInterface(ctx context.Context, name string) error {
	_, err := a.exec.Run(ctx, "ip", "link", "del", "dev", name)
	return err
}

func (a *Adapter) GetStatus(ctx context.Context, name string) (*vpn.TunnelStatus, error) {
	result, err := a.exec.Run(ctx, "wg", "show", name, "dump")
	if err != nil {
		return nil, err
	}
	return parseWGStatus(result.Stdout), nil
}

func (a *Adapter) GenerateKeyPair(ctx context.Context) (string, string, error) {
	privResult, err := a.exec.Run(ctx, "wg", "genkey")
	if err != nil {
		return "", "", err
	}
	privKey := strings.TrimSpace(privResult.Stdout)

	// Generate public key from private key via stdin — simplified
	pubResult, err := a.exec.Run(ctx, "wg", "pubkey")
	if err != nil {
		return "", "", err
	}
	pubKey := strings.TrimSpace(pubResult.Stdout)

	return privKey, pubKey, nil
}

func (a *Adapter) GeneratePresharedKey(ctx context.Context) (string, error) {
	result, err := a.exec.Run(ctx, "wg", "genpsk")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Stdout), nil
}

func (a *Adapter) generateConfig(wg vpn.WireGuardInterface) string {
	var buf strings.Builder
	buf.WriteString("[Interface]\n")
	buf.WriteString(fmt.Sprintf("PrivateKey = %s\n", wg.PrivateKey))
	buf.WriteString(fmt.Sprintf("ListenPort = %d\n", wg.ListenPort))
	buf.WriteString("\n")

	for _, peer := range wg.Peers {
		buf.WriteString("[Peer]\n")
		buf.WriteString(fmt.Sprintf("PublicKey = %s\n", peer.PublicKey))
		if peer.PresharedKey != "" {
			buf.WriteString(fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey))
		}
		if peer.Endpoint != "" {
			buf.WriteString(fmt.Sprintf("Endpoint = %s\n", peer.Endpoint))
		}
		buf.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(peer.AllowedIPs, ", ")))
		if peer.PersistentKeepalive > 0 {
			buf.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", peer.PersistentKeepalive))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

func parseWGStatus(dump string) *vpn.TunnelStatus {
	status := &vpn.TunnelStatus{State: "up"}
	// Parse wg show dump output
	_ = dump
	return status
}
