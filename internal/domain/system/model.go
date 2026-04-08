// Package system contains the domain model for system management.
// Maps to OPNsense: System > Settings, Firmware, Diagnostics, Backup.
package system

import "time"

// GeneralSettings holds global system configuration.
type GeneralSettings struct {
	Hostname     string   `json:"hostname" db:"hostname"`
	Domain       string   `json:"domain" db:"domain"`
	DNSServers   []string `json:"dns_servers"`
	Timezone     string   `json:"timezone" db:"timezone"`
	NTPServers   []string `json:"ntp_servers"`
	Language     string   `json:"language" db:"language"`
}

// Backup represents a configuration backup.
type Backup struct {
	ID          string    `json:"id" db:"id"`
	Filename    string    `json:"filename" db:"filename"`
	Size        int64     `json:"size" db:"size"`
	Type        string    `json:"type" db:"backup_type"` // manual, auto, pre_update
	Encrypted   bool      `json:"encrypted" db:"encrypted"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ConfigDiff is a change between two configurations.
type ConfigDiff struct {
	Section  string `json:"section"`
	Key      string `json:"key"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
	Action   string `json:"action"` // added, removed, changed
}

// UpdateInfo describes an available system update.
type UpdateInfo struct {
	CurrentVersion string    `json:"current_version"`
	LatestVersion  string    `json:"latest_version"`
	Channel        string    `json:"channel"` // stable, testing
	ReleaseDate    time.Time `json:"release_date"`
	Changelog      string    `json:"changelog"`
	UpdateAvailable bool     `json:"update_available"`
}

// TunableParam is a sysctl kernel parameter.
type TunableParam struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`       // e.g., net.ipv4.ip_forward
	Value       string `json:"value" db:"value"`
	Default     string `json:"default" db:"default_value"`
	Description string `json:"description" db:"description"`
}

// DiagnosticResult is the result of a diagnostic tool.
type DiagnosticResult struct {
	Tool      string    `json:"tool"` // ping, traceroute, dns_lookup, etc.
	Target    string    `json:"target"`
	Output    string    `json:"output"`
	ExitCode  int       `json:"exit_code"`
	Duration  float64   `json:"duration_ms"`
	Timestamp time.Time `json:"timestamp"`
}

// DDNSEntry is a Dynamic DNS update configuration.
type DDNSEntry struct {
	ID          string `json:"id" db:"id"`
	Provider    string `json:"provider" db:"provider"` // cloudflare, noip, etc.
	Hostname    string `json:"hostname" db:"hostname"`
	Username    string `json:"username" db:"username"`
	Password    string `json:"-" db:"password"`
	Interface   string `json:"interface" db:"interface_name"`
	UseIPv6     bool   `json:"use_ipv6" db:"use_ipv6"`
	ForceUpdate int    `json:"force_update" db:"force_update"` // days
	Enabled     bool   `json:"enabled" db:"enabled"`
	LastUpdate  *time.Time `json:"last_update" db:"last_update"`
	LastIP      string `json:"last_ip" db:"last_ip"`
}

// CronJob is a scheduled task.
type CronJob struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Command     string `json:"command" db:"command"`
	Schedule    string `json:"schedule" db:"schedule"` // cron expression
	Enabled     bool   `json:"enabled" db:"enabled"`
	Description string `json:"description" db:"description"`
}
