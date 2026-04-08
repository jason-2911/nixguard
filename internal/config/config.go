// Package config provides application configuration loading and validation.
// Supports YAML files, environment variables, and CLI flags.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config is the root configuration for nixguard-server.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Auth     AuthConfig     `yaml:"auth"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
	Agent    AgentConfig    `yaml:"agent"`
	Modules  ModulesConfig  `yaml:"modules"`
	GeoIP    GeoIPConfig    `yaml:"geoip"`
}

type GeoIPConfig struct {
	LicenseKey     string `yaml:"license_key"`
	UpdateInterval string `yaml:"update_interval"` // e.g. "24h"
	DataDir        string `yaml:"data_dir"`         // e.g. "./data/geoip"
}

type ServerConfig struct {
	ListenAddr  string   `yaml:"listen_addr"`
	TLSCert     string   `yaml:"tls_cert"`
	TLSKey      string   `yaml:"tls_key"`
	CORSOrigins []string `yaml:"cors_origins"`
	RateLimit   int      `yaml:"rate_limit"`
}

type AuthConfig struct {
	JWTSecret       string        `yaml:"jwt_secret"`
	TokenExpiry     time.Duration `yaml:"token_expiry"`
	SessionTimeout  time.Duration `yaml:"session_timeout"`
	MaxLoginAttempts int          `yaml:"max_login_attempts"`
	LockoutDuration time.Duration `yaml:"lockout_duration"`
	MFAEnabled      bool          `yaml:"mfa_enabled"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"` // sqlite, postgres
	DSN    string `yaml:"dsn"`
}

type LogConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, text
}

type AgentConfig struct {
	SocketPath      string   `yaml:"socket_path"`
	AllowedCommands []string `yaml:"allowed_commands"`
}

// ModulesConfig controls which modules are enabled.
type ModulesConfig struct {
	Firewall       ModuleToggle `yaml:"firewall"`
	Network        ModuleToggle `yaml:"network"`
	VPN            ModuleToggle `yaml:"vpn"`
	DNS            ModuleToggle `yaml:"dns"`
	DHCP           ModuleToggle `yaml:"dhcp"`
	IDS            ModuleToggle `yaml:"ids"`
	Proxy          ModuleToggle `yaml:"proxy"`
	LoadBalancer   ModuleToggle `yaml:"loadbalancer"`
	HA             ModuleToggle `yaml:"ha"`
	TrafficShaper  ModuleToggle `yaml:"traffic_shaper"`
	CaptivePortal  ModuleToggle `yaml:"captive_portal"`
	Monitor        ModuleToggle `yaml:"monitor"`
}

type ModuleToggle struct {
	Enabled bool `yaml:"enabled"`
}

// Load reads configuration from the specified YAML file.
// It also merges environment variable overrides (NIXGUARD_ prefix).
func Load(path string) (*Config, error) {
	cfg := defaultConfig()

	if _, err := os.Stat(path); err == nil {
		// TODO: parse YAML with viper
		_ = path
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

// LoadAgent reads the agent-specific configuration.
func LoadAgent(path string) (*Config, error) {
	return Load(path)
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			ListenAddr:  "0.0.0.0:8443",
			CORSOrigins: []string{"*"},
			RateLimit:   100,
		},
		Auth: AuthConfig{
			TokenExpiry:      24 * time.Hour,
			SessionTimeout:   30 * time.Minute,
			MaxLoginAttempts: 5,
			LockoutDuration:  15 * time.Minute,
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "./data/nixguard.db",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		Agent: AgentConfig{
			SocketPath: "/var/run/nixguard/agent.sock",
			AllowedCommands: []string{
				"nft", "iptables", "ip6tables",
				"ip", "ss", "tc", "bridge",
				"systemctl", "journalctl",
				"tcpdump", "conntrack",
				"sysctl", "ethtool",
				"wg", "strongswan",
			},
		},
		Modules: ModulesConfig{
			Firewall:      ModuleToggle{Enabled: true},
			Network:       ModuleToggle{Enabled: true},
			VPN:           ModuleToggle{Enabled: true},
			DNS:           ModuleToggle{Enabled: true},
			DHCP:          ModuleToggle{Enabled: true},
			IDS:           ModuleToggle{Enabled: false},
			Proxy:         ModuleToggle{Enabled: false},
			LoadBalancer:  ModuleToggle{Enabled: false},
			HA:            ModuleToggle{Enabled: false},
			TrafficShaper: ModuleToggle{Enabled: false},
			CaptivePortal: ModuleToggle{Enabled: false},
			Monitor:       ModuleToggle{Enabled: true},
		},
	}
}

// Validate checks configuration invariants.
func (c *Config) Validate() error {
	if c.Server.ListenAddr == "" {
		return fmt.Errorf("server.listen_addr is required")
	}
	if c.Database.Driver != "sqlite" && c.Database.Driver != "postgres" {
		return fmt.Errorf("database.driver must be sqlite or postgres")
	}
	return nil
}
