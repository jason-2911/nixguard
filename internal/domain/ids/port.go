package ids

import "context"

type ConfigRepository interface {
	Get(ctx context.Context) (*Config, error)
	Save(ctx context.Context, cfg *Config) error
}

type RulesetRepository interface {
	List(ctx context.Context) ([]Ruleset, error)
	GetByID(ctx context.Context, id string) (*Ruleset, error)
	Create(ctx context.Context, rs *Ruleset) error
	Update(ctx context.Context, rs *Ruleset) error
	Delete(ctx context.Context, id string) error
}

type RuleOverrideRepository interface {
	List(ctx context.Context) ([]RuleOverride, error)
	Save(ctx context.Context, override RuleOverride) error
	Delete(ctx context.Context, sid int) error
}

// IDSEngine interfaces with Suricata.
type IDSEngine interface {
	ApplyConfig(ctx context.Context, cfg Config) error
	UpdateRules(ctx context.Context) error
	GetAlerts(ctx context.Context, filter AlertFilter) ([]Alert, error)
	GetStats(ctx context.Context) (*Stats, error)
	Restart(ctx context.Context) error
	Reload(ctx context.Context) error
}

type Stats struct {
	Uptime        string  `json:"uptime"`
	PacketsCaptured uint64 `json:"packets_captured"`
	PacketsDropped  uint64 `json:"packets_dropped"`
	AlertsTotal   uint64  `json:"alerts_total"`
	AlertsToday   uint64  `json:"alerts_today"`
	FlowsActive   uint64  `json:"flows_active"`
}
