// Package dns contains the domain model for DNS services.
// Maps to OPNsense: Services > Unbound DNS, DNS Filtering.
package dns

import "time"

// ResolverConfig is the main Unbound DNS resolver configuration.
type ResolverConfig struct {
	ID               string   `json:"id" db:"id"`
	Enabled          bool     `json:"enabled" db:"enabled"`
	ListenInterfaces []string `json:"listen_interfaces"`
	Port             int      `json:"port" db:"port"`
	DNSSEC           bool     `json:"dnssec" db:"dnssec"`
	DNSOverTLS       bool     `json:"dns_over_tls" db:"dns_over_tls"`
	ForwardMode      bool     `json:"forward_mode" db:"forward_mode"`
	Forwarders       []string `json:"forwarders"`
	QueryMinimize    bool     `json:"query_minimize" db:"query_minimize"`
	AggressiveNSEC   bool     `json:"aggressive_nsec" db:"aggressive_nsec"`
	Prefetch         bool     `json:"prefetch" db:"prefetch"`
	ServeExpired     bool     `json:"serve_expired" db:"serve_expired"`
	CacheSize        int      `json:"cache_size" db:"cache_size"` // MB
	RebindProtection bool     `json:"rebind_protection" db:"rebind_protection"`
	PrivateNetworks  []string `json:"private_networks"`
	QueryLogging     bool     `json:"query_logging" db:"query_logging"`
}

// HostOverride is a manual DNS A/AAAA record.
type HostOverride struct {
	ID          string    `json:"id" db:"id"`
	Hostname    string    `json:"hostname" db:"hostname"`
	Domain      string    `json:"domain" db:"domain"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	Type        string    `json:"type" db:"record_type"` // A, AAAA
	MXPriority  int       `json:"mx_priority,omitempty" db:"mx_priority"`
	Description string    `json:"description" db:"description"`
	Aliases     []HostAlias `json:"aliases,omitempty"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type HostAlias struct {
	Hostname string `json:"hostname"`
	Domain   string `json:"domain"`
}

// DomainOverride forwards queries for a domain to specific DNS servers.
type DomainOverride struct {
	ID          string    `json:"id" db:"id"`
	Domain      string    `json:"domain" db:"domain"`
	IPAddress   string    `json:"ip_address" db:"ip_address"` // DNS server
	Port        int       `json:"port" db:"port"`
	UseTLS      bool      `json:"use_tls" db:"use_tls"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ─── DNS Filtering / Ad-Blocking ───────────────────────────────

type Blocklist struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	URL         string    `json:"url" db:"url"`
	Type        string    `json:"type" db:"list_type"` // domains, hosts
	Enabled     bool      `json:"enabled" db:"enabled"`
	EntryCount  int       `json:"entry_count" db:"entry_count"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
	Description string    `json:"description" db:"description"`
}

type Whitelist struct {
	ID          string    `json:"id" db:"id"`
	Domain      string    `json:"domain" db:"domain"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// DNSQueryLog represents a single DNS query for logging.
type DNSQueryLog struct {
	Timestamp   time.Time `json:"timestamp"`
	ClientIP    string    `json:"client_ip"`
	QueryName   string    `json:"query_name"`
	QueryType   string    `json:"query_type"` // A, AAAA, MX, etc.
	ResponseCode string   `json:"response_code"` // NOERROR, NXDOMAIN, etc.
	Blocked     bool      `json:"blocked"`
	BlockedBy   string    `json:"blocked_by,omitempty"` // blocklist name
	ResponseTime float64  `json:"response_time_ms"`
}

// AccessControl defines who can query the DNS resolver.
type AccessControl struct {
	Network string `json:"network"` // CIDR
	Action  string `json:"action"`  // allow, deny, refuse, allow_snoop
}
