// Package network contains the domain model for network interfaces, routing, and gateways.
// Maps to OPNsense: Interfaces, System > Routing, System > Gateways.
package network

import "time"

// ─── Interface ─────────────────────────────────────────────────

// Interface represents a network interface configuration.
type Interface struct {
	ID          string         `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`           // e.g., "eth0"
	Alias       string         `json:"alias" db:"alias_name"`    // e.g., "WAN", "LAN"
	Type        InterfaceType  `json:"type" db:"if_type"`
	Enabled     bool           `json:"enabled" db:"enabled"`
	Description string         `json:"description" db:"description"`
	MTU         int            `json:"mtu" db:"mtu"`
	MAC         string         `json:"mac" db:"mac_address"`
	IPv4Config  *IPv4Config    `json:"ipv4_config,omitempty"`
	IPv6Config  *IPv6Config    `json:"ipv6_config,omitempty"`
	VLANConfig  *VLANConfig    `json:"vlan_config,omitempty"`
	BondConfig  *BondConfig    `json:"bond_config,omitempty"`
	BridgeConfig *BridgeConfig `json:"bridge_config,omitempty"`
	PPPoEConfig *PPPoEConfig   `json:"pppoe_config,omitempty"`
	Status      InterfaceStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

type InterfaceType string

const (
	IfTypePhysical InterfaceType = "physical"
	IfTypeVLAN     InterfaceType = "vlan"
	IfTypeBond     InterfaceType = "bond"
	IfTypeBridge   InterfaceType = "bridge"
	IfTypePPPoE    InterfaceType = "pppoe"
	IfTypeVXLAN    InterfaceType = "vxlan"
	IfTypeGRE      InterfaceType = "gre"
	IfTypeGIF      InterfaceType = "gif"
	IfTypeWireless InterfaceType = "wireless"
	IfTypeLoopback InterfaceType = "loopback"
)

type IPv4Config struct {
	Mode    string `json:"mode"`     // static, dhcp
	Address string `json:"address"`  // CIDR notation
	Gateway string `json:"gateway"`
}

type IPv6Config struct {
	Mode    string `json:"mode"`     // static, dhcpv6, slaac, track
	Address string `json:"address"`
	Gateway string `json:"gateway"`
	PrefixID int   `json:"prefix_id"`
}

type VLANConfig struct {
	Parent string `json:"parent"`
	Tag    int    `json:"tag"`
	QinQ   bool   `json:"qinq"`
}

type BondConfig struct {
	Members []string `json:"members"`
	Mode    string   `json:"mode"` // balance-rr, active-backup, 802.3ad, etc.
	Primary string   `json:"primary"`
}

type BridgeConfig struct {
	Members []string `json:"members"`
	STP     bool     `json:"stp"`
}

type PPPoEConfig struct {
	Parent   string `json:"parent"`
	Username string `json:"username"`
	Password string `json:"password"`
	ServiceName string `json:"service_name"`
}

// InterfaceStatus is the runtime status of an interface.
type InterfaceStatus struct {
	OperState  string   `json:"oper_state"` // up, down, unknown
	Speed      string   `json:"speed"`      // 1000Mb/s
	Duplex     string   `json:"duplex"`     // full, half
	Carrier    bool     `json:"carrier"`
	Addresses  []string `json:"addresses"`
	RxBytes    uint64   `json:"rx_bytes"`
	TxBytes    uint64   `json:"tx_bytes"`
	RxPackets  uint64   `json:"rx_packets"`
	TxPackets  uint64   `json:"tx_packets"`
	RxErrors   uint64   `json:"rx_errors"`
	TxErrors   uint64   `json:"tx_errors"`
	RxDropped  uint64   `json:"rx_dropped"`
	TxDropped  uint64   `json:"tx_dropped"`
}

// ─── Route ─────────────────────────────────────────────────────

type Route struct {
	ID          string    `json:"id" db:"id"`
	Destination string    `json:"destination" db:"destination"` // CIDR
	Gateway     string    `json:"gateway" db:"gateway"`
	Interface   string    `json:"interface" db:"interface_name"`
	Metric      int       `json:"metric" db:"metric"`
	Table       int       `json:"table" db:"route_table"`
	Type        RouteType `json:"type" db:"route_type"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type RouteType string

const (
	RouteStatic  RouteType = "static"
	RouteDynamic RouteType = "dynamic"
	RoutePolicy  RouteType = "policy"
)

// ─── Gateway ───────────────────────────────────────────────────

type Gateway struct {
	ID              string        `json:"id" db:"id"`
	Name            string        `json:"name" db:"name"`
	Interface       string        `json:"interface" db:"interface_name"`
	Address         string        `json:"address" db:"address"`
	Protocol        string        `json:"protocol" db:"protocol"` // inet, inet6
	MonitorIP       string        `json:"monitor_ip" db:"monitor_ip"`
	Weight          int           `json:"weight" db:"weight"`
	Priority        int           `json:"priority" db:"priority"`
	IsDefault       bool          `json:"is_default" db:"is_default"`
	Enabled         bool          `json:"enabled" db:"enabled"`
	Description     string        `json:"description" db:"description"`
	MonitorConfig   MonitorConfig `json:"monitor_config"`
	Status          GatewayStatus `json:"status"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

type MonitorConfig struct {
	Interval     int    `json:"interval"`       // seconds between probes
	LossThreshold int   `json:"loss_threshold"` // percent
	LatencyThreshold int `json:"latency_threshold"` // milliseconds
	DownCount    int    `json:"down_count"`     // probes before marking down
	Method       string `json:"method"`         // icmp, tcp, http
	Port         int    `json:"port"`           // for TCP checks
	URL          string `json:"url"`            // for HTTP checks
}

type GatewayStatus struct {
	State      string  `json:"state"` // online, offline, unknown
	Latency    float64 `json:"latency_ms"`
	PacketLoss float64 `json:"packet_loss_percent"`
	LastCheck  time.Time `json:"last_check"`
}

// GatewayGroup allows load balancing and failover across gateways.
type GatewayGroup struct {
	ID        string               `json:"id" db:"id"`
	Name      string               `json:"name" db:"name"`
	Members   []GatewayGroupMember `json:"members"`
	Trigger   string               `json:"trigger" db:"trigger"` // member_down, packet_loss, high_latency
	Description string             `json:"description" db:"description"`
}

type GatewayGroupMember struct {
	GatewayID string `json:"gateway_id"`
	Tier      int    `json:"tier"`   // 1 = primary, 2 = secondary, etc.
	Weight    int    `json:"weight"` // for load balancing within a tier
}
