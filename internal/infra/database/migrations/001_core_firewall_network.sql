-- NixGuard Schema: Core Firewall & Network
-- Migration 001 — Initial tables for firewall rules, aliases, NAT, interfaces, routes, gateways

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- ═══════════════════════════════════════════════════════════════
-- FIREWALL
-- ═══════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS firewall_rules (
    id              TEXT PRIMARY KEY,
    interface_name  TEXT NOT NULL DEFAULT '',
    direction       TEXT NOT NULL CHECK (direction IN ('in', 'out')),
    action          TEXT NOT NULL CHECK (action IN ('pass', 'block', 'reject')),
    protocol        TEXT NOT NULL DEFAULT 'any',
    source_type     TEXT NOT NULL DEFAULT 'any',
    source_value    TEXT NOT NULL DEFAULT '',
    source_port     TEXT NOT NULL DEFAULT '',
    source_not      INTEGER NOT NULL DEFAULT 0,
    dest_type       TEXT NOT NULL DEFAULT 'any',
    dest_value      TEXT NOT NULL DEFAULT '',
    dest_port       TEXT NOT NULL DEFAULT '',
    dest_not        INTEGER NOT NULL DEFAULT 0,
    log_enabled     INTEGER NOT NULL DEFAULT 0,
    description     TEXT NOT NULL DEFAULT '',
    enabled         INTEGER NOT NULL DEFAULT 1,
    rule_order      INTEGER NOT NULL DEFAULT 0,
    category        TEXT NOT NULL DEFAULT '',
    is_floating     INTEGER NOT NULL DEFAULT 0,
    floating_ifaces TEXT NOT NULL DEFAULT '',  -- JSON array
    gateway         TEXT NOT NULL DEFAULT '',
    state_type      TEXT NOT NULL DEFAULT 'keep',
    max_states      INTEGER NOT NULL DEFAULT 0,
    tag             TEXT NOT NULL DEFAULT '',
    tagged          TEXT NOT NULL DEFAULT '',
    schedule_name   TEXT NOT NULL DEFAULT '',
    schedule_start  TEXT NOT NULL DEFAULT '',
    schedule_end    TEXT NOT NULL DEFAULT '',
    schedule_days   TEXT NOT NULL DEFAULT '',  -- JSON array
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_firewall_rules_interface ON firewall_rules(interface_name);
CREATE INDEX idx_firewall_rules_order ON firewall_rules(rule_order);
CREATE INDEX idx_firewall_rules_enabled ON firewall_rules(enabled);

CREATE TABLE IF NOT EXISTS firewall_aliases (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL UNIQUE,
    alias_type      TEXT NOT NULL CHECK (alias_type IN ('host','network','port','url','url_table','geoip','nested')),
    description     TEXT NOT NULL DEFAULT '',
    entries         TEXT NOT NULL DEFAULT '[]',  -- JSON array
    update_freq     TEXT NOT NULL DEFAULT '',
    enabled         INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX idx_firewall_aliases_name ON firewall_aliases(name);

CREATE TABLE IF NOT EXISTS firewall_nat_rules (
    id              TEXT PRIMARY KEY,
    nat_type        TEXT NOT NULL CHECK (nat_type IN ('port_forward','one_to_one','outbound')),
    interface_name  TEXT NOT NULL,
    protocol        TEXT NOT NULL DEFAULT 'tcp',
    source_type     TEXT NOT NULL DEFAULT 'any',
    source_value    TEXT NOT NULL DEFAULT '',
    source_port     TEXT NOT NULL DEFAULT '',
    source_not      INTEGER NOT NULL DEFAULT 0,
    dest_type       TEXT NOT NULL DEFAULT 'any',
    dest_value      TEXT NOT NULL DEFAULT '',
    dest_port       TEXT NOT NULL DEFAULT '',
    dest_not        INTEGER NOT NULL DEFAULT 0,
    redirect_target TEXT NOT NULL DEFAULT '',
    redirect_port   TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    enabled         INTEGER NOT NULL DEFAULT 1,
    nat_reflection  INTEGER NOT NULL DEFAULT 0,
    rule_order      INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_nat_rules_type ON firewall_nat_rules(nat_type);

-- ═══════════════════════════════════════════════════════════════
-- NETWORK
-- ═══════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS network_interfaces (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL UNIQUE,
    alias_name      TEXT NOT NULL DEFAULT '',
    if_type         TEXT NOT NULL DEFAULT 'physical',
    enabled         INTEGER NOT NULL DEFAULT 1,
    description     TEXT NOT NULL DEFAULT '',
    mtu             INTEGER NOT NULL DEFAULT 1500,
    mac_address     TEXT NOT NULL DEFAULT '',
    ipv4_mode       TEXT NOT NULL DEFAULT '',   -- static, dhcp, ''
    ipv4_address    TEXT NOT NULL DEFAULT '',
    ipv4_gateway    TEXT NOT NULL DEFAULT '',
    ipv6_mode       TEXT NOT NULL DEFAULT '',
    ipv6_address    TEXT NOT NULL DEFAULT '',
    ipv6_gateway    TEXT NOT NULL DEFAULT '',
    vlan_parent     TEXT NOT NULL DEFAULT '',
    vlan_tag        INTEGER NOT NULL DEFAULT 0,
    bond_members    TEXT NOT NULL DEFAULT '[]',
    bond_mode       TEXT NOT NULL DEFAULT '',
    bridge_members  TEXT NOT NULL DEFAULT '[]',
    bridge_stp      INTEGER NOT NULL DEFAULT 0,
    pppoe_parent    TEXT NOT NULL DEFAULT '',
    pppoe_user      TEXT NOT NULL DEFAULT '',
    pppoe_pass      TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS network_routes (
    id              TEXT PRIMARY KEY,
    destination     TEXT NOT NULL,
    gateway         TEXT NOT NULL DEFAULT '',
    interface_name  TEXT NOT NULL DEFAULT '',
    metric          INTEGER NOT NULL DEFAULT 0,
    route_table     INTEGER NOT NULL DEFAULT 254,
    route_type      TEXT NOT NULL DEFAULT 'static',
    enabled         INTEGER NOT NULL DEFAULT 1,
    description     TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS network_gateways (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL UNIQUE,
    interface_name  TEXT NOT NULL,
    address         TEXT NOT NULL,
    protocol        TEXT NOT NULL DEFAULT 'inet',   -- inet, inet6
    monitor_ip      TEXT NOT NULL DEFAULT '',
    weight          INTEGER NOT NULL DEFAULT 1,
    priority        INTEGER NOT NULL DEFAULT 255,
    is_default      INTEGER NOT NULL DEFAULT 0,
    enabled         INTEGER NOT NULL DEFAULT 1,
    description     TEXT NOT NULL DEFAULT '',
    monitor_interval    INTEGER NOT NULL DEFAULT 5,
    loss_threshold      INTEGER NOT NULL DEFAULT 20,
    latency_threshold   INTEGER NOT NULL DEFAULT 500,
    down_count          INTEGER NOT NULL DEFAULT 3,
    monitor_method      TEXT NOT NULL DEFAULT 'icmp',
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS network_gateway_groups (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL UNIQUE,
    members         TEXT NOT NULL DEFAULT '[]',  -- JSON array of {gateway_id, tier, weight}
    trigger_level   TEXT NOT NULL DEFAULT 'member_down',
    description     TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);
