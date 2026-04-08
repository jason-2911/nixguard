package dhcp

import "context"

type ServerConfigRepository interface {
	List(ctx context.Context) ([]ServerConfig, error)
	GetByInterface(ctx context.Context, iface string) (*ServerConfig, error)
	Save(ctx context.Context, cfg *ServerConfig) error
	Delete(ctx context.Context, id string) error
}

type StaticMappingRepository interface {
	List(ctx context.Context, iface string) ([]StaticMapping, error)
	GetByID(ctx context.Context, id string) (*StaticMapping, error)
	Create(ctx context.Context, mapping *StaticMapping) error
	Update(ctx context.Context, mapping *StaticMapping) error
	Delete(ctx context.Context, id string) error
}

// DHCPEngine interfaces with ISC DHCP or Kea.
type DHCPEngine interface {
	ApplyConfig(ctx context.Context, configs []ServerConfig, mappings []StaticMapping) error
	GetLeases(ctx context.Context, iface string) ([]Lease, error)
	DeleteLease(ctx context.Context, ip string) error
	Restart(ctx context.Context) error
	GetStats(ctx context.Context) (*DHCPStats, error)
}

type DHCPStats struct {
	TotalLeases   int            `json:"total_leases"`
	ActiveLeases  int            `json:"active_leases"`
	ExpiredLeases int            `json:"expired_leases"`
	PoolUsage     map[string]int `json:"pool_usage"` // interface -> percent used
}
