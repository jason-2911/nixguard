package dns

import "context"

type ResolverConfigRepository interface {
	Get(ctx context.Context) (*ResolverConfig, error)
	Save(ctx context.Context, cfg *ResolverConfig) error
}

type HostOverrideRepository interface {
	List(ctx context.Context) ([]HostOverride, error)
	GetByID(ctx context.Context, id string) (*HostOverride, error)
	Create(ctx context.Context, override *HostOverride) error
	Update(ctx context.Context, override *HostOverride) error
	Delete(ctx context.Context, id string) error
}

type DomainOverrideRepository interface {
	List(ctx context.Context) ([]DomainOverride, error)
	GetByID(ctx context.Context, id string) (*DomainOverride, error)
	Create(ctx context.Context, override *DomainOverride) error
	Update(ctx context.Context, override *DomainOverride) error
	Delete(ctx context.Context, id string) error
}

type BlocklistRepository interface {
	List(ctx context.Context) ([]Blocklist, error)
	GetByID(ctx context.Context, id string) (*Blocklist, error)
	Create(ctx context.Context, bl *Blocklist) error
	Update(ctx context.Context, bl *Blocklist) error
	Delete(ctx context.Context, id string) error
}

// DNSEngine interfaces with Unbound resolver.
type DNSEngine interface {
	ApplyConfig(ctx context.Context, cfg ResolverConfig, overrides []HostOverride, domains []DomainOverride) error
	ApplyBlocklists(ctx context.Context, blocklists []Blocklist, whitelists []Whitelist) error
	FlushCache(ctx context.Context) error
	GetStats(ctx context.Context) (*ResolverStats, error)
	Restart(ctx context.Context) error
	QueryLog(ctx context.Context, filter QueryLogFilter) ([]DNSQueryLog, error)
}

type ResolverStats struct {
	TotalQueries    uint64  `json:"total_queries"`
	CacheHits       uint64  `json:"cache_hits"`
	CacheMisses     uint64  `json:"cache_misses"`
	CacheHitPercent float64 `json:"cache_hit_percent"`
	BlockedQueries  uint64  `json:"blocked_queries"`
	Uptime          string  `json:"uptime"`
}

type QueryLogFilter struct {
	ClientIP  string
	Domain    string
	Blocked   *bool
	StartTime string
	EndTime   string
	Limit     int
	Offset    int
}
