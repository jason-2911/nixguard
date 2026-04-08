// Package captiveportal contains the domain model for the Captive Portal.
// Maps to OPNsense: Services > Captive Portal.
package captiveportal

import "time"

// Zone is a captive portal zone (one per interface/VLAN).
type Zone struct {
	ID             string `json:"id" db:"id"`
	Name           string `json:"name" db:"name"`
	Interface      string `json:"interface" db:"interface_name"`
	AuthMethod     string `json:"auth_method" db:"auth_method"` // local, radius, ldap, voucher, none
	IdleTimeout    int    `json:"idle_timeout" db:"idle_timeout"`       // minutes
	HardTimeout    int    `json:"hard_timeout" db:"hard_timeout"`       // minutes
	MaxConcurrent  int    `json:"max_concurrent" db:"max_concurrent"`
	BWUpload       int    `json:"bw_upload" db:"bw_upload"`     // Kbit/s per user
	BWDownload     int    `json:"bw_download" db:"bw_download"` // Kbit/s per user
	TrafficQuota   int    `json:"traffic_quota_mb" db:"traffic_quota_mb"`
	AllowedMACs    []string `json:"allowed_macs"`    // bypass list
	AllowedIPs     []string `json:"allowed_ips"`     // bypass list
	AllowedDomains []string `json:"allowed_domains"` // bypass list
	TemplateID     string `json:"template_id" db:"template_id"`
	CertID         string `json:"cert_id" db:"cert_id"`
	Enabled        bool   `json:"enabled" db:"enabled"`
	Description    string `json:"description" db:"description"`
}

// PortalTemplate customizes the captive portal login page.
type PortalTemplate struct {
	ID           string `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	HTML         string `json:"html" db:"html"`
	CSS          string `json:"css" db:"css"`
	LogoURL      string `json:"logo_url" db:"logo_url"`
	BackgroundURL string `json:"background_url" db:"background_url"`
	TermsOfService string `json:"terms_of_service" db:"terms_of_service"`
}

// Voucher is a single-use or time-limited access code.
type Voucher struct {
	ID          string     `json:"id" db:"id"`
	Code        string     `json:"code" db:"code"`
	ZoneID      string     `json:"zone_id" db:"zone_id"`
	ValidMinutes int       `json:"valid_minutes" db:"valid_minutes"`
	BWUpload    int        `json:"bw_upload" db:"bw_upload"`
	BWDownload  int        `json:"bw_download" db:"bw_download"`
	MaxUses     int        `json:"max_uses" db:"max_uses"`
	UsesCount   int        `json:"uses_count" db:"uses_count"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Session is an active captive portal user session.
type PortalSession struct {
	ID         string    `json:"id" db:"id"`
	ZoneID     string    `json:"zone_id" db:"zone_id"`
	Username   string    `json:"username" db:"username"`
	MACAddress string    `json:"mac_address" db:"mac_address"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	BytesIn    uint64    `json:"bytes_in" db:"bytes_in"`
	BytesOut   uint64    `json:"bytes_out" db:"bytes_out"`
	StartedAt  time.Time `json:"started_at" db:"started_at"`
	LastSeen   time.Time `json:"last_seen" db:"last_seen"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
}
