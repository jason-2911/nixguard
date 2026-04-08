package monitor

import "context"

// MetricsCollector gathers system metrics.
type MetricsCollector interface {
	CollectSystem(ctx context.Context) (*SystemMetrics, error)
	CollectTraffic(ctx context.Context, iface string) (*TrafficSample, error)
	CollectTopTalkers(ctx context.Context, limit int) ([]TopTalker, error)
	GetServiceStatus(ctx context.Context, name string) (*ServiceStatus, error)
	ListServices(ctx context.Context) ([]ServiceStatus, error)
}

// MetricsStore persists time-series metrics.
type MetricsStore interface {
	WriteTraffic(ctx context.Context, sample TrafficSample) error
	WriteSystem(ctx context.Context, metrics SystemMetrics) error
	QueryTraffic(ctx context.Context, iface string, from, to string, interval string) ([]TrafficSample, error)
	QuerySystem(ctx context.Context, from, to string, interval string) ([]SystemMetrics, error)
}

// LogReader reads system and service logs.
type LogReader interface {
	ReadSystemLog(ctx context.Context, filter LogFilter) ([]LogEntry, error)
	ReadFirewallLog(ctx context.Context, filter LogFilter) ([]FirewallLogEntry, error)
	ReadServiceLog(ctx context.Context, service string, filter LogFilter) ([]LogEntry, error)
}

// SyslogTargetRepository persists syslog forwarding configs.
type SyslogTargetRepository interface {
	List(ctx context.Context) ([]SyslogTarget, error)
	GetByID(ctx context.Context, id string) (*SyslogTarget, error)
	Create(ctx context.Context, target *SyslogTarget) error
	Update(ctx context.Context, target *SyslogTarget) error
	Delete(ctx context.Context, id string) error
}

// AlertRuleRepository persists alert configurations.
type AlertRuleRepository interface {
	List(ctx context.Context) ([]AlertRule, error)
	GetByID(ctx context.Context, id string) (*AlertRule, error)
	Create(ctx context.Context, rule *AlertRule) error
	Update(ctx context.Context, rule *AlertRule) error
	Delete(ctx context.Context, id string) error
}

// NotificationSender delivers alerts to configured channels.
type NotificationSender interface {
	Send(ctx context.Context, channel NotificationChannel, subject, body string) error
	TestChannel(ctx context.Context, channel NotificationChannel) error
}

// PacketCapture interfaces with tcpdump for live capture.
type PacketCapture interface {
	Start(ctx context.Context, opts CaptureOptions) (captureID string, err error)
	Stop(ctx context.Context, captureID string) error
	GetStatus(ctx context.Context, captureID string) (*CaptureStatus, error)
	Download(ctx context.Context, captureID string) ([]byte, error)
}

type CaptureOptions struct {
	Interface  string `json:"interface"`
	Filter     string `json:"filter"`      // BPF filter expression
	MaxPackets int    `json:"max_packets"`
	Duration   int    `json:"duration_sec"`
	SnapLen    int    `json:"snap_len"`
}

type CaptureStatus struct {
	ID        string `json:"id"`
	State     string `json:"state"` // running, stopped, completed
	Packets   int    `json:"packets"`
	FileSize  int64  `json:"file_size"`
	Interface string `json:"interface"`
	StartedAt string `json:"started_at"`
}
