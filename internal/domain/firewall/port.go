// Port interfaces define what the firewall module needs from infrastructure.
// These are implemented by the nftables/iptables adapter.
package firewall

import "context"

// RuleRepository persists firewall rules.
type RuleRepository interface {
	List(ctx context.Context, filter RuleFilter) ([]Rule, error)
	GetByID(ctx context.Context, id string) (*Rule, error)
	Create(ctx context.Context, rule *Rule) error
	Update(ctx context.Context, rule *Rule) error
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, ruleIDs []string) error
	GetByInterface(ctx context.Context, iface string) ([]Rule, error)
}

// RuleFilter defines query filters for listing rules.
type RuleFilter struct {
	Interface  string
	Action     Action
	Protocol   Protocol
	Enabled    *bool
	IsFloating *bool
	Category   string
	Search     string
	Offset     int
	Limit      int
}

// AliasRepository persists aliases.
type AliasRepository interface {
	List(ctx context.Context) ([]Alias, error)
	GetByID(ctx context.Context, id string) (*Alias, error)
	GetByName(ctx context.Context, name string) (*Alias, error)
	Create(ctx context.Context, alias *Alias) error
	Update(ctx context.Context, alias *Alias) error
	Delete(ctx context.Context, id string) error
}

// NATRepository persists NAT rules.
type NATRepository interface {
	List(ctx context.Context, natType NATType) ([]NATRule, error)
	GetByID(ctx context.Context, id string) (*NATRule, error)
	Create(ctx context.Context, rule *NATRule) error
	Update(ctx context.Context, rule *NATRule) error
	Delete(ctx context.Context, id string) error
}

// FirewallEngine is the interface to the underlying firewall system (nftables/iptables).
type FirewallEngine interface {
	// ApplyRules compiles and applies all firewall rules atomically.
	ApplyRules(ctx context.Context, rules []Rule, aliases []Alias) error

	// ApplyNAT compiles and applies all NAT rules atomically.
	ApplyNAT(ctx context.Context, rules []NATRule) error

	// GetStates returns current connection tracking entries.
	GetStates(ctx context.Context, filter StateFilter) ([]State, error)

	// FlushStates removes matching connection tracking entries.
	FlushStates(ctx context.Context, filter StateFilter) error

	// GetRuleStats returns per-rule packet/byte counters.
	GetRuleStats(ctx context.Context) (map[string]RuleStats, error)

	// CaptureTraffic returns a bounded live snapshot of packets for the GUI.
	CaptureTraffic(ctx context.Context, filter TrafficFilter) ([]CapturedPacket, error)

	// ExportPCAP captures packets to a pcap file for download.
	ExportPCAP(ctx context.Context, filter TrafficFilter) (*PCAPExport, error)
}

// StateFilter filters connection tracking entries.
type StateFilter struct {
	Protocol  string
	SourceIP  string
	DestIP    string
	Interface string
	Limit     int
}

// GeoIPProvider resolves country codes to IP ranges.
type GeoIPProvider interface {
	Resolve(ctx context.Context, countryCode string) ([]string, error)
	Update(ctx context.Context) error
	LastUpdated(ctx context.Context) (string, error)
}
