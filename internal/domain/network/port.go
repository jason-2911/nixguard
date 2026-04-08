package network

import "context"

// InterfaceRepository persists interface configurations.
type InterfaceRepository interface {
	List(ctx context.Context) ([]Interface, error)
	GetByID(ctx context.Context, id string) (*Interface, error)
	GetByName(ctx context.Context, name string) (*Interface, error)
	Create(ctx context.Context, iface *Interface) error
	Update(ctx context.Context, iface *Interface) error
	Delete(ctx context.Context, id string) error
}

// RouteRepository persists route configurations.
type RouteRepository interface {
	List(ctx context.Context) ([]Route, error)
	GetByID(ctx context.Context, id string) (*Route, error)
	Create(ctx context.Context, route *Route) error
	Update(ctx context.Context, route *Route) error
	Delete(ctx context.Context, id string) error
}

// GatewayRepository persists gateway configurations.
type GatewayRepository interface {
	List(ctx context.Context) ([]Gateway, error)
	GetByID(ctx context.Context, id string) (*Gateway, error)
	Create(ctx context.Context, gw *Gateway) error
	Update(ctx context.Context, gw *Gateway) error
	Delete(ctx context.Context, id string) error
}

// GatewayGroupRepository persists gateway group configurations.
type GatewayGroupRepository interface {
	List(ctx context.Context) ([]GatewayGroup, error)
	GetByID(ctx context.Context, id string) (*GatewayGroup, error)
	Create(ctx context.Context, group *GatewayGroup) error
	Update(ctx context.Context, group *GatewayGroup) error
	Delete(ctx context.Context, id string) error
}

// NetworkEngine is the interface to Linux networking subsystem (iproute2).
type NetworkEngine interface {
	// Interface operations
	GetInterfaceStatus(ctx context.Context, name string) (*InterfaceStatus, error)
	SetInterfaceUp(ctx context.Context, name string) error
	SetInterfaceDown(ctx context.Context, name string) error
	SetInterfaceAddress(ctx context.Context, name string, addr string) error
	SetInterfaceMTU(ctx context.Context, name string, mtu int) error

	// VLAN operations
	CreateVLAN(ctx context.Context, parent string, tag int) (string, error)
	DeleteVLAN(ctx context.Context, name string) error

	// Bond operations
	CreateBond(ctx context.Context, name string, cfg BondConfig) error
	DeleteBond(ctx context.Context, name string) error

	// Bridge operations
	CreateBridge(ctx context.Context, name string, cfg BridgeConfig) error
	DeleteBridge(ctx context.Context, name string) error

	// Route operations
	AddRoute(ctx context.Context, route Route) error
	DeleteRoute(ctx context.Context, route Route) error
	ListRoutes(ctx context.Context, table int) ([]Route, error)

	// Gateway monitoring
	PingGateway(ctx context.Context, addr string) (*GatewayStatus, error)

	// CheckGatewayTCP checks gateway via TCP connect.
	CheckGatewayTCP(ctx context.Context, addr string, port int) (*GatewayStatus, error)

	// CheckGatewayHTTP checks gateway via HTTP GET.
	CheckGatewayHTTP(ctx context.Context, url string) (*GatewayStatus, error)

	// Sysctl operations
	SetSysctl(ctx context.Context, key, value string) error

	// Policy routing
	AddPolicyRoute(ctx context.Context, mark uint32, gateway string, table int) error
	DeletePolicyRoute(ctx context.Context, mark uint32, table int) error
}
