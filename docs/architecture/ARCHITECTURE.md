# NixGuard - Solution Architecture

## Overview

NixGuard is a production-grade Linux network security framework with 100% OPNsense feature parity.
Built with **Go backend** + **React frontend**, following Google Product Standards and Facebook Engineering Conventions.

---

## Architecture Principles

### Google Product Standards
1. **API-First Design** — All functionality exposed via well-defined gRPC + REST APIs
2. **Protocol Buffers** — Single source of truth for data contracts
3. **Hermetic Builds** — Reproducible builds with pinned dependencies
4. **Observability** — Structured logging, metrics, distributed tracing from day one
5. **Graceful Degradation** — System remains functional when individual modules fail
6. **Defense in Depth** — Security at every layer, not just the perimeter

### Facebook Engineering Conventions
1. **Module Isolation** — Each feature is a self-contained module with explicit boundaries
2. **Dependency Injection** — All dependencies injected via interfaces, never concrete types
3. **Consistent Naming** — snake_case for files, PascalCase for types, camelCase for functions
4. **Component-Based Frontend** — Reusable, composable UI components
5. **Centralized State** — Redux Toolkit for predictable state management
6. **Type Safety** — TypeScript on frontend, strong typing on backend

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT LAYER                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────┐  │
│  │ Web GUI  │  │   CLI    │  │ REST API │  │  gRPC Clients  │  │
│  │ (React)  │  │  (cobra) │  │ (curl)   │  │  (automation)  │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └──────┬─────────┘  │
│       │              │              │               │            │
└───────┼──────────────┼──────────────┼───────────────┼────────────┘
        │              │              │               │
┌───────┼──────────────┼──────────────┼───────────────┼────────────┐
│       ▼              ▼              ▼               ▼            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    API GATEWAY                            │   │
│  │  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌─────────────┐  │   │
│  │  │  Auth   │ │Rate Limit│ │  CORS    │ │  Audit Log  │  │   │
│  │  │Middleware│ │Middleware│ │Middleware│ │  Middleware  │  │   │
│  │  └─────────┘ └──────────┘ └──────────┘ └─────────────┘  │   │
│  └──────────────────────┬───────────────────────────────────┘   │
│                         │                                        │
│  ┌──────────────────────▼───────────────────────────────────┐   │
│  │                 APPLICATION LAYER                         │   │
│  │                                                           │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐       │   │
│  │  │Firewall │ │ Network │ │   VPN   │ │   DNS   │       │   │
│  │  │ Service │ │ Service │ │ Service │ │ Service │       │   │
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘       │   │
│  │  ┌────┴────┐ ┌────┴────┐ ┌────┴────┐ ┌────┴────┐       │   │
│  │  │  DHCP  │ │  IDS   │ │ Proxy  │ │   HA   │       │   │
│  │  │ Service │ │ Service │ │ Service │ │ Service │       │   │
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘       │   │
│  │  ┌────┴────┐ ┌────┴────┐ ┌────┴────┐ ┌────┴────┐       │   │
│  │  │Monitor │ │  Auth  │ │LoadBal │ │ System │       │   │
│  │  │ Service │ │ Service │ │ Service │ │ Service │       │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘       │   │
│  └──────────────────────┬───────────────────────────────────┘   │
│                         │                                        │
│  ┌──────────────────────▼───────────────────────────────────┐   │
│  │                   EVENT BUS (Internal)                    │   │
│  │            Async inter-module communication               │   │
│  └──────────────────────┬───────────────────────────────────┘   │
│                         │                                        │
│  ┌──────────────────────▼───────────────────────────────────┐   │
│  │               INFRASTRUCTURE LAYER                        │   │
│  │                                                           │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐       │   │
│  │  │nftables │ │Suricata │ │StrongSw │ │OpenVPN │       │   │
│  │  │ Adapter │ │ Adapter │ │ Adapter │ │ Adapter │       │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘       │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐       │   │
│  │  │WireGuar │ │ Unbound │ │  Kea   │ │HAProxy │       │   │
│  │  │ Adapter │ │ Adapter │ │ Adapter │ │ Adapter │       │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘       │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐       │   │
│  │  │  Squid  │ │   FRR   │ │NetFlow │ │  SNMP  │       │   │
│  │  │ Adapter │ │ Adapter │ │ Adapter │ │ Adapter │       │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘       │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│                     NIXGUARD SERVER                              │
└─────────────────────────────────────────────────────────────────┘
        │
        │  Unix Socket / gRPC
        ▼
┌─────────────────────────────────────────────────────────────────┐
│                     NIXGUARD AGENT (root)                        │
│                                                                  │
│  Privileged operations: nftables, routing, interface config,     │
│  service management, packet capture, system commands             │
│                                                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│  │ nftables │ │  iproute │ │ systemd  │ │ tcpdump  │          │
│  │ executor │ │ executor │ │ executor │ │ executor │          │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘          │
└─────────────────────────────────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────────┐
│                      LINUX KERNEL                                │
│  netfilter / nftables / conntrack / tc / ip route / netns       │
└─────────────────────────────────────────────────────────────────┘
```

---

## Security Architecture (Privilege Separation)

```
┌─────────────────────────────────────────┐
│  nixguard-server (unprivileged)         │
│  - Web GUI / API                         │
│  - Business logic                        │
│  - No direct system access               │
│  - Runs as: nixguard user                │
└──────────────┬──────────────────────────┘
               │ Unix Domain Socket (gRPC)
               │ Authenticated + Authorized
               ▼
┌─────────────────────────────────────────┐
│  nixguard-agent (privileged)            │
│  - Executes system commands              │
│  - Manages nftables/iptables             │
│  - Controls services (systemd)           │
│  - Network interface configuration       │
│  - Runs as: root                         │
│  - Minimal attack surface                │
│  - Command whitelist only                │
└─────────────────────────────────────────┘
```

---

## Module Architecture (Clean Architecture per Module)

Each module follows the same 4-layer pattern:

```
module/
├── domain/          # Entities, Value Objects, Domain Events
│   ├── model.go     # Core domain models
│   ├── event.go     # Domain events
│   └── errors.go    # Domain-specific errors
│
├── app/             # Application Services (Use Cases)
│   ├── service.go   # Business logic orchestration
│   ├── dto.go       # Data Transfer Objects (input/output)
│   └── port.go      # Port interfaces (what infra must implement)
│
├── infra/           # Infrastructure (external system adapters)
│   ├── adapter.go   # Implements port interfaces
│   ├── config.go    # Adapter configuration
│   └── mapper.go    # Domain <-> External model mapping
│
└── adapter/         # Delivery (HTTP/gRPC/CLI handlers)
    ├── handler.go   # Request handlers
    ├── request.go   # Request validation/parsing
    └── response.go  # Response formatting
```

---

## Data Flow

```
HTTP Request
    │
    ▼
[Router] → [Auth Middleware] → [Rate Limiter] → [Audit Logger]
    │
    ▼
[Handler] — validates request, creates DTO
    │
    ▼
[App Service] — orchestrates business logic
    │
    ├──→ [Domain Model] — enforces invariants
    │
    ├──→ [Repository] — persists state (SQLite/PostgreSQL)
    │
    ├──→ [Infra Adapter] — calls external system (nftables, etc.)
    │         │
    │         ▼
    │    [Agent Client] — sends command to privileged agent
    │         │
    │         ▼
    │    [Agent] — executes on Linux kernel
    │
    └──→ [Event Bus] — notifies other modules
              │
              ▼
         [Subscribers] — react to domain events
```

---

## Technology Stack

| Layer              | Technology                    | Rationale                              |
|--------------------|-------------------------------|----------------------------------------|
| Language (Backend) | Go 1.22+                      | Performance, concurrency, static binary|
| Language (Frontend)| TypeScript + React 18         | Type safety, component ecosystem       |
| API Protocol       | gRPC + REST (grpc-gateway)    | Performance + browser compatibility    |
| Data Contracts     | Protocol Buffers v3           | Single source of truth for API         |
| Database           | SQLite (single) / PostgreSQL  | Simple default, scale when needed      |
| ORM                | sqlc                          | Type-safe SQL, no magic                |
| Migrations         | golang-migrate                | Versioned schema migrations            |
| Event Bus          | In-process (channels)         | No external dependency needed          |
| Logging            | slog (stdlib)                 | Structured, zero-dependency            |
| Metrics            | Prometheus client             | Industry standard observability        |
| Frontend State     | Redux Toolkit                 | Predictable, scalable state            |
| Frontend UI        | Ant Design / Tailwind CSS     | Enterprise-grade components            |
| Build              | Go toolchain + Vite           | Fast, reliable builds                  |
| Container          | Docker multi-stage            | Minimal production image               |
| Testing            | Go testing + React Testing Lib| Native tooling, no bloat               |
