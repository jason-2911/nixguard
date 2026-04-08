package vpn

import "context"

// IPsecRepository persists IPsec tunnel configurations.
type IPsecRepository interface {
	List(ctx context.Context) ([]IPsecTunnel, error)
	GetByID(ctx context.Context, id string) (*IPsecTunnel, error)
	Create(ctx context.Context, tunnel *IPsecTunnel) error
	Update(ctx context.Context, tunnel *IPsecTunnel) error
	Delete(ctx context.Context, id string) error
}

// OpenVPNServerRepository persists OpenVPN server configurations.
type OpenVPNServerRepository interface {
	List(ctx context.Context) ([]OpenVPNServer, error)
	GetByID(ctx context.Context, id string) (*OpenVPNServer, error)
	Create(ctx context.Context, server *OpenVPNServer) error
	Update(ctx context.Context, server *OpenVPNServer) error
	Delete(ctx context.Context, id string) error
}

// WireGuardRepository persists WireGuard configurations.
type WireGuardRepository interface {
	List(ctx context.Context) ([]WireGuardInterface, error)
	GetByID(ctx context.Context, id string) (*WireGuardInterface, error)
	Create(ctx context.Context, wg *WireGuardInterface) error
	Update(ctx context.Context, wg *WireGuardInterface) error
	Delete(ctx context.Context, id string) error
}

// IPsecEngine interfaces with StrongSwan.
type IPsecEngine interface {
	ApplyTunnel(ctx context.Context, tunnel IPsecTunnel) error
	RemoveTunnel(ctx context.Context, id string) error
	GetStatus(ctx context.Context) ([]TunnelStatus, error)
	RestartDaemon(ctx context.Context) error
}

// OpenVPNEngine interfaces with OpenVPN process.
type OpenVPNEngine interface {
	ApplyServer(ctx context.Context, server OpenVPNServer) error
	RemoveServer(ctx context.Context, id string) error
	GetServerStatus(ctx context.Context, id string) (*TunnelStatus, error)
	GetClients(ctx context.Context, serverID string) ([]VPNClientSession, error)
	KillClient(ctx context.Context, serverID, username string) error
	GenerateClientConfig(ctx context.Context, serverID, username string) ([]byte, error)
}

// WireGuardEngine interfaces with WireGuard kernel module.
type WireGuardEngine interface {
	ApplyInterface(ctx context.Context, wg WireGuardInterface) error
	RemoveInterface(ctx context.Context, name string) error
	GetStatus(ctx context.Context, name string) (*TunnelStatus, error)
	GenerateKeyPair(ctx context.Context) (privateKey, publicKey string, err error)
	GeneratePresharedKey(ctx context.Context) (string, error)
}
