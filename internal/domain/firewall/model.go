// Package firewall contains the domain model for NixGuard's firewall engine.
// Maps to OPNsense: Firewall > Rules, NAT, Aliases, Schedules.
package firewall

import (
	"net"
	"time"
)

// ─── Rule ──────────────────────────────────────────────────────

// Rule represents a single firewall rule.
type Rule struct {
	ID          string    `json:"id" db:"id"`
	Interface   string    `json:"interface" db:"interface_name"`
	Direction   Direction `json:"direction" db:"direction"`
	Action      Action    `json:"action" db:"action"`
	Protocol    Protocol  `json:"protocol" db:"protocol"`
	Source      Address   `json:"source" db:"source"`
	Destination Address   `json:"destination" db:"destination"`
	Log         bool      `json:"log" db:"log_enabled"`
	Description string    `json:"description" db:"description"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	Order       int       `json:"order" db:"rule_order"`
	Category    string    `json:"category" db:"category"`
	Schedule    *Schedule `json:"schedule,omitempty"`
	Stats       RuleStats `json:"stats"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Floating rules can apply to multiple interfaces.
	IsFloating bool     `json:"is_floating" db:"is_floating"`
	Interfaces []string `json:"interfaces,omitempty"`

	// Advanced
	Gateway   string `json:"gateway,omitempty" db:"gateway"`
	StateType string `json:"state_type,omitempty" db:"state_type"` // keep, sloppy, synproxy
	MaxStates int    `json:"max_states,omitempty" db:"max_states"`
	Tag       string `json:"tag,omitempty" db:"tag"`
	Tagged    string `json:"tagged,omitempty" db:"tagged"`
}

type Direction string

const (
	DirectionIn  Direction = "in"
	DirectionOut Direction = "out"
)

type Action string

const (
	ActionPass   Action = "pass"
	ActionBlock  Action = "block"
	ActionReject Action = "reject"
)

type Protocol string

const (
	ProtoAny    Protocol = "any"
	ProtoTCP    Protocol = "tcp"
	ProtoUDP    Protocol = "udp"
	ProtoICMP   Protocol = "icmp"
	ProtoICMPv6 Protocol = "icmpv6"
	ProtoESP    Protocol = "esp"
	ProtoAH     Protocol = "ah"
	ProtoGRE    Protocol = "gre"
)

// Address represents a source or destination in a rule.
type Address struct {
	Type  AddressType `json:"type"`  // any, single, network, alias, geoip
	Value string      `json:"value"` // IP, CIDR, alias name, country code
	Port  string      `json:"port"`  // single port, range "80-443", or alias
	Not   bool        `json:"not"`   // invert match
}

type AddressType string

const (
	AddrAny     AddressType = "any"
	AddrSingle  AddressType = "single"
	AddrAddress AddressType = "address" // alias for AddrSingle, used by API
	AddrNetwork AddressType = "network"
	AddrAlias   AddressType = "alias"
	AddrGeoIP   AddressType = "geoip"
)

// RuleStats holds rule hit counters.
type RuleStats struct {
	Evaluations uint64 `json:"evaluations"`
	Packets     uint64 `json:"packets"`
	Bytes       uint64 `json:"bytes"`
}

// Schedule defines when a rule is active.
type Schedule struct {
	Name      string    `json:"name"`
	StartTime string    `json:"start_time"` // HH:MM
	EndTime   string    `json:"end_time"`   // HH:MM
	Weekdays  []int     `json:"weekdays"`   // 0=Sun, 6=Sat
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
}

// ─── Alias ─────────────────────────────────────────────────────

// Alias is a named group of addresses, ports, or networks.
type Alias struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Type        AliasType `json:"type" db:"alias_type"`
	Description string    `json:"description" db:"description"`
	Entries     []string  `json:"entries"`     // IPs, CIDRs, ports, URLs
	UpdateFreq  string    `json:"update_freq"` // for URL tables: "1h", "24h"
	Enabled     bool      `json:"enabled" db:"enabled"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Resolved contains the effective entries after resolving nested aliases.
	Resolved []net.IP `json:"-"`
}

type AliasType string

const (
	AliasHost     AliasType = "host"      // IP addresses
	AliasNetwork  AliasType = "network"   // CIDRs
	AliasPort     AliasType = "port"      // port numbers/ranges
	AliasURL      AliasType = "url"       // fetched from URL
	AliasURLTable AliasType = "url_table" // large lists from URL
	AliasGeoIP    AliasType = "geoip"     // country codes
	AliasNested   AliasType = "nested"    // references other aliases
)

// ─── NAT ───────────────────────────────────────────────────────

// NATRule represents a NAT rule (port forward, 1:1, outbound).
type NATRule struct {
	ID             string    `json:"id" db:"id"`
	Type           NATType   `json:"type" db:"nat_type"`
	Interface      string    `json:"interface" db:"interface_name"`
	Protocol       Protocol  `json:"protocol" db:"protocol"`
	Source         Address   `json:"source"`
	Destination    Address   `json:"destination"`
	RedirectTarget string    `json:"redirect_target" db:"redirect_target"`
	RedirectPort   string    `json:"redirect_port" db:"redirect_port"`
	Description    string    `json:"description" db:"description"`
	Enabled        bool      `json:"enabled" db:"enabled"`
	NATReflection       bool      `json:"nat_reflection" db:"nat_reflection"`
	ReflectionInterface string    `json:"reflection_interface" db:"reflection_interface"` // LAN interface for reflection
	ReflectionNetwork   string    `json:"reflection_network" db:"reflection_network"`     // LAN subnet e.g. "192.168.1.0/24"
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

type NATType string

const (
	NATPortForward NATType = "port_forward"
	NATOneToOne    NATType = "one_to_one"
	NATOutbound    NATType = "outbound"
)

// ─── State Table ───────────────────────────────────────────────

// State represents a connection tracking entry.
type State struct {
	Protocol   string `json:"protocol"`
	SourceIP   string `json:"source_ip"`
	SourcePort int    `json:"source_port"`
	DestIP     string `json:"dest_ip"`
	DestPort   int    `json:"dest_port"`
	State      string `json:"state"`
	Direction  string `json:"direction"`
	Interface  string `json:"interface"`
	RuleID     string `json:"rule_id"`
	Packets    uint64 `json:"packets"`
	Bytes      uint64 `json:"bytes"`
	Age        string `json:"age"`
	Expires    string `json:"expires"`
}

// TrafficFilter controls packet capture used by the live traffic view.
type TrafficFilter struct {
	Interface string `json:"interface"`
	SourceIP  string `json:"source_ip"`
	DestIP    string `json:"dest_ip"`
	Protocol  string `json:"protocol"`
	Count     int    `json:"count"`
	SnapLen   int    `json:"snap_len"`
}

// CapturedPacket is a summarized tcpdump frame shown in the GUI.
type CapturedPacket struct {
	Timestamp   string `json:"timestamp"`
	Interface   string `json:"interface"`
	Protocol    string `json:"protocol"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Length      int    `json:"length"`
	Verdict     string `json:"verdict"`
	Summary     string `json:"summary"`
	Detail      string `json:"detail"`
}

// PCAPExport describes an exported pcap capture file.
type PCAPExport struct {
	Name        string    `json:"name"`
	DownloadURL string    `json:"download_url"`
	Bytes       int64     `json:"bytes"`
	CreatedAt   time.Time `json:"created_at"`
}
