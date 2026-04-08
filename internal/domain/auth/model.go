// Package auth contains the domain model for user management and authentication.
// Maps to OPNsense: System > Access > Users, Groups, Servers.
package auth

import "time"

// User represents a NixGuard user account.
type User struct {
	ID              string    `json:"id" db:"id"`
	Username        string    `json:"username" db:"username"`
	PasswordHash    string    `json:"-" db:"password_hash"`
	FullName        string    `json:"full_name" db:"full_name"`
	Email           string    `json:"email" db:"email"`
	Groups          []string  `json:"groups"`
	Enabled         bool      `json:"enabled" db:"enabled"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	MFAEnabled      bool      `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret       string    `json:"-" db:"mfa_secret"`
	APIKey          string    `json:"-" db:"api_key"`
	Shell           string    `json:"shell" db:"shell"`
	AuthSource      string    `json:"auth_source" db:"auth_source"` // local, ldap, radius
	LastLogin       *time.Time `json:"last_login,omitempty" db:"last_login"`
	LoginAttempts   int       `json:"-" db:"login_attempts"`
	LockedUntil     *time.Time `json:"-" db:"locked_until"`
	Description     string    `json:"description" db:"description"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Group represents a user group with permissions.
type Group struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Privileges  []string  `json:"privileges"`
	Members     []string  `json:"members"` // user IDs
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Privilege represents a granular permission.
type Privilege struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"` // firewall, vpn, system, etc.
}

// Session represents an active user session.
type Session struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"-" db:"token"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

// AuditLog records user actions for compliance.
type AuditLog struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	Action    string    `json:"action" db:"action"`
	Resource  string    `json:"resource" db:"resource"`
	Detail    string    `json:"detail" db:"detail"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// LDAPServer is an external LDAP authentication source.
type LDAPServer struct {
	ID         string `json:"id" db:"id"`
	Name       string `json:"name" db:"name"`
	Host       string `json:"host" db:"host"`
	Port       int    `json:"port" db:"port"`
	UseTLS     bool   `json:"use_tls" db:"use_tls"`
	BaseDN     string `json:"base_dn" db:"base_dn"`
	BindDN     string `json:"bind_dn" db:"bind_dn"`
	BindPassword string `json:"-" db:"bind_password"`
	UserFilter string `json:"user_filter" db:"user_filter"`
	GroupFilter string `json:"group_filter" db:"group_filter"`
	Enabled    bool   `json:"enabled" db:"enabled"`
}

// RADIUSServer is an external RADIUS authentication source.
type RADIUSServer struct {
	ID           string `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	Host         string `json:"host" db:"host"`
	Port         int    `json:"port" db:"port"`
	Secret       string `json:"-" db:"secret"`
	AuthProtocol string `json:"auth_protocol" db:"auth_protocol"` // PAP, CHAP, MSCHAPv2
	Timeout      int    `json:"timeout" db:"timeout"`
	Enabled      bool   `json:"enabled" db:"enabled"`
}

// Certificate represents a TLS/SSL certificate.
type Certificate struct {
	ID            string    `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Type          string    `json:"type" db:"cert_type"` // ca, server, client
	CommonName    string    `json:"common_name" db:"common_name"`
	SANs          []string  `json:"sans"`
	Issuer        string    `json:"issuer" db:"issuer"`
	SerialNumber  string    `json:"serial_number" db:"serial_number"`
	KeyType       string    `json:"key_type" db:"key_type"` // RSA, ECDSA
	KeyLength     int       `json:"key_length" db:"key_length"`
	CertPEM       string    `json:"-" db:"cert_pem"`
	KeyPEM        string    `json:"-" db:"key_pem"`
	CAPEM         string    `json:"-" db:"ca_pem"`
	NotBefore     time.Time `json:"not_before" db:"not_before"`
	NotAfter      time.Time `json:"not_after" db:"not_after"`
	IsCA          bool      `json:"is_ca" db:"is_ca"`
	ParentCAID    string    `json:"parent_ca_id,omitempty" db:"parent_ca_id"`
	ACMEEnabled   bool      `json:"acme_enabled" db:"acme_enabled"`
	ACMEProvider  string    `json:"acme_provider,omitempty" db:"acme_provider"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
