package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the complete homelab.yml schema.
type Config struct {
	Cluster    ClusterConfig          `yaml:"cluster"`
	Hosts      map[string]HostConfig  `yaml:"hosts"`
	VPN        VPNConfig              `yaml:"vpn"`
	Storage    StorageConfig          `yaml:"storage"`
	Apps       map[string]AppConfig   `yaml:"apps"`
	AutoUpdate AutoUpdateConfig       `yaml:"auto_update"`
	Secrets    SecretsConfig          `yaml:"secrets"`
}

type ClusterConfig struct {
	Name string `yaml:"name"`
}

type HostConfig struct {
	IP                 string `yaml:"ip"`
	NATSubnet          string `yaml:"nat_subnet"`
	WireGuardEndpoint  bool   `yaml:"wireguard_endpoint"`
}

type VPNConfig struct {
	Subnet    string   `yaml:"subnet"`
	Endpoint  string   `yaml:"endpoint"`
	Port      int      `yaml:"port"`
	LANRoutes []string `yaml:"lan_routes"`
}

type StorageConfig struct {
	MediaPath     string `yaml:"media_path"`
	DownloadsPath string `yaml:"downloads_path"`
	ConfigPath    string `yaml:"config_path"`
}

type AppConfig struct {
	Enabled bool   `yaml:"enabled"`
	Image   string `yaml:"image"`
	Port    int    `yaml:"port"`
}

type AutoUpdateConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Schedule string `yaml:"schedule"`
}

type SecretsConfig struct {
	SSHKeyPath string `yaml:"ssh_key_path"`
	SSHUser    string `yaml:"ssh_user"`
}

// Load reads and parses a homelab.yml file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	// Apply defaults
	cfg.applyDefaults()

	return &cfg, nil
}

// applyDefaults sets default values for optional fields.
func (c *Config) applyDefaults() {
	if c.VPN.Port == 0 {
		c.VPN.Port = 51820
	}
	if c.AutoUpdate.Schedule == "" {
		c.AutoUpdate.Schedule = "0 3 * * *"
	}
	if c.Secrets.SSHUser == "" {
		c.Secrets.SSHUser = "admin"
	}
	if c.Secrets.SSHKeyPath == "" {
		c.Secrets.SSHKeyPath = "~/.ssh/id_ed25519"
	}

	// Default app images
	for name, app := range c.Apps {
		if app.Image == "" {
			app.Image = fmt.Sprintf("linuxserver/%s:latest", name)
			c.Apps[name] = app
		}
	}
}

// WireGuardHost returns the host name designated as the WireGuard endpoint.
// Returns empty string if none is designated.
func (c *Config) WireGuardHost() string {
	for name, host := range c.Hosts {
		if host.WireGuardEndpoint {
			return name
		}
	}
	return ""
}

// EnabledApps returns a map of only the enabled applications.
func (c *Config) EnabledApps() map[string]AppConfig {
	enabled := make(map[string]AppConfig)
	for name, app := range c.Apps {
		if app.Enabled {
			enabled[name] = app
		}
	}
	return enabled
}
