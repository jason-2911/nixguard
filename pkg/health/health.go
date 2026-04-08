// Package health provides health check infrastructure.
package health

import (
	"context"
	"sync"
	"time"
)

// Status represents a component's health.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Check is a health check function.
type Check func(ctx context.Context) ComponentHealth

// ComponentHealth is the health of a single component.
type ComponentHealth struct {
	Name    string        `json:"name"`
	Status  Status        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency_ms"`
}

// Registry holds all health checks.
type Registry struct {
	mu     sync.RWMutex
	checks map[string]Check
}

// NewRegistry creates a health check registry.
func NewRegistry() *Registry {
	return &Registry{checks: make(map[string]Check)}
}

// Register adds a named health check.
func (r *Registry) Register(name string, check Check) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks[name] = check
}

// RunAll executes all checks and returns aggregate health.
func (r *Registry) RunAll(ctx context.Context) SystemHealth {
	r.mu.RLock()
	defer r.mu.RUnlock()

	health := SystemHealth{
		Status:     StatusHealthy,
		Components: make([]ComponentHealth, 0, len(r.checks)),
		Timestamp:  time.Now(),
	}

	for name, check := range r.checks {
		start := time.Now()
		ch := check(ctx)
		ch.Name = name
		ch.Latency = time.Since(start)
		health.Components = append(health.Components, ch)

		if ch.Status == StatusUnhealthy {
			health.Status = StatusUnhealthy
		} else if ch.Status == StatusDegraded && health.Status == StatusHealthy {
			health.Status = StatusDegraded
		}
	}

	return health
}

// SystemHealth is the overall system health.
type SystemHealth struct {
	Status     Status            `json:"status"`
	Components []ComponentHealth `json:"components"`
	Timestamp  time.Time         `json:"timestamp"`
}
