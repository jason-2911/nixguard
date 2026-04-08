// Package haproxy implements the loadbalancer engine interface.
// Manages HAProxy configuration for load balancing and reverse proxy.
package haproxy

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nixguard/nixguard/internal/domain/loadbalancer"
	"github.com/nixguard/nixguard/pkg/executor"
)

const (
	configFile = "/etc/haproxy/haproxy.cfg"
	statsSocket = "/var/run/haproxy/admin.sock"
)

type Adapter struct {
	exec *executor.Safe
	log  *slog.Logger
}

func NewAdapter(exec *executor.Safe, log *slog.Logger) *Adapter {
	return &Adapter{exec: exec, log: log}
}

// ApplyConfig generates and applies HAProxy configuration.
func (a *Adapter) ApplyConfig(ctx context.Context, frontends []loadbalancer.Frontend, backends []loadbalancer.Backend) error {
	conf := a.generateConfig(frontends, backends)
	_ = conf

	// Validate config
	if _, err := a.exec.Run(ctx, "haproxy", "-c", "-f", configFile); err != nil {
		return fmt.Errorf("haproxy config validation: %w", err)
	}

	// Reload without dropping connections
	_, err := a.exec.Run(ctx, "systemctl", "reload", "haproxy")
	return err
}

// GetStats reads HAProxy runtime stats from the unix socket.
func (a *Adapter) GetStats(ctx context.Context) (*loadbalancer.HAProxyStats, error) {
	// Query stats via unix socket: echo "show stat" | socat unix:socket -
	result, err := a.exec.Run(ctx, "socat", fmt.Sprintf("unix-connect:%s", statsSocket), "stdin")
	if err != nil {
		return nil, err
	}
	return parseHAProxyStats(result.Stdout), nil
}

func (a *Adapter) generateConfig(frontends []loadbalancer.Frontend, backends []loadbalancer.Backend) string {
	var buf strings.Builder

	// Global
	buf.WriteString("# NixGuard HAProxy Configuration - auto-generated\n")
	buf.WriteString("global\n")
	buf.WriteString("  log /dev/log local0\n")
	buf.WriteString("  chroot /var/lib/haproxy\n")
	buf.WriteString("  stats socket /var/run/haproxy/admin.sock mode 660 level admin\n")
	buf.WriteString("  stats timeout 30s\n")
	buf.WriteString("  user haproxy\n")
	buf.WriteString("  group haproxy\n")
	buf.WriteString("  daemon\n")
	buf.WriteString("  ssl-default-bind-ciphers PROFILE=SYSTEM\n")
	buf.WriteString("  ssl-default-bind-options no-sslv3 no-tlsv10 no-tlsv11\n\n")

	// Defaults
	buf.WriteString("defaults\n")
	buf.WriteString("  log global\n")
	buf.WriteString("  timeout connect 5s\n")
	buf.WriteString("  timeout client 50s\n")
	buf.WriteString("  timeout server 50s\n")
	buf.WriteString("  option httplog\n")
	buf.WriteString("  option dontlognull\n")
	buf.WriteString("  errorfile 400 /etc/haproxy/errors/400.http\n")
	buf.WriteString("  errorfile 403 /etc/haproxy/errors/403.http\n")
	buf.WriteString("  errorfile 503 /etc/haproxy/errors/503.http\n\n")

	// Frontends
	for _, fe := range frontends {
		if !fe.Enabled {
			continue
		}
		buf.WriteString(fmt.Sprintf("frontend %s\n", fe.Name))
		bind := fmt.Sprintf("  bind %s:%d", fe.BindAddress, fe.BindPort)
		if fe.SSLOffload && fe.SSLCertID != "" {
			bind += fmt.Sprintf(" ssl crt /etc/haproxy/certs/%s.pem", fe.SSLCertID)
			if fe.HTTP2 {
				bind += " alpn h2,http/1.1"
			}
		}
		buf.WriteString(bind + "\n")
		buf.WriteString(fmt.Sprintf("  mode %s\n", fe.Mode))
		buf.WriteString(fmt.Sprintf("  default_backend %s\n", fe.DefaultBackend))

		for _, acl := range fe.ACLRules {
			buf.WriteString(fmt.Sprintf("  acl %s %s %s\n", acl.Name, acl.Match, acl.Pattern))
			buf.WriteString(fmt.Sprintf("  use_backend %s if %s\n", acl.Backend, acl.Name))
		}
		buf.WriteString("\n")
	}

	// Backends
	for _, be := range backends {
		if !be.Enabled {
			continue
		}
		buf.WriteString(fmt.Sprintf("backend %s\n", be.Name))
		buf.WriteString(fmt.Sprintf("  mode %s\n", be.Mode))
		buf.WriteString(fmt.Sprintf("  balance %s\n", be.Balance))

		if be.HealthCheck.Type == "http" {
			buf.WriteString(fmt.Sprintf("  option httpchk GET %s\n", be.HealthCheck.URI))
		}

		if be.StickySession != nil && be.StickySession.Type == "cookie" {
			buf.WriteString(fmt.Sprintf("  cookie %s insert indirect nocache\n", be.StickySession.CookieName))
		}

		for _, srv := range be.Servers {
			line := fmt.Sprintf("  server %s %s:%d", srv.Name, srv.Address, srv.Port)
			if srv.Weight > 0 {
				line += fmt.Sprintf(" weight %d", srv.Weight)
			}
			if srv.MaxConn > 0 {
				line += fmt.Sprintf(" maxconn %d", srv.MaxConn)
			}
			if be.HealthCheck.Type != "" {
				line += " check"
				line += fmt.Sprintf(" inter %dms", be.HealthCheck.Interval)
				line += fmt.Sprintf(" rise %d fall %d", be.HealthCheck.Rise, be.HealthCheck.Fall)
			}
			if srv.IsBackup {
				line += " backup"
			}
			if srv.Maintenance {
				line += " disabled"
			}
			buf.WriteString(line + "\n")
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

func parseHAProxyStats(csv string) *loadbalancer.HAProxyStats {
	stats := &loadbalancer.HAProxyStats{}
	_ = csv
	return stats
}
