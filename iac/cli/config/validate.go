package config

import (
	"fmt"
	"net"
	"strings"
)

// Validate checks the config for correctness and returns all errors found.
func (c *Config) Validate() []error {
	var errs []error

	// Cluster validation
	if c.Cluster.Name == "" {
		errs = append(errs, fmt.Errorf("cluster.name is required"))
	}

	// Hosts validation
	if len(c.Hosts) == 0 {
		errs = append(errs, fmt.Errorf("at least one host must be defined"))
	}

	seenIPs := make(map[string]string)
	seenSubnets := make(map[string]string)

	for name, host := range c.Hosts {
		if host.IP == "" {
			errs = append(errs, fmt.Errorf("host %q: ip is required", name))
		} else if ip := net.ParseIP(host.IP); ip == nil {
			errs = append(errs, fmt.Errorf("host %q: invalid IP address %q", name, host.IP))
		} else if prev, ok := seenIPs[host.IP]; ok {
			errs = append(errs, fmt.Errorf("host %q: duplicate IP %q (also used by %q)", name, host.IP, prev))
		} else {
			seenIPs[host.IP] = name
		}

		if host.NATSubnet == "" {
			errs = append(errs, fmt.Errorf("host %q: nat_subnet is required", name))
		} else if _, _, err := net.ParseCIDR(host.NATSubnet); err != nil {
			errs = append(errs, fmt.Errorf("host %q: invalid nat_subnet %q: %w", name, host.NATSubnet, err))
		} else if prev, ok := seenSubnets[host.NATSubnet]; ok {
			errs = append(errs, fmt.Errorf("host %q: duplicate nat_subnet %q (also used by %q)", name, host.NATSubnet, prev))
		} else {
			seenSubnets[host.NATSubnet] = name
		}
	}

	// VPN validation
	if c.VPN.Subnet != "" {
		if _, _, err := net.ParseCIDR(c.VPN.Subnet); err != nil {
			errs = append(errs, fmt.Errorf("vpn.subnet: invalid CIDR %q: %w", c.VPN.Subnet, err))
		}
	}

	if c.VPN.Port < 0 || c.VPN.Port > 65535 {
		errs = append(errs, fmt.Errorf("vpn.port: must be between 0 and 65535, got %d", c.VPN.Port))
	}

	// WireGuard endpoint check
	wgEndpoints := 0
	for _, host := range c.Hosts {
		if host.WireGuardEndpoint {
			wgEndpoints++
		}
	}
	if wgEndpoints > 1 {
		errs = append(errs, fmt.Errorf("only one host can have wireguard_endpoint: true (found %d)", wgEndpoints))
	}

	// Storage validation
	if c.Storage.MediaPath == "" {
		errs = append(errs, fmt.Errorf("storage.media_path is required"))
	}
	if c.Storage.DownloadsPath == "" {
		errs = append(errs, fmt.Errorf("storage.downloads_path is required"))
	}
	if c.Storage.ConfigPath == "" {
		errs = append(errs, fmt.Errorf("storage.config_path is required"))
	}

	// Apps validation
	seenPorts := make(map[int]string)
	for name, app := range c.Apps {
		if app.Port < 1 || app.Port > 65535 {
			errs = append(errs, fmt.Errorf("app %q: port must be between 1 and 65535, got %d", name, app.Port))
		} else if prev, ok := seenPorts[app.Port]; ok {
			errs = append(errs, fmt.Errorf("app %q: duplicate port %d (also used by %q)", name, app.Port, prev))
		} else {
			seenPorts[app.Port] = name
		}
	}

	// Auto-update validation
	if c.AutoUpdate.HealthCheckTimeout != "" {
		timeout := c.AutoUpdate.HealthCheckTimeout
		if !strings.HasSuffix(timeout, "s") && !strings.HasSuffix(timeout, "m") && !strings.HasSuffix(timeout, "h") {
			errs = append(errs, fmt.Errorf("auto_update.health_check_timeout: must end with s, m, or h (got %q)", timeout))
		}
	}

	return errs
}

// ValidateOrError validates the config and returns a single combined error, or nil.
func (c *Config) ValidateOrError() error {
	errs := c.Validate()
	if len(errs) == 0 {
		return nil
	}

	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = fmt.Sprintf("  - %s", e.Error())
	}
	return fmt.Errorf("config validation failed:\n%s", strings.Join(msgs, "\n"))
}
