// Package loadbalancer contains the domain model for load balancing (HAProxy).
// Maps to OPNsense: Services > HAProxy.
package loadbalancer

import "time"

// Frontend defines how HAProxy receives connections.
type Frontend struct {
	ID           string   `json:"id" db:"id"`
	Name         string   `json:"name" db:"name"`
	Mode         string   `json:"mode" db:"mode"` // http, tcp
	BindAddress  string   `json:"bind_address" db:"bind_address"`
	BindPort     int      `json:"bind_port" db:"bind_port"`
	DefaultBackend string `json:"default_backend" db:"default_backend"`
	SSLCertID    string   `json:"ssl_cert_id,omitempty" db:"ssl_cert_id"`
	SSLOffload   bool     `json:"ssl_offload" db:"ssl_offload"`
	HTTP2        bool     `json:"http2" db:"http2"`
	ACLRules     []ACLRule `json:"acl_rules"`
	MaxConn      int      `json:"max_conn" db:"max_conn"`
	Enabled      bool     `json:"enabled" db:"enabled"`
	Description  string   `json:"description" db:"description"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Backend defines a pool of servers.
type Backend struct {
	ID              string         `json:"id" db:"id"`
	Name            string         `json:"name" db:"name"`
	Mode            string         `json:"mode" db:"mode"` // http, tcp
	Balance         string         `json:"balance" db:"balance"` // roundrobin, leastconn, source
	Servers         []BackendServer `json:"servers"`
	HealthCheck     HealthCheck    `json:"health_check"`
	StickySession   *StickySession `json:"sticky_session,omitempty"`
	Compression     bool           `json:"compression" db:"compression"`
	Enabled         bool           `json:"enabled" db:"enabled"`
	Description     string         `json:"description" db:"description"`
}

// BackendServer is a single server in a backend pool.
type BackendServer struct {
	ID          string `json:"id" db:"id"`
	BackendID   string `json:"backend_id" db:"backend_id"`
	Name        string `json:"name" db:"name"`
	Address     string `json:"address" db:"address"`
	Port        int    `json:"port" db:"port"`
	Weight      int    `json:"weight" db:"weight"`
	MaxConn     int    `json:"max_conn" db:"max_conn"`
	IsBackup    bool   `json:"is_backup" db:"is_backup"`
	SSLEnabled  bool   `json:"ssl_enabled" db:"ssl_enabled"`
	Maintenance bool   `json:"maintenance" db:"maintenance"`
	Status      string `json:"status"` // up, down, maint
}

type HealthCheck struct {
	Type     string `json:"type"`     // http, tcp, agent
	Interval int    `json:"interval"` // milliseconds
	Timeout  int    `json:"timeout"`
	Rise     int    `json:"rise"`     // checks before marking up
	Fall     int    `json:"fall"`     // checks before marking down
	URI      string `json:"uri,omitempty"`      // for http checks
	ExpectStatus int `json:"expect_status,omitempty"` // expected HTTP status
}

type StickySession struct {
	Type     string `json:"type"`      // cookie, source_ip
	CookieName string `json:"cookie_name,omitempty"`
	TTL      int    `json:"ttl"`       // seconds
}

// ACLRule for HAProxy frontend routing.
type ACLRule struct {
	Name      string `json:"name"`
	Match     string `json:"match"`     // host, path, header, src
	Pattern   string `json:"pattern"`
	Backend   string `json:"backend"`   // route to this backend
	Condition string `json:"condition"` // if, unless
}

// HAProxyStats holds runtime statistics.
type HAProxyStats struct {
	Frontends []FrontendStats `json:"frontends"`
	Backends  []BackendStats  `json:"backends"`
}

type FrontendStats struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	CurrentConn   int    `json:"current_conn"`
	MaxConn       int    `json:"max_conn"`
	TotalSessions uint64 `json:"total_sessions"`
	BytesIn       uint64 `json:"bytes_in"`
	BytesOut      uint64 `json:"bytes_out"`
	RequestRate   int    `json:"request_rate"`
	ErrorRate     int    `json:"error_rate"`
}

type BackendStats struct {
	Name    string              `json:"name"`
	Servers []BackendServerStats `json:"servers"`
}

type BackendServerStats struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Weight     int    `json:"weight"`
	CurrentConn int   `json:"current_conn"`
	TotalConn  uint64 `json:"total_conn"`
	BytesIn    uint64 `json:"bytes_in"`
	BytesOut   uint64 `json:"bytes_out"`
	ResponseTime int  `json:"response_time_ms"`
	HealthChecks int  `json:"health_checks_passed"`
}
