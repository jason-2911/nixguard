// Package dhcp contains the domain model for DHCP services.
// Maps to OPNsense: Services > DHCPv4, DHCPv6, Router Advertisements.
package dhcp

import "time"

// ServerConfig is the DHCP server configuration for an interface.
type ServerConfig struct {
	ID           string    `json:"id" db:"id"`
	Interface    string    `json:"interface" db:"interface_name"`
	Enabled      bool      `json:"enabled" db:"enabled"`
	RangeStart   string    `json:"range_start" db:"range_start"`
	RangeEnd     string    `json:"range_end" db:"range_end"`
	SubnetMask   string    `json:"subnet_mask" db:"subnet_mask"`
	Gateway      string    `json:"gateway" db:"gateway"`
	DNSServers   []string  `json:"dns_servers"`
	DomainName   string    `json:"domain_name" db:"domain_name"`
	DefaultLease int       `json:"default_lease" db:"default_lease"` // seconds
	MaxLease     int       `json:"max_lease" db:"max_lease"`         // seconds
	NTPServers   []string  `json:"ntp_servers"`
	TFTPServer   string    `json:"tftp_server" db:"tftp_server"`
	BootFile     string    `json:"boot_file" db:"boot_file"`
	WINSServers  []string  `json:"wins_servers"`
	Description  string    `json:"description" db:"description"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// StaticMapping maps a MAC address to a fixed IP.
type StaticMapping struct {
	ID          string    `json:"id" db:"id"`
	Interface   string    `json:"interface" db:"interface_name"`
	MACAddress  string    `json:"mac_address" db:"mac_address"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	Hostname    string    `json:"hostname" db:"hostname"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Lease represents an active DHCP lease.
type Lease struct {
	IPAddress  string    `json:"ip_address"`
	MACAddress string    `json:"mac_address"`
	Hostname   string    `json:"hostname"`
	Interface  string    `json:"interface"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	State      string    `json:"state"` // active, expired, released
	IsStatic   bool      `json:"is_static"`
}

// DHCPv6Config is the DHCPv6 server configuration.
type DHCPv6Config struct {
	ID           string    `json:"id" db:"id"`
	Interface    string    `json:"interface" db:"interface_name"`
	Enabled      bool      `json:"enabled" db:"enabled"`
	Mode         string    `json:"mode" db:"mode"` // stateful, stateless
	RangeStart   string    `json:"range_start" db:"range_start"`
	RangeEnd     string    `json:"range_end" db:"range_end"`
	PrefixLength int       `json:"prefix_length" db:"prefix_length"`
	DNSServers   []string  `json:"dns_servers"`
	DomainName   string    `json:"domain_name" db:"domain_name"`
}

// RouterAdvertisement configures IPv6 RA for an interface.
type RouterAdvertisement struct {
	ID          string `json:"id" db:"id"`
	Interface   string `json:"interface" db:"interface_name"`
	Enabled     bool   `json:"enabled" db:"enabled"`
	Mode        string `json:"mode" db:"mode"` // assisted, unmanaged, managed, disabled
	Prefix      string `json:"prefix" db:"prefix"`
	DNSServers  []string `json:"dns_servers"`
	DomainSearch []string `json:"domain_search"`
	Priority    string `json:"priority" db:"priority"` // high, medium, low
}

// RelayConfig is DHCP relay configuration.
type RelayConfig struct {
	ID            string   `json:"id" db:"id"`
	Enabled       bool     `json:"enabled" db:"enabled"`
	Interface     string   `json:"interface" db:"interface_name"`
	Destinations  []string `json:"destinations"` // DHCP server IPs
	AgentInfo     bool     `json:"agent_info" db:"agent_info"` // RFC 3046
}
