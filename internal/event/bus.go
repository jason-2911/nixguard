// Package event provides an in-process event bus for inter-module communication.
// Modules publish domain events; other modules subscribe to react to them.
// This decouples modules while allowing coordinated behavior.
//
// Example: When firewall module creates a rule, it publishes FirewallRuleCreated.
// The monitor module subscribes to log the change. The HA module subscribes to
// sync the rule to the backup node.
package event

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Event represents a domain event published by a module.
type Event struct {
	// Type identifies the event (e.g., "firewall.rule.created").
	Type string

	// Source is the module that published the event.
	Source string

	// Timestamp is when the event was created.
	Timestamp time.Time

	// Payload contains event-specific data.
	Payload interface{}
}

// Handler processes an event. Handlers must be idempotent.
type Handler func(ctx context.Context, event Event) error

// Bus is the central event dispatcher.
type Bus struct {
	mu       sync.RWMutex
	log      *slog.Logger
	handlers map[string][]Handler
	closed   bool
}

// NewBus creates a new event bus.
func NewBus(log *slog.Logger) *Bus {
	return &Bus{
		log:      log,
		handlers: make(map[string][]Handler),
	}
}

// Subscribe registers a handler for a specific event type.
// Use "*" to subscribe to all events.
func (b *Bus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
	b.log.Debug("event subscriber registered",
		slog.String("event_type", eventType),
	)
}

// Publish dispatches an event to all registered handlers.
// Handlers are called asynchronously. Errors are logged but don't propagate.
func (b *Bus) Publish(ctx context.Context, evt Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	evt.Timestamp = time.Now()

	// Specific handlers
	for _, h := range b.handlers[evt.Type] {
		go b.safeCall(ctx, h, evt)
	}

	// Wildcard handlers
	for _, h := range b.handlers["*"] {
		go b.safeCall(ctx, h, evt)
	}
}

func (b *Bus) safeCall(ctx context.Context, h Handler, evt Event) {
	defer func() {
		if r := recover(); r != nil {
			b.log.Error("event handler panic",
				slog.String("event_type", evt.Type),
				slog.Any("panic", r),
			)
		}
	}()

	if err := h(ctx, evt); err != nil {
		b.log.Error("event handler error",
			slog.String("event_type", evt.Type),
			slog.String("error", err.Error()),
		)
	}
}

// Close stops the event bus from accepting new events.
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
}

// ─── Well-Known Event Types ────────────────────────────────────

const (
	// Firewall events
	FirewallRuleCreated  = "firewall.rule.created"
	FirewallRuleUpdated  = "firewall.rule.updated"
	FirewallRuleDeleted  = "firewall.rule.deleted"
	FirewallRulesApplied = "firewall.rules.applied"

	// Network events
	InterfaceUp       = "network.interface.up"
	InterfaceDown     = "network.interface.down"
	GatewayUp         = "network.gateway.up"
	GatewayDown       = "network.gateway.down"
	RouteChanged      = "network.route.changed"

	// VPN events
	VPNTunnelUp       = "vpn.tunnel.up"
	VPNTunnelDown     = "vpn.tunnel.down"
	VPNClientConnect  = "vpn.client.connected"
	VPNClientDisconnect = "vpn.client.disconnected"

	// DNS events
	DNSConfigChanged  = "dns.config.changed"
	DNSBlocklistUpdated = "dns.blocklist.updated"

	// DHCP events
	DHCPLeaseCreated  = "dhcp.lease.created"
	DHCPLeaseExpired  = "dhcp.lease.expired"

	// IDS events
	IDSAlertTriggered = "ids.alert.triggered"
	IDSRulesUpdated   = "ids.rules.updated"

	// System events
	ConfigChanged     = "system.config.changed"
	BackupCreated     = "system.backup.created"
	UpdateAvailable   = "system.update.available"
	ServiceRestarted  = "system.service.restarted"

	// HA events
	HAFailover        = "ha.failover"
	HAStateSync       = "ha.state.sync"

	// Auth events
	UserLogin         = "auth.user.login"
	UserLogout        = "auth.user.logout"
	UserLoginFailed   = "auth.user.login_failed"
)
