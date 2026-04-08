// Package ids contains the domain model for Intrusion Detection/Prevention.
// Maps to OPNsense: Services > Intrusion Detection (Suricata).
package ids

import "time"

// Config is the Suricata IDS/IPS configuration.
type Config struct {
	ID               string   `json:"id" db:"id"`
	Enabled          bool     `json:"enabled" db:"enabled"`
	Mode             Mode     `json:"mode" db:"mode"`
	Interfaces       []string `json:"interfaces"`
	Pattern          string   `json:"pattern" db:"pattern"`           // ac, ac-bs, hs (hyperscan)
	DefaultAction    string   `json:"default_action" db:"default_action"` // alert, drop
	PromiscuousMode  bool     `json:"promiscuous_mode" db:"promiscuous_mode"`
	EVELogEnabled    bool     `json:"eve_log_enabled" db:"eve_log_enabled"`
	SyslogEnabled    bool     `json:"syslog_enabled" db:"syslog_enabled"`
	HomepNetworks    []string `json:"home_networks"`
}

type Mode string

const (
	ModeIDS Mode = "ids"
	ModeIPS Mode = "ips"
)

// Ruleset is a collection of IDS rules from a source.
type Ruleset struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Source      string    `json:"source" db:"source"`   // et/open, et/pro, snort
	URL         string    `json:"url" db:"url"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	RuleCount   int       `json:"rule_count" db:"rule_count"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}

// RuleCategory is a group of related rules.
type RuleCategory struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	RuleCount   int    `json:"rule_count"`
}

// RuleOverride customizes the action for a specific rule SID.
type RuleOverride struct {
	SID    int    `json:"sid" db:"sid"`
	Action string `json:"action" db:"action"` // alert, drop, disable
}

// Alert represents a detected intrusion event.
type Alert struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    int       `json:"severity"`   // 1 (high) to 4 (low)
	Category    string    `json:"category"`
	Signature   string    `json:"signature"`
	SID         int       `json:"sid"`
	GID         int       `json:"gid"`
	Rev         int       `json:"rev"`
	Protocol    string    `json:"protocol"`
	SourceIP    string    `json:"source_ip"`
	SourcePort  int       `json:"source_port"`
	DestIP      string    `json:"dest_ip"`
	DestPort    int       `json:"dest_port"`
	Interface   string    `json:"interface"`
	Action      string    `json:"action"` // allowed, blocked
	Payload     string    `json:"payload,omitempty"`
}

// AlertFilter for querying alerts.
type AlertFilter struct {
	Severity  int
	Category  string
	SourceIP  string
	DestIP    string
	Interface string
	SID       int
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Offset    int
}
