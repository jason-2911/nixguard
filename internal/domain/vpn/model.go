// Package vpn contains the domain model for VPN services.
// Supports IPsec (StrongSwan), OpenVPN, and WireGuard.
// Maps to OPNsense: VPN > IPsec, VPN > OpenVPN, VPN > WireGuard.
package vpn

import "time"

// ─── IPsec ─────────────────────────────────────────────────────

type IPsecTunnel struct {
	ID           string        `json:"id" db:"id"`
	Name         string        `json:"name" db:"name"`
	Type         IPsecType     `json:"type" db:"tunnel_type"` // site-to-site, roadwarrior
	Enabled      bool          `json:"enabled" db:"enabled"`
	RemoteGateway string       `json:"remote_gateway" db:"remote_gateway"`
	LocalID      string        `json:"local_id" db:"local_id"`
	RemoteID     string        `json:"remote_id" db:"remote_id"`
	Phase1       Phase1Config  `json:"phase1"`
	Phase2       []Phase2Config `json:"phase2"`
	AuthMethod   string        `json:"auth_method" db:"auth_method"` // psk, cert, eap
	PSK          string        `json:"psk,omitempty" db:"psk"`
	CertID       string        `json:"cert_id,omitempty" db:"cert_id"`
	DPD          DPDConfig     `json:"dpd"`
	Status       TunnelStatus  `json:"status"`
	Description  string        `json:"description" db:"description"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
}

type IPsecType string

const (
	IPsecSiteToSite  IPsecType = "site_to_site"
	IPsecRoadWarrior IPsecType = "road_warrior"
)

type Phase1Config struct {
	Version     int      `json:"version"`     // 1 or 2 (IKEv1/IKEv2)
	Encryption  []string `json:"encryption"`  // aes256, aes128, 3des
	Hash        []string `json:"hash"`        // sha256, sha512, sha1
	DHGroup     []int    `json:"dh_group"`    // 14, 19, 20, 21
	Lifetime    int      `json:"lifetime"`    // seconds
	NATTraversal bool    `json:"nat_traversal"`
}

type Phase2Config struct {
	LocalNetwork  string   `json:"local_network"`  // CIDR
	RemoteNetwork string   `json:"remote_network"` // CIDR
	Protocol      string   `json:"protocol"`       // esp, ah
	Encryption    []string `json:"encryption"`
	Hash          []string `json:"hash"`
	PFS           int      `json:"pfs"`            // DH group, 0 = disabled
	Lifetime      int      `json:"lifetime"`       // seconds
}

type DPDConfig struct {
	Enabled  bool   `json:"enabled"`
	Delay    int    `json:"delay"`    // seconds
	Timeout  int    `json:"timeout"`  // seconds
	Action   string `json:"action"`   // restart, clear, hold
}

// ─── OpenVPN ───────────────────────────────────────────────────

type OpenVPNServer struct {
	ID          string          `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Mode        OpenVPNMode     `json:"mode" db:"mode"`
	Protocol    string          `json:"protocol" db:"protocol"` // udp, tcp
	Port        int             `json:"port" db:"port"`
	DeviceMode  string          `json:"device_mode" db:"device_mode"` // tun, tap
	TunnelNet   string          `json:"tunnel_network" db:"tunnel_network"`
	LocalNet    []string        `json:"local_networks"`
	Cipher      string          `json:"cipher" db:"cipher"`
	Auth        string          `json:"auth" db:"auth_digest"`
	TLSVersion  string          `json:"tls_version" db:"tls_version"`
	CAID        string          `json:"ca_id" db:"ca_id"`
	CertID      string          `json:"cert_id" db:"cert_id"`
	DHParamLen  int             `json:"dh_param_length" db:"dh_param_length"`
	Compression string          `json:"compression" db:"compression"`
	DNSServers  []string        `json:"dns_servers"`
	PushRoutes  []string        `json:"push_routes"`
	ClientToClient bool         `json:"client_to_client" db:"client_to_client"`
	MaxClients  int             `json:"max_clients" db:"max_clients"`
	Enabled     bool            `json:"enabled" db:"enabled"`
	Description string          `json:"description" db:"description"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type OpenVPNMode string

const (
	OVPNModePKI     OpenVPNMode = "pki"
	OVPNModeUserPass OpenVPNMode = "userpass"
	OVPNModeBoth    OpenVPNMode = "pki_userpass"
)

type OpenVPNClient struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	ServerAddr  string    `json:"server_addr" db:"server_addr"`
	ServerPort  int       `json:"server_port" db:"server_port"`
	Protocol    string    `json:"protocol" db:"protocol"`
	CAID        string    `json:"ca_id" db:"ca_id"`
	CertID      string    `json:"cert_id" db:"cert_id"`
	Username    string    `json:"username" db:"username"`
	Password    string    `json:"password,omitempty" db:"password"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	Description string    `json:"description" db:"description"`
}

// ─── WireGuard ─────────────────────────────────────────────────

type WireGuardInterface struct {
	ID         string          `json:"id" db:"id"`
	Name       string          `json:"name" db:"name"`
	PrivateKey string          `json:"private_key,omitempty" db:"private_key"`
	PublicKey  string          `json:"public_key" db:"public_key"`
	ListenPort int             `json:"listen_port" db:"listen_port"`
	Address    []string        `json:"addresses"`
	DNS        []string        `json:"dns"`
	Peers      []WireGuardPeer `json:"peers"`
	Enabled    bool            `json:"enabled" db:"enabled"`
	Description string         `json:"description" db:"description"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

type WireGuardPeer struct {
	ID                string   `json:"id" db:"id"`
	InterfaceID       string   `json:"interface_id" db:"interface_id"`
	PublicKey         string   `json:"public_key" db:"public_key"`
	PresharedKey      string   `json:"preshared_key,omitempty" db:"preshared_key"`
	Endpoint          string   `json:"endpoint" db:"endpoint"`
	AllowedIPs        []string `json:"allowed_ips"`
	PersistentKeepalive int    `json:"persistent_keepalive" db:"persistent_keepalive"`
	Description       string   `json:"description" db:"description"`
}

// ─── Common ────────────────────────────────────────────────────

type TunnelStatus struct {
	State        string    `json:"state"` // up, down, connecting, error
	Uptime       string    `json:"uptime"`
	BytesIn      uint64    `json:"bytes_in"`
	BytesOut     uint64    `json:"bytes_out"`
	ConnectedAt  time.Time `json:"connected_at"`
	RemoteIP     string    `json:"remote_ip"`
	LastHandshake time.Time `json:"last_handshake,omitempty"`
}

type VPNClientSession struct {
	Username    string    `json:"username"`
	RealAddress string    `json:"real_address"`
	VirtualIP   string    `json:"virtual_ip"`
	BytesIn     uint64    `json:"bytes_in"`
	BytesOut    uint64    `json:"bytes_out"`
	ConnectedAt time.Time `json:"connected_at"`
	TunnelID    string    `json:"tunnel_id"`
}
