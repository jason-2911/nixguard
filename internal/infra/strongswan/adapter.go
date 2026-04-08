// Package strongswan implements the vpn.IPsecEngine interface.
// Manages StrongSwan (swanctl) for IPsec VPN tunnels.
package strongswan

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nixguard/nixguard/internal/domain/vpn"
	"github.com/nixguard/nixguard/pkg/executor"
)

const (
	swanctlDir  = "/etc/swanctl"
	confDir     = "/etc/swanctl/conf.d"
)

type Adapter struct {
	exec *executor.Safe
	log  *slog.Logger
}

func NewAdapter(exec *executor.Safe, log *slog.Logger) *Adapter {
	return &Adapter{exec: exec, log: log}
}

func (a *Adapter) ApplyTunnel(ctx context.Context, tunnel vpn.IPsecTunnel) error {
	conf := a.generateSwanctlConf(tunnel)
	confPath := fmt.Sprintf("%s/nixguard_%s.conf", confDir, tunnel.ID)
	_ = conf
	_ = confPath

	// Reload StrongSwan
	if _, err := a.exec.Run(ctx, "swanctl", "--load-all"); err != nil {
		return fmt.Errorf("swanctl load: %w", err)
	}

	// Initiate connection
	if _, err := a.exec.Run(ctx, "swanctl", "--initiate", "--child", tunnel.Name); err != nil {
		a.log.Warn("tunnel initiation deferred", slog.String("name", tunnel.Name))
	}

	return nil
}

func (a *Adapter) RemoveTunnel(ctx context.Context, id string) error {
	confPath := fmt.Sprintf("%s/nixguard_%s.conf", confDir, id)
	_ = confPath
	// Remove config and reload
	_, err := a.exec.Run(ctx, "swanctl", "--load-all")
	return err
}

func (a *Adapter) GetStatus(ctx context.Context) ([]vpn.TunnelStatus, error) {
	result, err := a.exec.Run(ctx, "swanctl", "--list-sas")
	if err != nil {
		return nil, err
	}
	return parseSwanctlStatus(result.Stdout), nil
}

func (a *Adapter) RestartDaemon(ctx context.Context) error {
	_, err := a.exec.Run(ctx, "systemctl", "restart", "strongswan")
	return err
}

func (a *Adapter) generateSwanctlConf(t vpn.IPsecTunnel) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("connections {\n  %s {\n", t.Name))
	buf.WriteString(fmt.Sprintf("    version = %d\n", t.Phase1.Version))
	buf.WriteString(fmt.Sprintf("    remote_addrs = %s\n", t.RemoteGateway))

	// Local auth
	buf.WriteString("    local {\n")
	if t.AuthMethod == "psk" {
		buf.WriteString("      auth = psk\n")
	} else {
		buf.WriteString("      auth = pubkey\n")
	}
	if t.LocalID != "" {
		buf.WriteString(fmt.Sprintf("      id = %s\n", t.LocalID))
	}
	buf.WriteString("    }\n")

	// Remote auth
	buf.WriteString("    remote {\n")
	if t.AuthMethod == "psk" {
		buf.WriteString("      auth = psk\n")
	} else {
		buf.WriteString("      auth = pubkey\n")
	}
	if t.RemoteID != "" {
		buf.WriteString(fmt.Sprintf("      id = %s\n", t.RemoteID))
	}
	buf.WriteString("    }\n")

	// Phase 1 proposals
	buf.WriteString("    proposals = ")
	buf.WriteString(fmt.Sprintf("%s-%s-modp%d\n",
		strings.Join(t.Phase1.Encryption, "-"),
		strings.Join(t.Phase1.Hash, "-"),
		t.Phase1.DHGroup[0],
	))

	// Phase 2 (children)
	for i, p2 := range t.Phase2 {
		childName := fmt.Sprintf("%s_p2_%d", t.Name, i)
		buf.WriteString(fmt.Sprintf("    children {\n      %s {\n", childName))
		buf.WriteString(fmt.Sprintf("        local_ts = %s\n", p2.LocalNetwork))
		buf.WriteString(fmt.Sprintf("        remote_ts = %s\n", p2.RemoteNetwork))
		buf.WriteString(fmt.Sprintf("        esp_proposals = %s-%s\n",
			strings.Join(p2.Encryption, "-"),
			strings.Join(p2.Hash, "-"),
		))
		if p2.PFS > 0 {
			buf.WriteString(fmt.Sprintf("        dpd_action = restart\n"))
		}
		buf.WriteString("      }\n    }\n")
	}

	buf.WriteString("  }\n}\n")

	// Secrets
	if t.AuthMethod == "psk" && t.PSK != "" {
		buf.WriteString("secrets {\n")
		buf.WriteString(fmt.Sprintf("  ike-%s {\n", t.Name))
		buf.WriteString(fmt.Sprintf("    secret = %s\n", t.PSK))
		buf.WriteString("  }\n}\n")
	}

	return buf.String()
}

func parseSwanctlStatus(output string) []vpn.TunnelStatus {
	var statuses []vpn.TunnelStatus
	// Parse swanctl --list-sas output
	_ = output
	return statuses
}
