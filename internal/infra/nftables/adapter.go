// Package nftables implements the firewall.FirewallEngine interface using nftables.
// It compiles NixGuard firewall rules into nft ruleset and applies atomically.
//
// Architecture:
//
//	NixGuard Rule Model → compiler.go → nft ruleset string → agent executor → kernel netfilter
//
// The adapter generates a complete nftables ruleset and applies it via
// the privileged agent using "nft -f" for atomic rule replacement.
package nftables

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nixguard/nixguard/internal/domain/firewall"
	"github.com/nixguard/nixguard/pkg/executor"
)

const (
	rulesetPath = "./data/nftables/nixguard.nft"
	natPath     = "./data/nftables/nixguard_nat.nft"
	pcapDir     = "./data/pcap"
)

// Adapter implements firewall.FirewallEngine using nftables.
type Adapter struct {
	exec *executor.Safe
	log  *slog.Logger
}

// NewAdapter creates an nftables adapter.
func NewAdapter(exec *executor.Safe, log *slog.Logger) *Adapter {
	return &Adapter{exec: exec, log: log}
}

// ApplyRules generates and applies a complete nftables ruleset atomically.
func (a *Adapter) ApplyRules(ctx context.Context, rules []firewall.Rule, aliases []firewall.Alias) error {
	ruleset := CompileRuleset(rules, aliases, nil)

	a.log.Info("applying nftables ruleset",
		slog.Int("rules", len(rules)),
		slog.Int("aliases", len(aliases)),
		slog.Int("ruleset_bytes", len(ruleset)),
	)

	// Write to file, then apply atomically
	if err := writeRulesetFile(rulesetPath, ruleset); err != nil {
		return fmt.Errorf("write ruleset: %w", err)
	}

	_, err := a.exec.Run(ctx, "nft", "-f", rulesetPath)
	if err != nil {
		if isPrivilegeError(err) {
			a.log.Warn("nftables apply failed — needs root privileges",
				slog.String("ruleset_file", rulesetPath),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("apply nftables ruleset: insufficient privileges (run nixguard-agent as root): %w", err)
		}
		return fmt.Errorf("apply nftables ruleset: %w", err)
	}

	a.log.Info("nftables ruleset applied successfully")
	return nil
}

// ApplyNAT generates and applies NAT rules atomically.
func (a *Adapter) ApplyNAT(ctx context.Context, rules []firewall.NATRule) error {
	ruleset := CompileNATRuleset(rules)

	a.log.Info("applying NAT ruleset", slog.Int("rules", len(rules)))

	if err := writeRulesetFile(natPath, ruleset); err != nil {
		return fmt.Errorf("write NAT ruleset: %w", err)
	}

	_, err := a.exec.Run(ctx, "nft", "-f", natPath)
	if err != nil {
		if isPrivilegeError(err) {
			a.log.Warn("NAT apply failed — needs root privileges",
				slog.String("ruleset_file", natPath),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("apply NAT ruleset: insufficient privileges (run nixguard-agent as root): %w", err)
		}
		return fmt.Errorf("apply NAT ruleset: %w", err)
	}

	return nil
}

// CaptureTraffic returns a bounded tcpdump snapshot for the live traffic page.
func (a *Adapter) CaptureTraffic(ctx context.Context, filter firewall.TrafficFilter) ([]firewall.CapturedPacket, error) {
	args := buildTCPDumpArgs(filter, "")
	result, err := a.exec.Run(ctx, "tcpdump", args...)
	if err != nil {
		return nil, fmt.Errorf("tcpdump capture: %w", err)
	}
	return parseTCPDumpOutput(result.Stdout, filter.Interface), nil
}

// ExportPCAP captures packets to a pcap file for download from the GUI.
func (a *Adapter) ExportPCAP(ctx context.Context, filter firewall.TrafficFilter) (*firewall.PCAPExport, error) {
	if err := os.MkdirAll(pcapDir, 0750); err != nil {
		return nil, fmt.Errorf("create pcap dir: %w", err)
	}

	name := fmt.Sprintf("capture_%s.pcap", time.Now().UTC().Format("20060102T150405Z"))
	path := filepath.Join(pcapDir, name)

	args := buildTCPDumpArgs(filter, path)
	if _, err := a.exec.Run(ctx, "tcpdump", args...); err != nil {
		return nil, fmt.Errorf("tcpdump export: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat pcap: %w", err)
	}

	return &firewall.PCAPExport{
		Name:        name,
		DownloadURL: "/api/v1/firewall/traffic/export/" + name,
		Bytes:       info.Size(),
		CreatedAt:   info.ModTime().UTC(),
	}, nil
}

// GetStates returns connection tracking entries via conntrack.
func (a *Adapter) GetStates(ctx context.Context, filter firewall.StateFilter) ([]firewall.State, error) {
	args := []string{"-L", "-o", "extended"}
	if filter.Protocol != "" {
		args = append(args, "-p", filter.Protocol)
	}
	if filter.SourceIP != "" {
		args = append(args, "-s", filter.SourceIP)
	}
	if filter.DestIP != "" {
		args = append(args, "-d", filter.DestIP)
	}

	result, err := a.exec.Run(ctx, "conntrack", args...)
	if err != nil {
		// conntrack may return error if no entries
		if result != nil && result.ExitCode == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("conntrack: %w", err)
	}

	states := ParseConntrack(result.Stdout)

	// Apply limit
	if filter.Limit > 0 && len(states) > filter.Limit {
		states = states[:filter.Limit]
	}

	return states, nil
}

// FlushStates removes connection tracking entries.
func (a *Adapter) FlushStates(ctx context.Context, filter firewall.StateFilter) error {
	if filter.SourceIP == "" && filter.DestIP == "" && filter.Protocol == "" {
		// Flush all
		_, err := a.exec.Run(ctx, "conntrack", "-F")
		return err
	}

	args := []string{"-D"}
	if filter.SourceIP != "" {
		args = append(args, "-s", filter.SourceIP)
	}
	if filter.DestIP != "" {
		args = append(args, "-d", filter.DestIP)
	}
	if filter.Protocol != "" {
		args = append(args, "-p", filter.Protocol)
	}

	_, err := a.exec.Run(ctx, "conntrack", args...)
	return err
}

// GetRuleStats queries nftables counters for each rule.
func (a *Adapter) GetRuleStats(ctx context.Context) (map[string]firewall.RuleStats, error) {
	result, err := a.exec.Run(ctx, "nft", "-j", "list", "counters")
	if err != nil {
		return nil, fmt.Errorf("list nft counters: %w", err)
	}

	return parseNFTJSONCounters(result.Stdout), nil
}

// Verify checks that nftables is available.
func (a *Adapter) Verify(ctx context.Context) error {
	_, err := a.exec.Run(ctx, "nft", "--version")
	return err
}

// ─── Helpers ───────────────────────────────────────────────────

func writeRulesetFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0640)
}

func buildTCPDumpArgs(filter firewall.TrafficFilter, outputPath string) []string {
	iface := filter.Interface
	if iface == "" {
		iface = "any"
	}

	count := filter.Count
	if count <= 0 {
		count = 50
	}
	if count > 500 {
		count = 500
	}

	snapLen := filter.SnapLen
	if snapLen <= 0 {
		snapLen = 160
	}

	args := []string{"-nn", "-tt", "-c", strconv.Itoa(count), "-s", strconv.Itoa(snapLen), "-i", iface}
	if outputPath == "" {
		args = append(args, "-l")
	} else {
		args = append(args, "-w", outputPath)
	}

	if expr := buildBPFExpression(filter); expr != "" {
		args = append(args, strings.Fields(expr)...)
	}

	return args
}

func buildBPFExpression(filter firewall.TrafficFilter) string {
	var parts []string

	if filter.Protocol != "" && filter.Protocol != "any" {
		parts = append(parts, filter.Protocol)
	}
	if filter.SourceIP != "" {
		parts = append(parts, "src host "+filter.SourceIP)
	}
	if filter.DestIP != "" {
		parts = append(parts, "dst host "+filter.DestIP)
	}

	return strings.Join(parts, " and ")
}

func parseTCPDumpOutput(output string, iface string) []firewall.CapturedPacket {
	var packets []firewall.CapturedPacket

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		packets = append(packets, parseTCPDumpLine(line, iface))
	}

	return packets
}

func parseTCPDumpLine(line, iface string) firewall.CapturedPacket {
	verdict := "observed"
	switch {
	case strings.Contains(line, "nixguard_block_"):
		verdict = "blocked"
	case strings.Contains(line, "nixguard_reject_"):
		verdict = "rejected"
	case strings.Contains(line, "nixguard_pass_"):
		verdict = "passed"
	}

	packet := firewall.CapturedPacket{
		Interface: iface,
		Verdict:   verdict,
		Summary:   line,
		Detail:    line,
	}

	fields := strings.Fields(line)
	if len(fields) > 0 {
		packet.Timestamp = fields[0]
	}

	switch {
	case strings.Contains(line, "Flags ["), strings.Contains(line, " tcp "):
		packet.Protocol = "tcp"
	case strings.Contains(line, "UDP"), strings.Contains(line, " udp "):
		packet.Protocol = "udp"
	case strings.Contains(line, "ICMP6"), strings.Contains(line, "icmp6"):
		packet.Protocol = "icmpv6"
	case strings.Contains(line, "ICMP"), strings.Contains(line, "icmp"):
		packet.Protocol = "icmp"
	case strings.Contains(line, "IP6 "):
		packet.Protocol = "ipv6"
	case strings.Contains(line, "IP "):
		packet.Protocol = "ipv4"
	}

	if idx := strings.Index(line, " > "); idx != -1 {
		left := strings.TrimSpace(line[:idx])
		right := line[idx+3:]

		leftFields := strings.Fields(left)
		if len(leftFields) > 0 {
			packet.Source = leftFields[len(leftFields)-1]
		}

		if colon := strings.Index(right, ":"); colon != -1 {
			packet.Destination = strings.TrimSpace(right[:colon])
		}
	}

	lengthRe := regexp.MustCompile(`length\s+(\d+)`)
	if matches := lengthRe.FindStringSubmatch(line); len(matches) == 2 {
		if n, err := strconv.Atoi(matches[1]); err == nil {
			packet.Length = n
		}
	}

	return packet
}

// parseNFTJSONCounters parses `nft -j list counters` output.
// Format: {"nftables": [{"counter": {"name":"cnt_xxx","table":"nixguard","packets":N,"bytes":N}}, ...]}
func parseNFTJSONCounters(jsonStr string) map[string]firewall.RuleStats {
	stats := make(map[string]firewall.RuleStats)

	var result struct {
		Nftables []json.RawMessage `json:"nftables"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return stats
	}

	for _, raw := range result.Nftables {
		var obj struct {
			Counter *struct {
				Name    string `json:"name"`
				Table   string `json:"table"`
				Packets uint64 `json:"packets"`
				Bytes   uint64 `json:"bytes"`
			} `json:"counter"`
		}
		if err := json.Unmarshal(raw, &obj); err != nil || obj.Counter == nil {
			continue
		}
		if obj.Counter.Table != "" && obj.Counter.Table != "nixguard" {
			continue
		}

		// Counter names are "cnt_<ruleID>"
		name := obj.Counter.Name
		if len(name) > 4 && name[:4] == "cnt_" {
			ruleID := name[4:]
			stats[ruleID] = firewall.RuleStats{
				Packets: obj.Counter.Packets,
				Bytes:   obj.Counter.Bytes,
			}
		}
	}

	return stats
}

func isPrivilegeError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "you must be root")
}

// Ensure interface compliance at compile time
var _ firewall.FirewallEngine = (*Adapter)(nil)
