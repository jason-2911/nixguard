// Package suricata implements the ids.IDSEngine interface.
// Manages Suricata IDS/IPS configuration, rules, and alert retrieval.
package suricata

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/nixguard/nixguard/internal/domain/ids"
	"github.com/nixguard/nixguard/pkg/executor"
)

const (
	configDir   = "/etc/suricata"
	configFile  = "/etc/suricata/suricata.yaml"
	ruleDir     = "/etc/suricata/rules"
	logDir      = "/var/log/suricata"
	eveLogFile  = "/var/log/suricata/eve.json"
)

// Adapter implements ids.IDSEngine.
type Adapter struct {
	exec *executor.Safe
	log  *slog.Logger
}

// NewAdapter creates a Suricata adapter.
func NewAdapter(exec *executor.Safe, log *slog.Logger) *Adapter {
	return &Adapter{exec: exec, log: log}
}

// ApplyConfig generates and writes the Suricata configuration.
func (a *Adapter) ApplyConfig(ctx context.Context, cfg ids.Config) error {
	yamlContent := a.generateConfig(cfg)

	// Write config via agent
	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("write suricata config: %w", err)
	}

	return a.Reload(ctx)
}

// UpdateRules downloads and updates IDS rulesets.
func (a *Adapter) UpdateRules(ctx context.Context) error {
	_, err := a.exec.Run(ctx, "suricata-update", "update")
	if err != nil {
		return fmt.Errorf("suricata-update: %w", err)
	}
	return a.Reload(ctx)
}

// GetAlerts reads alerts from the EVE JSON log.
func (a *Adapter) GetAlerts(ctx context.Context, filter ids.AlertFilter) ([]ids.Alert, error) {
	data, err := os.ReadFile(eveLogFile)
	if err != nil {
		return nil, fmt.Errorf("read eve.json: %w", err)
	}

	var alerts []ids.Alert
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}

		var eveEvent map[string]interface{}
		if err := json.Unmarshal([]byte(line), &eveEvent); err != nil {
			continue
		}

		eventType, _ := eveEvent["event_type"].(string)
		if eventType != "alert" {
			continue
		}

		alert := parseEVEAlert(eveEvent)
		if matchesFilter(alert, filter) {
			alerts = append(alerts, alert)
		}

		if filter.Limit > 0 && len(alerts) >= filter.Limit {
			break
		}
	}

	return alerts, nil
}

// GetStats returns Suricata runtime statistics.
func (a *Adapter) GetStats(ctx context.Context) (*ids.Stats, error) {
	// Read stats from Suricata unix socket or stats log
	statsFile := filepath.Join(logDir, "stats.log")
	data, err := os.ReadFile(statsFile)
	if err != nil {
		return &ids.Stats{}, nil
	}

	return parseStats(string(data)), nil
}

// Restart stops and starts Suricata.
func (a *Adapter) Restart(ctx context.Context) error {
	_, err := a.exec.Run(ctx, "systemctl", "restart", "suricata")
	return err
}

// Reload sends SIGUSR2 to reload rules without dropping packets.
func (a *Adapter) Reload(ctx context.Context) error {
	_, err := a.exec.Run(ctx, "systemctl", "kill", "-s", "USR2", "suricata")
	return err
}

// ─── Config Generation ─────────────────────────────────────────

func (a *Adapter) generateConfig(cfg ids.Config) string {
	// Generate suricata.yaml from domain config
	// This is simplified; real implementation uses a full template
	mode := "af-packet"
	if cfg.Mode == ids.ModeIPS {
		mode = "nfqueue"
	}

	return fmt.Sprintf(`%%YAML 1.1
---
# NixGuard - Suricata Configuration (auto-generated)

vars:
  address-groups:
    HOME_NET: "[%s]"
    EXTERNAL_NET: "!$HOME_NET"
  port-groups:
    HTTP_PORTS: "80"
    SHELLCODE_PORTS: "!80"
    SSH_PORTS: "22"

default-log-dir: %s

outputs:
  - eve-log:
      enabled: %v
      filetype: regular
      filename: eve.json
      types:
        - alert
        - http
        - dns
        - tls
        - files
        - smtp
        - flow
  - fast:
      enabled: yes
      filename: fast.log

%s:
  - interface: %s
    cluster-id: 99
    cluster-type: cluster_flow
    defrag: yes

detect:
  profile: medium
  sgh-mpm-context: auto
  inspection-recursion-limit: 3000

threading:
  set-cpu-affinity: no
  detect-thread-ratio: 1.0

default-rule-path: %s
rule-files:
  - suricata.rules
`,
		strings.Join(cfg.HomepNetworks, ", "),
		logDir,
		cfg.EVELogEnabled,
		mode,
		strings.Join(cfg.Interfaces, "\n  - interface: "),
		ruleDir,
	)
}

// ─── Parsers ───────────────────────────────────────────────────

func parseEVEAlert(evt map[string]interface{}) ids.Alert {
	alert := ids.Alert{}
	// Parse EVE JSON alert fields
	if ts, ok := evt["timestamp"].(string); ok {
		_ = ts // parse timestamp
	}
	if a, ok := evt["alert"].(map[string]interface{}); ok {
		if sig, ok := a["signature"].(string); ok {
			alert.Signature = sig
		}
		if cat, ok := a["category"].(string); ok {
			alert.Category = cat
		}
		if sev, ok := a["severity"].(float64); ok {
			alert.Severity = int(sev)
		}
		if sid, ok := a["signature_id"].(float64); ok {
			alert.SID = int(sid)
		}
	}
	if src, ok := evt["src_ip"].(string); ok {
		alert.SourceIP = src
	}
	if dst, ok := evt["dest_ip"].(string); ok {
		alert.DestIP = dst
	}
	if proto, ok := evt["proto"].(string); ok {
		alert.Protocol = proto
	}
	return alert
}

func matchesFilter(alert ids.Alert, filter ids.AlertFilter) bool {
	if filter.Severity > 0 && alert.Severity != filter.Severity {
		return false
	}
	if filter.SourceIP != "" && alert.SourceIP != filter.SourceIP {
		return false
	}
	if filter.DestIP != "" && alert.DestIP != filter.DestIP {
		return false
	}
	if filter.SID > 0 && alert.SID != filter.SID {
		return false
	}
	return true
}

func parseStats(data string) *ids.Stats {
	stats := &ids.Stats{}
	// Parse Suricata stats.log format
	_ = data
	return stats
}
