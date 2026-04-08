// Package ha contains the domain model for High Availability.
// Maps to OPNsense: System > High Availability, Virtual IPs (CARP).
package ha

import "time"

// VirtualIP represents a CARP virtual IP address.
type VirtualIP struct {
	ID          string `json:"id" db:"id"`
	Interface   string `json:"interface" db:"interface_name"`
	Type        string `json:"type" db:"vip_type"` // carp, ipalias, proxyarp
	Address     string `json:"address" db:"address"`
	SubnetMask  int    `json:"subnet_mask" db:"subnet_mask"`
	VHID        int    `json:"vhid" db:"vhid"`
	AdvBase     int    `json:"adv_base" db:"adv_base"`
	AdvSkew     int    `json:"adv_skew" db:"adv_skew"`
	Password    string `json:"-" db:"password"`
	Description string `json:"description" db:"description"`
	Status      string `json:"status"` // master, backup, init
}

// HAConfig is the overall HA configuration.
type HAConfig struct {
	Enabled           bool   `json:"enabled" db:"enabled"`
	SyncInterface     string `json:"sync_interface" db:"sync_interface"`
	PeerIP            string `json:"peer_ip" db:"peer_ip"`
	PeerUser          string `json:"peer_user" db:"peer_user"`
	PeerPassword      string `json:"-" db:"peer_password"`
	SyncFirewall      bool   `json:"sync_firewall" db:"sync_firewall"`
	SyncNAT           bool   `json:"sync_nat" db:"sync_nat"`
	SyncDHCP          bool   `json:"sync_dhcp" db:"sync_dhcp"`
	SyncVPN           bool   `json:"sync_vpn" db:"sync_vpn"`
	SyncUsers         bool   `json:"sync_users" db:"sync_users"`
	SyncCerts         bool   `json:"sync_certs" db:"sync_certs"`
	SyncDNS           bool   `json:"sync_dns" db:"sync_dns"`
	SyncAliases       bool   `json:"sync_aliases" db:"sync_aliases"`
	StateSync         bool   `json:"state_sync" db:"state_sync"` // pfsync equivalent
	StateSyncInterface string `json:"state_sync_interface" db:"state_sync_interface"`
}

// HAStatus shows current HA cluster state.
type HAStatus struct {
	Role          string       `json:"role"` // master, backup
	PeerState     string       `json:"peer_state"` // online, offline
	LastSync      time.Time    `json:"last_sync"`
	VirtualIPs    []VirtualIP  `json:"virtual_ips"`
	StateSyncOK   bool         `json:"state_sync_ok"`
}
