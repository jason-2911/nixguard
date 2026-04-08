// Package monitor contains the domain model for monitoring, logging, and alerting.
// Maps to OPNsense: Reporting, Diagnostics, System > Log Files.
package monitor

import "time"

// ─── Traffic Monitoring ────────────────────────────────────────

type TrafficSample struct {
	Timestamp  time.Time `json:"timestamp"`
	Interface  string    `json:"interface"`
	RxBytesPS  uint64    `json:"rx_bytes_per_sec"`
	TxBytesPS  uint64    `json:"tx_bytes_per_sec"`
	RxPacketsPS uint64   `json:"rx_packets_per_sec"`
	TxPacketsPS uint64   `json:"tx_packets_per_sec"`
}

type TopTalker struct {
	IPAddress string `json:"ip_address"`
	Hostname  string `json:"hostname,omitempty"`
	BytesIn   uint64 `json:"bytes_in"`
	BytesOut  uint64 `json:"bytes_out"`
	Connections int  `json:"connections"`
}

// ─── System Metrics ────────────────────────────────────────────

type SystemMetrics struct {
	Timestamp   time.Time      `json:"timestamp"`
	CPU         CPUMetrics     `json:"cpu"`
	Memory      MemoryMetrics  `json:"memory"`
	Disk        []DiskMetrics  `json:"disk"`
	Temperature []TempSensor   `json:"temperature"`
	LoadAvg     [3]float64     `json:"load_avg"`
	Uptime      string         `json:"uptime"`
}

type CPUMetrics struct {
	UsagePercent float64   `json:"usage_percent"`
	PerCore      []float64 `json:"per_core"`
	IOWait       float64   `json:"iowait"`
	System       float64   `json:"system"`
	User         float64   `json:"user"`
}

type MemoryMetrics struct {
	TotalMB   uint64  `json:"total_mb"`
	UsedMB    uint64  `json:"used_mb"`
	FreeMB    uint64  `json:"free_mb"`
	CachedMB  uint64  `json:"cached_mb"`
	SwapTotal uint64  `json:"swap_total_mb"`
	SwapUsed  uint64  `json:"swap_used_mb"`
	UsagePercent float64 `json:"usage_percent"`
}

type DiskMetrics struct {
	Filesystem   string  `json:"filesystem"`
	MountPoint   string  `json:"mount_point"`
	TotalGB      float64 `json:"total_gb"`
	UsedGB       float64 `json:"used_gb"`
	AvailGB      float64 `json:"avail_gb"`
	UsagePercent float64 `json:"usage_percent"`
}

type TempSensor struct {
	Name        string  `json:"name"`
	Temperature float64 `json:"temperature_c"`
	Critical    float64 `json:"critical_c"`
}

// ─── Service Status ────────────────────────────────────────────

type ServiceStatus struct {
	Name       string `json:"name"`
	Status     string `json:"status"` // running, stopped, failed
	Enabled    bool   `json:"enabled"`
	PID        int    `json:"pid,omitempty"`
	Memory     uint64 `json:"memory_bytes,omitempty"`
	CPU        float64 `json:"cpu_percent,omitempty"`
	Uptime     string `json:"uptime,omitempty"`
}

// ─── Logging ───────────────────────────────────────────────────

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Facility  string    `json:"facility"`
	Severity  string    `json:"severity"`
	Process   string    `json:"process"`
	PID       int       `json:"pid,omitempty"`
	Message   string    `json:"message"`
}

type LogFilter struct {
	Facility  string
	Severity  string
	Process   string
	Search    string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Offset    int
}

// FirewallLogEntry is a parsed firewall log line.
type FirewallLogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Interface   string    `json:"interface"`
	RuleID      string    `json:"rule_id"`
	Action      string    `json:"action"`
	Direction   string    `json:"direction"`
	Protocol    string    `json:"protocol"`
	SourceIP    string    `json:"source_ip"`
	SourcePort  int       `json:"source_port"`
	DestIP      string    `json:"dest_ip"`
	DestPort    int       `json:"dest_port"`
	Length      int       `json:"length"`
}

// SyslogTarget is a remote syslog destination.
type SyslogTarget struct {
	ID        string `json:"id" db:"id"`
	Host      string `json:"host" db:"host"`
	Port      int    `json:"port" db:"port"`
	Protocol  string `json:"protocol" db:"protocol"` // udp, tcp, tls
	Format    string `json:"format" db:"format"`     // rfc3164, rfc5424
	Facilities []string `json:"facilities"`
	Enabled   bool   `json:"enabled" db:"enabled"`
}

// ─── Alerts & Notifications ────────────────────────────────────

type AlertRule struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Type        string `json:"type" db:"alert_type"` // threshold, change, absence
	Metric      string `json:"metric" db:"metric"`
	Condition   string `json:"condition" db:"condition"` // gt, lt, eq
	Value       float64 `json:"value" db:"threshold_value"`
	Duration    string `json:"duration" db:"duration"`
	Channels    []string `json:"channels"` // email, webhook, slack
	Enabled     bool   `json:"enabled" db:"enabled"`
	Description string `json:"description" db:"description"`
}

type NotificationChannel struct {
	ID     string                 `json:"id" db:"id"`
	Name   string                 `json:"name" db:"name"`
	Type   string                 `json:"type" db:"channel_type"` // email, webhook, slack, telegram
	Config map[string]string      `json:"config"`
	Enabled bool                  `json:"enabled" db:"enabled"`
}
