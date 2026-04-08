# NixGuard Architecture

**Version:** 1.0 | **Last Updated:** 2026-04-08

---

## 1. Project Overview

NixGuard is a Linux network security framework targeting 100% OPNsense feature parity. Built with Go (backend) + React/TypeScript (frontend), it manages firewall rules, NAT, routing, VPN, and network services through a web GUI and REST API. The system uses **nftables** as the primary firewall engine and follows **clean architecture** principles.

---

## 2. File Structure

```
nixguard/
├── cmd/                              # Binary entry points
│   ├── nixguard-server/              # Main API server (unprivileged)
│   │   ├── main.go                   # HTTP server bootstrap, graceful shutdown
│   │   └── app.go                    # Module registry, dependency injection
│   ├── nixguard-agent/               # Privileged executor (runs as root)
│   │   └── main.go                   # Unix socket listener, command executor
│   ├── nixguard-cli/                 # CLI tool (stub)
│   │   └── main.go
│   └── nixguard-worker/              # Background worker (stub)
│       └── main.go
│
├── internal/                         # Private application code
│   ├── domain/                       # Domain models & ports (interfaces)
│   │   ├── firewall/
│   │   │   ├── model.go              # Rule, Alias, NATRule, State, Address (212 lines)
│   │   │   └── port.go               # Repository & Engine interfaces (89 lines)
│   │   ├── network/
│   │   │   ├── model.go              # Interface, Route, Gateway, GatewayGroup (172 lines)
│   │   │   └── port.go               # Repository & Engine interfaces
│   │   ├── auth/model.go             # User, Group, Session, Certificate (120 lines)
│   │   ├── dns/model.go              # ResolverConfig, HostOverride, Blocklist (93 lines)
│   │   ├── dhcp/model.go             # ServerConfig, Lease, StaticMapping (84 lines)
│   │   ├── vpn/model.go              # IPsec, OpenVPN, WireGuard configs (159 lines)
│   │   ├── ids/model.go              # Suricata config, Alert, Ruleset (86 lines)
│   │   ├── proxy/model.go            # Squid config, ACL, URLFilter (60 lines)
│   │   ├── loadbalancer/model.go     # HAProxy Frontend/Backend (113 lines)
│   │   ├── ha/model.go               # CARP VirtualIP, HAConfig (49 lines)
│   │   ├── traffic_shaper/model.go   # Pipe, Queue, ShaperRule (57 lines)
│   │   ├── captiveportal/model.go    # Zone, Voucher, PortalSession (66 lines)
│   │   ├── monitor/model.go          # SystemMetrics, LogEntry, AlertRule (153 lines)
│   │   └── system/model.go           # Backup, Settings, Diagnostics (90 lines)
│   │
│   ├── app/                          # Application services (use cases)
│   │   ├── firewall/
│   │   │   ├── service.go            # Rule/Alias/NAT/State management (631 lines)
│   │   │   └── dto.go                # Input/output DTOs (105 lines)
│   │   └── network/
│   │       ├── service.go            # Interface/Route/Gateway management (505 lines)
│   │       └── dto.go                # Input/output DTOs
│   │
│   ├── adapter/                      # Inbound adapters (HTTP/gRPC)
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── firewall/handler.go   # REST endpoints for firewall (394 lines)
│   │       │   └── network/handler.go    # REST endpoints for network (219 lines)
│   │       ├── middleware/               # Recovery, RequestID, Logger, CORS
│   │       └── router/                   # Mux route registration
│   │
│   ├── infra/                        # Outbound adapters (infrastructure)
│   │   ├── nftables/
│   │   │   ├── adapter.go            # FirewallEngine implementation (400+ lines)
│   │   │   ├── compiler.go           # nft ruleset compiler (750+ lines)
│   │   │   ├── predefined.go         # Bogon/RFC1918 predefined nft sets
│   │   │   ├── adapter_test.go       # Adapter tests
│   │   │   └── compiler_test.go      # Compiler tests (25 test cases)
│   │   ├── iproute2/
│   │   │   └── adapter.go            # NetworkEngine implementation (500+ lines)
│   │   ├── geoip/
│   │   │   └── maxmind.go            # GeoIPProvider — MaxMind GeoLite2 download/parse
│   │   ├── database/
│   │   │   ├── sqlite/
│   │   │   │   ├── firewall_rule_repo.go     # Firewall CRUD repository
│   │   │   │   ├── firewall_rule_repo_test.go
│   │   │   │   └── network_repo.go           # Network CRUD repository
│   │   │   └── migrations/
│   │   │       └── 001_core_firewall_network.sql  # Schema (163 lines)
│   │   ├── strongswan/adapter.go     # IPsec VPN adapter (143 lines, partial)
│   │   ├── wireguard/adapter.go      # WireGuard adapter (126 lines, partial)
│   │   ├── unbound/adapter.go        # DNS resolver adapter (200 lines, partial)
│   │   ├── suricata/adapter.go       # IDS adapter (stub)
│   │   └── haproxy/adapter.go        # Load balancer adapter (stub)
│   │
│   ├── config/                       # Configuration loading
│   │   └── config.go                 # YAML config parser
│   ├── event/                        # Event bus
│   │   └── bus.go                    # Pub/sub for inter-module communication
│   └── plugin/                       # Plugin framework (stub)
│       └── registry.go
│
├── pkg/                              # Public shared libraries
│   ├── executor/                     # Safe command execution (whitelisted)
│   │   └── safe.go
│   ├── logger/                       # Structured logging (slog)
│   │   └── logger.go
│   ├── crypto/                       # Random ID generation
│   │   └── id.go
│   ├── validator/                    # Input validation
│   │   └── validator.go
│   ├── netutil/                      # CIDR parsing, IP helpers
│   │   ├── cidr.go
│   │   └── bogons.go                # BogonRangesV4/V6, RFC1918Ranges exports
│   ├── errors/                       # Error types
│   ├── health/                       # Health check
│   └── version/                      # Build version info
│
├── web/                              # React frontend
│   ├── src/
│   │   ├── api/client.ts             # Axios HTTP client
│   │   ├── app/
│   │   │   ├── App.tsx               # Router, lazy loading
│   │   │   └── main.tsx              # React entry point
│   │   ├── components/
│   │   │   ├── layout/MainLayout.tsx # Sidebar + header + outlet
│   │   │   ├── common/              # Reusable components
│   │   │   ├── forms/               # Form components
│   │   │   └── widgets/             # UI widgets
│   │   ├── hooks/
│   │   │   ├── useStore.ts          # Typed Redux hooks
│   │   │   └── useAutoRefresh.ts    # Auto-polling hook (5s interval)
│   │   ├── pages/                   # 30+ page components
│   │   │   ├── dashboard/DashboardPage.tsx
│   │   │   ├── firewall/
│   │   │   │   ├── RulesPage.tsx    # Rule CRUD + apply
│   │   │   │   ├── AliasesPage.tsx  # Alias management
│   │   │   │   ├── NATPage.tsx      # NAT rule management
│   │   │   │   └── TrafficPage.tsx  # Live packet capture
│   │   │   ├── network/
│   │   │   │   ├── InterfacesPage.tsx
│   │   │   │   ├── RoutingPage.tsx
│   │   │   │   └── GatewaysPage.tsx
│   │   │   └── [vpn, dns, dhcp, ids, monitor, system, auth]/ # Stub pages
│   │   ├── store/
│   │   │   ├── store.ts             # Redux store config
│   │   │   └── slices/
│   │   │       ├── firewallSlice.ts # Firewall state + async thunks
│   │   │       ├── authSlice.ts     # Auth state
│   │   │       └── uiSlice.ts       # Dark mode, sidebar, notifications
│   │   └── types/
│   │       ├── firewall.ts          # TypeScript interfaces
│   │       └── network.ts
│   ├── package.json                 # React 18, Ant Design 5, Vite 5
│   └── vite.config.ts
│
├── configs/defaults/server.yaml     # Default server configuration
├── data/
│   ├── nftables/nixguard.nft        # Generated nftables ruleset
│   ├── geoip/                        # MaxMind GeoLite2 database cache
│   ├── nixguard.db                  # SQLite database
│   └── pcap/                        # Captured traffic files
├── deployments/
│   ├── docker/                      # Dockerfile + docker-compose
│   ├── systemd/                     # Service units
│   ├── ansible/                     # Deployment playbooks
│   └── terraform/                   # IaC modules
├── api/
│   ├── proto/                       # gRPC definitions
│   └── openapi/                     # REST API specs
├── docs/                            # Architecture, API, guides
├── test/                            # Integration & E2E tests
├── Makefile                         # Build targets
├── go.mod / go.sum                  # Go module dependencies
├── CODEX.md                         # Feature requirements (OPNsense parity)
└── ARCHITECTURE.md                  # This file
```

---

## 3. Key Modules & Dependencies

### 3.1 Dependency Graph

```
cmd/nixguard-server
  └── internal/config           (YAML config)
  └── internal/event            (event bus)
  └── internal/app/firewall     (use cases)
  │     └── internal/domain/firewall  (models + ports)
  │     └── internal/infra/nftables   (nftables adapter)
  │     └── internal/infra/database   (SQLite repository)
  └── internal/app/network      (use cases)
  │     └── internal/domain/network   (models + ports)
  │     └── internal/infra/iproute2   (iproute2 adapter)
  │     └── internal/infra/database   (SQLite repository)
  └── internal/adapter/http     (REST handlers + middleware + router)
  └── pkg/executor              (safe command execution)
  └── pkg/logger                (structured logging)
  └── pkg/crypto                (ID generation)
  └── pkg/validator             (input validation)
```

### 3.2 Module Status

| Module | Domain | Service | HTTP | Infra | DB | Frontend | Status |
|--------|:------:|:-------:|:----:|:-----:|:--:|:--------:|--------|
| **Firewall** | 220L | 700L | 394L | 1400L | 80L | 244L | **Production** |
| **Network** | 180L | 520L | 219L | 550L | 80L | 200L | **Production** |
| **GeoIP** | - | - | - | 250L | - | - | **Production** |
| DNS | 93L | - | - | 200L* | - | stub | Partial |
| VPN | 159L | - | - | 269L* | - | stub | Partial |
| DHCP | 84L | - | - | - | - | stub | Model only |
| IDS | 86L | - | - | - | - | stub | Model only |
| Proxy | 60L | - | - | - | - | stub | Model only |
| Load Balancer | 113L | - | - | - | - | stub | Model only |
| HA | 49L | - | - | - | - | stub | Model only |
| Traffic Shaper | 57L | - | - | - | - | stub | Model only |
| Captive Portal | 66L | - | - | - | - | stub | Model only |
| Monitoring | 153L | - | - | - | - | stub | Model only |
| Auth | 120L | - | - | - | - | stub | Model only |
| System | 90L | - | - | - | - | stub | Model only |

*Partial infrastructure adapters

### 3.3 External Tool Dependencies

| Tool | Module | Usage |
|------|--------|-------|
| `nft` | Firewall | Atomic ruleset application |
| `conntrack` | Firewall | Connection state table |
| `tcpdump` | Firewall | Live traffic capture |
| `ip` | Network | Interface/route/link management |
| `ethtool` | Network | Interface speed/duplex stats |
| `wg` | VPN | WireGuard interface management |
| `swanctl` | VPN | StrongSwan IPsec |
| `unbound` | DNS | Recursive DNS resolver |

### 3.4 Frontend Stack

| Library | Version | Purpose |
|---------|---------|---------|
| React | 18.3.0 | UI framework |
| TypeScript | 5.4.0 | Type safety |
| Ant Design | 5.15.0 | Component library |
| Redux Toolkit | 2.2.0 | State management |
| Axios | 1.6.7 | HTTP client |
| Vite | 5.2.0 | Build tool |
| Recharts | 2.12.0 | Traffic graphs |
| React Router | 6.22.0 | Client routing |

---

## 4. Design Patterns

### 4.1 Clean Architecture (Ports & Adapters)

```
          HTTP Request
               │
    ┌──────────▼──────────┐
    │   adapter/http/     │  Inbound adapter — translates HTTP to DTOs
    │   handler/firewall  │
    └──────────┬──────────┘
               │ calls
    ┌──────────▼──────────┐
    │   app/firewall/     │  Application service — orchestrates use cases
    │   service.go        │  Validates input, coordinates domain + infra
    └──────────┬──────────┘
          ┌────┴────┐
          │         │
    ┌─────▼─────┐ ┌─▼──────────────┐
    │ domain/   │ │ infra/nftables/ │  Outbound adapters
    │ firewall/ │ │ infra/database/ │  Implement domain ports
    │ port.go   │ └────────────────┘
    └───────────┘
```

**Key principle:** Domain layer defines interfaces (ports). Infrastructure layer provides implementations. Application layer orchestrates. No layer depends inward.

### 4.2 Repository Pattern

```go
// Domain port (interface)
type RuleRepository interface {
    FindAll(ctx context.Context, filter RuleFilter) ([]Rule, error)
    FindByID(ctx context.Context, id string) (*Rule, error)
    Create(ctx context.Context, rule *Rule) error
    Update(ctx context.Context, rule *Rule) error
    Delete(ctx context.Context, id string) error
    Reorder(ctx context.Context, ids []string) error
}

// Infrastructure implementation (SQLite)
type sqliteRuleRepo struct { db *sql.DB }
func (r *sqliteRuleRepo) Create(ctx context.Context, rule *firewall.Rule) error { ... }
```

### 4.3 Engine Pattern (Strategy)

```go
// Domain port
type FirewallEngine interface {
    ApplyRules(ctx context.Context, rules []Rule, aliases []Alias) error
    ApplyNAT(ctx context.Context, rules []NATRule) error
    GetStates(ctx context.Context, filter StateFilter) ([]State, error)
    FlushStates(ctx context.Context, filter StateFilter) error
    CaptureTraffic(ctx context.Context, filter TrafficFilter) ([]CapturedPacket, error)
    GetRuleStats(ctx context.Context) (map[string]RuleStats, error)
}

// Infrastructure implementation (nftables)
type Adapter struct { exec *executor.Safe; log *slog.Logger }
func (a *Adapter) ApplyRules(ctx context.Context, ...) error {
    ruleset := CompileRuleset(rules, aliases)    // Generate nft script
    writeRulesetFile(rulesetPath, ruleset)        // Write to disk
    a.exec.Run(ctx, "nft", "-f", rulesetPath)    // Atomic apply
}
```

### 4.4 Atomic Ruleset Application

Rules are compiled into a complete nftables script and applied atomically with `nft -f`:

```
Rule CRUD → DB persist → Fetch all rules → Compile to nft → Write file → nft -f (atomic)
```

This ensures the kernel always has a consistent ruleset. No partial states.

### 4.5 Privilege Separation

```
┌────────────────────────┐     Unix Socket     ┌─────────────────────┐
│  nixguard-server       │◄────────────────────►│  nixguard-agent     │
│  (unprivileged user)   │                      │  (root)             │
│                        │                      │                     │
│  - Web GUI / API       │   executes via       │  - nft -f           │
│  - Business logic      │   Safe Executor      │  - conntrack        │
│  - Database            │                      │  - ip link/route    │
│  - Authentication      │                      │  - tcpdump          │
└────────────────────────┘                      └─────────────────────┘
```

### 4.6 Event-Driven Communication

```go
// Event types
const (
    FirewallRuleCreated  = "firewall.rule.created"
    FirewallRulesApplied = "firewall.rules.applied"
    GatewayDown          = "network.gateway.down"
    GatewayUp            = "network.gateway.up"
)

// Publishing
bus.Publish(event.FirewallRuleCreated, rule)

// Subscribing (other modules react)
bus.Subscribe(event.GatewayDown, func(data any) {
    // Trigger failover logic
})
```

### 4.7 Frontend Patterns

- **Redux Toolkit** for global state (firewall rules, auth)
- **Local state** (`useState`) for page-specific data (aliases, NAT)
- **Auto-refresh hook** (`useAutoRefresh`) for polling every 5s
- **Ant Design** tables with inline editing, modals for CRUD

---

## 5. Entry Points

### 5.1 Server (`cmd/nixguard-server/main.go`)

```
main()
  ├── config.Load("configs/defaults/server.yaml")
  ├── logger.New()
  ├── event.NewBus()
  ├── initApp(cfg, bus, log)           # app.go
  │     ├── database.Open(cfg.Database.DSN)
  │     ├── database.Migrate()
  │     ├── nftables.NewAdapter()      # Firewall engine
  │     ├── iproute2.NewAdapter()      # Network engine
  │     ├── firewallService.New()      # Application service
  │     ├── networkService.New()       # Application service
  │     ├── firewallHandler.New()      # HTTP handler
  │     ├── networkHandler.New()       # HTTP handler
  │     └── router.New()              # Mux with middleware
  ├── http.Server{Addr: ":8443", Handler: app.Router()}
  ├── srv.ListenAndServe()            # Start serving
  └── signal.NotifyContext(SIGINT, SIGTERM)  # Graceful shutdown
```

### 5.2 REST API Routes

```
/api/v1/firewall/
  ├── GET    /rules                    List all rules
  ├── GET    /rules/{id}               Get single rule
  ├── POST   /rules                    Create rule (auto-applies)
  ├── PUT    /rules/{id}               Update rule (auto-applies)
  ├── DELETE /rules/{id}               Delete rule (auto-applies)
  ├── POST   /rules/reorder            Reorder rules
  ├── POST   /apply                    Force recompile & apply
  ├── GET    /aliases                  List aliases
  ├── POST   /aliases                  Create alias
  ├── PUT    /aliases/{id}             Update alias
  ├── DELETE /aliases/{id}             Delete alias
  ├── GET    /nat                      List NAT rules
  ├── POST   /nat                      Create NAT rule
  ├── PUT    /nat/{id}                 Update NAT rule
  ├── DELETE /nat/{id}                 Delete NAT rule
  ├── GET    /states                   List conntrack entries
  ├── DELETE /states                   Flush conntrack
  ├── GET    /traffic                  Capture packets
  └── POST   /traffic/export           Export PCAP

/api/v1/network/
  ├── GET    /interfaces               List interfaces + status
  ├── GET    /interfaces/{id}          Get interface details
  ├── POST   /interfaces               Create interface
  ├── DELETE /interfaces/{id}          Delete interface
  ├── GET    /routes                   List routes
  ├── POST   /routes                   Create route
  ├── DELETE /routes/{id}              Delete route
  ├── GET    /gateways                 List gateways
  ├── POST   /gateways                Create gateway
  ├── PUT    /gateways/{id}            Update gateway
  └── DELETE /gateways/{id}            Delete gateway
```

### 5.3 Frontend Routes

```
/                           → Dashboard
/firewall/rules             → Firewall Rules (CRUD + Apply)
/firewall/aliases           → Alias Management
/firewall/nat               → NAT Rules
/firewall/traffic           → Live Traffic Capture
/network/interfaces         → Interface Management
/network/routing            → Static/Dynamic Routes
/network/gateways           → Gateway Management
/vpn/*                      → VPN pages (stub)
/dns/*                      → DNS pages (stub)
/dhcp/*                     → DHCP pages (stub)
/ids/*                      → IDS pages (stub)
/monitor/*                  → Monitoring pages (stub)
/system/*                   → System pages (stub)
/auth/login                 → Login page
```

### 5.4 Database Schema

```sql
-- Core Firewall
firewall_rules       (id, interface, direction, action, protocol, source/dest, log, order, ...)
firewall_aliases     (id, name UNIQUE, type, entries JSON, enabled, ...)
firewall_nat_rules   (id, type, interface, protocol, source/dest, redirect, ...)

-- Network
network_interfaces   (id, name UNIQUE, type, ipv4/ipv6 config, mtu, ...)
network_routes       (id, type, destination, gateway, metric, table, ...)
network_gateways     (id, name, interface, address, monitor config, priority, ...)
network_gateway_groups (id, name, members JSON, ...)
```

### 5.5 Build & Run

```bash
make build              # Build all binaries
make dev-server         # Run API server (dev mode)
make dev-web            # Run React frontend (Vite dev server)
make dev                # Run both concurrently
make test               # Unit tests
make test-integration   # Integration tests
make db-migrate         # Apply database migrations

# Docker
cd deployments/docker && docker-compose up -d
```
