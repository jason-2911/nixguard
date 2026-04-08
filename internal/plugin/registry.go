// Package plugin provides the NixGuard plugin system.
// Plugins extend NixGuard functionality without modifying core code.
// Each plugin implements the Plugin interface and registers via the Registry.
package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Plugin defines the interface that all NixGuard plugins must implement.
type Plugin interface {
	// Metadata returns plugin identification and requirements.
	Metadata() Metadata

	// Init initializes the plugin with the provided context.
	Init(ctx context.Context, host Host) error

	// Start activates the plugin.
	Start(ctx context.Context) error

	// Stop deactivates the plugin gracefully.
	Stop(ctx context.Context) error

	// Health returns the plugin's health status.
	Health(ctx context.Context) HealthStatus
}

// Metadata describes a plugin.
type Metadata struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Author       string   `json:"author"`
	Dependencies []string `json:"dependencies"`
	Category     Category `json:"category"`
}

// Category classifies plugins.
type Category string

const (
	CategoryNetwork   Category = "network"
	CategorySecurity  Category = "security"
	CategoryMonitor   Category = "monitoring"
	CategoryService   Category = "service"
	CategoryUtility   Category = "utility"
)

// HealthStatus represents plugin health.
type HealthStatus struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
}

// Host provides plugins with access to NixGuard core services.
type Host interface {
	// Logger returns a logger scoped to the plugin.
	Logger() *slog.Logger

	// EventBus returns the event bus for publishing/subscribing.
	EventBus() EventPublisher

	// Config returns plugin-specific configuration.
	Config(key string) (string, bool)

	// RegisterRoute adds an HTTP route to the API.
	RegisterRoute(method, path string, handler interface{})
}

// EventPublisher is the subset of event.Bus that plugins can use.
type EventPublisher interface {
	Publish(ctx context.Context, evt interface{})
	Subscribe(eventType string, handler interface{})
}

// Registry manages plugin lifecycle.
type Registry struct {
	mu      sync.RWMutex
	log     *slog.Logger
	plugins map[string]Plugin
	order   []string // startup order
}

// NewRegistry creates a plugin registry.
func NewRegistry(log *slog.Logger) *Registry {
	return &Registry{
		log:     log,
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	meta := p.Metadata()
	if _, exists := r.plugins[meta.Name]; exists {
		return fmt.Errorf("plugin %q already registered", meta.Name)
	}

	r.plugins[meta.Name] = p
	r.order = append(r.order, meta.Name)
	r.log.Info("plugin registered",
		slog.String("name", meta.Name),
		slog.String("version", meta.Version),
	)
	return nil
}

// StartAll initializes and starts all registered plugins in order.
func (r *Registry) StartAll(ctx context.Context, host Host) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, name := range r.order {
		p := r.plugins[name]
		if err := p.Init(ctx, host); err != nil {
			return fmt.Errorf("plugin %q init: %w", name, err)
		}
		if err := p.Start(ctx); err != nil {
			return fmt.Errorf("plugin %q start: %w", name, err)
		}
		r.log.Info("plugin started", slog.String("name", name))
	}
	return nil
}

// StopAll stops all plugins in reverse order.
func (r *Registry) StopAll(ctx context.Context) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i := len(r.order) - 1; i >= 0; i-- {
		name := r.order[i]
		if err := r.plugins[name].Stop(ctx); err != nil {
			r.log.Error("plugin stop error",
				slog.String("name", name),
				slog.String("error", err.Error()),
			)
		}
	}
}

// List returns metadata for all registered plugins.
func (r *Registry) List() []Metadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Metadata, 0, len(r.plugins))
	for _, name := range r.order {
		result = append(result, r.plugins[name].Metadata())
	}
	return result
}
