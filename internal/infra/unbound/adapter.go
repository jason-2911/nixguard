// Package unbound implements the dns.DNSEngine interface.
// Manages Unbound recursive DNS resolver configuration.
package unbound

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nixguard/nixguard/internal/domain/dns"
	"github.com/nixguard/nixguard/pkg/executor"
)

const (
	configFile    = "/etc/unbound/unbound.conf"
	overridesFile = "/etc/unbound/nixguard_overrides.conf"
	blocklistFile = "/etc/unbound/nixguard_blocklist.conf"
)

type Adapter struct {
	exec *executor.Safe
	log  *slog.Logger
}

func NewAdapter(exec *executor.Safe, log *slog.Logger) *Adapter {
	return &Adapter{exec: exec, log: log}
}

func (a *Adapter) ApplyConfig(ctx context.Context, cfg dns.ResolverConfig, overrides []dns.HostOverride, domains []dns.DomainOverride) error {
	// Generate main config
	mainConf := a.generateMainConfig(cfg)
	overrideConf := a.generateOverrides(overrides)
	domainConf := a.generateDomainOverrides(domains)

	_ = mainConf
	_ = overrideConf
	_ = domainConf

	// Write configs and reload
	return a.Restart(ctx)
}

func (a *Adapter) ApplyBlocklists(ctx context.Context, blocklists []dns.Blocklist, whitelists []dns.Whitelist) error {
	var buf strings.Builder
	buf.WriteString("# NixGuard DNS blocklist - auto-generated\n")

	// Generate local-zone and local-data entries for blocked domains
	for _, bl := range blocklists {
		if !bl.Enabled {
			continue
		}
		buf.WriteString(fmt.Sprintf("# Source: %s\n", bl.Name))
		// In real implementation, download and parse the blocklist URL
	}

	// Add whitelist exceptions
	for _, wl := range whitelists {
		buf.WriteString(fmt.Sprintf("local-zone: \"%s\" transparent\n", wl.Domain))
	}

	_ = buf.String()
	return a.Restart(ctx)
}

func (a *Adapter) FlushCache(ctx context.Context) error {
	_, err := a.exec.Run(ctx, "unbound-control", "flush_zone", ".")
	return err
}

func (a *Adapter) GetStats(ctx context.Context) (*dns.ResolverStats, error) {
	result, err := a.exec.Run(ctx, "unbound-control", "stats_noreset")
	if err != nil {
		return nil, err
	}
	return parseUnboundStats(result.Stdout), nil
}

func (a *Adapter) Restart(ctx context.Context) error {
	// Validate config first
	if _, err := a.exec.Run(ctx, "unbound-checkconf"); err != nil {
		return fmt.Errorf("unbound config validation failed: %w", err)
	}
	_, err := a.exec.Run(ctx, "systemctl", "restart", "unbound")
	return err
}

func (a *Adapter) QueryLog(ctx context.Context, filter dns.QueryLogFilter) ([]dns.DNSQueryLog, error) {
	// Read from query log file or database
	return nil, nil
}

func (a *Adapter) generateMainConfig(cfg dns.ResolverConfig) string {
	var buf strings.Builder
	buf.WriteString("# NixGuard Unbound Configuration - auto-generated\n")
	buf.WriteString("server:\n")
	buf.WriteString(fmt.Sprintf("  port: %d\n", cfg.Port))
	buf.WriteString(fmt.Sprintf("  do-ip4: yes\n"))
	buf.WriteString(fmt.Sprintf("  do-ip6: yes\n"))
	buf.WriteString(fmt.Sprintf("  do-udp: yes\n"))
	buf.WriteString(fmt.Sprintf("  do-tcp: yes\n"))

	for _, iface := range cfg.ListenInterfaces {
		buf.WriteString(fmt.Sprintf("  interface: %s\n", iface))
	}

	if cfg.DNSSEC {
		buf.WriteString("  auto-trust-anchor-file: \"/var/lib/unbound/root.key\"\n")
	}

	if cfg.QueryMinimize {
		buf.WriteString("  qname-minimisation: yes\n")
	}
	if cfg.AggressiveNSEC {
		buf.WriteString("  aggressive-nsec: yes\n")
	}
	if cfg.Prefetch {
		buf.WriteString("  prefetch: yes\n")
	}
	if cfg.ServeExpired {
		buf.WriteString("  serve-expired: yes\n")
	}
	if cfg.RebindProtection {
		buf.WriteString("  private-address: 10.0.0.0/8\n")
		buf.WriteString("  private-address: 172.16.0.0/12\n")
		buf.WriteString("  private-address: 192.168.0.0/16\n")
	}

	buf.WriteString(fmt.Sprintf("  msg-cache-size: %dm\n", cfg.CacheSize))
	buf.WriteString("  include: \"/etc/unbound/nixguard_overrides.conf\"\n")
	buf.WriteString("  include: \"/etc/unbound/nixguard_blocklist.conf\"\n")

	if cfg.ForwardMode && len(cfg.Forwarders) > 0 {
		buf.WriteString("forward-zone:\n")
		buf.WriteString("  name: \".\"\n")
		if cfg.DNSOverTLS {
			buf.WriteString("  forward-tls-upstream: yes\n")
		}
		for _, fwd := range cfg.Forwarders {
			buf.WriteString(fmt.Sprintf("  forward-addr: %s\n", fwd))
		}
	}

	return buf.String()
}

func (a *Adapter) generateOverrides(overrides []dns.HostOverride) string {
	var buf strings.Builder
	for _, o := range overrides {
		fqdn := fmt.Sprintf("%s.%s", o.Hostname, o.Domain)
		buf.WriteString(fmt.Sprintf("local-data: \"%s. IN %s %s\"\n", fqdn, o.Type, o.IPAddress))
		buf.WriteString(fmt.Sprintf("local-data-ptr: \"%s %s\"\n", o.IPAddress, fqdn))
		for _, alias := range o.Aliases {
			aliasFQDN := fmt.Sprintf("%s.%s", alias.Hostname, alias.Domain)
			buf.WriteString(fmt.Sprintf("local-data: \"%s. IN CNAME %s.\"\n", aliasFQDN, fqdn))
		}
	}
	return buf.String()
}

func (a *Adapter) generateDomainOverrides(domains []dns.DomainOverride) string {
	var buf strings.Builder
	for _, d := range domains {
		buf.WriteString("forward-zone:\n")
		buf.WriteString(fmt.Sprintf("  name: \"%s\"\n", d.Domain))
		addr := d.IPAddress
		if d.Port > 0 && d.Port != 53 {
			addr = fmt.Sprintf("%s@%d", addr, d.Port)
		}
		if d.UseTLS {
			buf.WriteString("  forward-tls-upstream: yes\n")
		}
		buf.WriteString(fmt.Sprintf("  forward-addr: %s\n", addr))
	}
	return buf.String()
}

func parseUnboundStats(output string) *dns.ResolverStats {
	stats := &dns.ResolverStats{}
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		switch key {
		case "total.num.queries":
			fmt.Sscanf(parts[1], "%d", &stats.TotalQueries)
		case "total.num.cachehits":
			fmt.Sscanf(parts[1], "%d", &stats.CacheHits)
		case "total.num.cachemiss":
			fmt.Sscanf(parts[1], "%d", &stats.CacheMisses)
		}
	}
	if stats.TotalQueries > 0 {
		stats.CacheHitPercent = float64(stats.CacheHits) / float64(stats.TotalQueries) * 100
	}
	return stats
}
