// Package proxy contains the domain model for proxy and web filtering.
// Maps to OPNsense: Services > Squid, SquidGuard, Web Application Firewall.
package proxy

import "time"

// ProxyConfig is the main Squid proxy configuration.
type ProxyConfig struct {
	ID                string   `json:"id" db:"id"`
	Enabled           bool     `json:"enabled" db:"enabled"`
	Mode              string   `json:"mode" db:"mode"` // forward, transparent, reverse
	ListenPort        int      `json:"listen_port" db:"listen_port"`
	SSLBump           bool     `json:"ssl_bump" db:"ssl_bump"`
	CacheEnabled      bool     `json:"cache_enabled" db:"cache_enabled"`
	CacheSizeMB       int      `json:"cache_size_mb" db:"cache_size_mb"`
	MemCacheMB        int      `json:"mem_cache_mb" db:"mem_cache_mb"`
	MaxObjectSize     int      `json:"max_object_size_kb" db:"max_object_size_kb"`
	VisibleHostname   string   `json:"visible_hostname" db:"visible_hostname"`
	AllowedSubnets    []string `json:"allowed_subnets"`
	AuthEnabled       bool     `json:"auth_enabled" db:"auth_enabled"`
	AuthMethod        string   `json:"auth_method" db:"auth_method"` // local, ldap, radius
	LoggingEnabled    bool     `json:"logging_enabled" db:"logging_enabled"`
}

// ACLRule is a proxy access control rule.
type ACLRule struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Type        string `json:"type" db:"acl_type"` // src_ip, dst_domain, url_regex, time
	Value       string `json:"value" db:"value"`
	Action      string `json:"action" db:"action"` // allow, deny
	Order       int    `json:"order" db:"rule_order"`
	Enabled     bool   `json:"enabled" db:"enabled"`
	Description string `json:"description" db:"description"`
}

// URLFilter is a category-based URL filtering rule (SquidGuard).
type URLFilter struct {
	ID          string   `json:"id" db:"id"`
	Name        string   `json:"name" db:"name"`
	Categories  []string `json:"categories"`
	Action      string   `json:"action" db:"action"` // allow, block, redirect
	RedirectURL string   `json:"redirect_url" db:"redirect_url"`
	Users       []string `json:"users"`
	Groups      []string `json:"groups"`
	TimeRule    string   `json:"time_rule" db:"time_rule"`
	Enabled     bool     `json:"enabled" db:"enabled"`
}

// URLCategory is a set of domain lists for filtering.
type URLCategory struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Source      string    `json:"source" db:"source"` // builtin, custom, url
	URL         string    `json:"url,omitempty" db:"url"`
	DomainCount int       `json:"domain_count" db:"domain_count"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
	Enabled     bool      `json:"enabled" db:"enabled"`
}
