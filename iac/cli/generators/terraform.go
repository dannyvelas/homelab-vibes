package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/homelab-vibe/iac/config"
)

// GenerateTerraformVars creates a terraform.tfvars file in HCL format.
func GenerateTerraformVars(cfg *config.Config, outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", outputDir, err)
	}

	var sb strings.Builder

	// Get first host for libvirt_uri and nomad_address
	firstHostIP := getFirstHostIP(cfg)

	// Cluster configuration
	sb.WriteString(fmt.Sprintf("cluster_name = %q\n", cfg.Cluster.Name))
	sb.WriteString(fmt.Sprintf("datacenter = %q\n", cfg.Cluster.Datacenter))
	sb.WriteString("\n")

	// Libvirt and Nomad connection
	sb.WriteString(fmt.Sprintf("libvirt_uri = \"qemu+ssh://%s@%s/system\"\n", cfg.Secrets.SSHUser, firstHostIP))
	sb.WriteString(fmt.Sprintf("nomad_address = \"http://%s:4646\"\n", firstHostIP))
	sb.WriteString("\n")

	// SSH configuration
	sb.WriteString(fmt.Sprintf("ssh_user = %q\n", cfg.Secrets.SSHUser))
	sb.WriteString(fmt.Sprintf("ssh_key_path = %q\n", cfg.Secrets.SSHKeyPath))
	sb.WriteString("\n")

	// Hosts map
	sb.WriteString("hosts = {\n")
	for hostName, hostCfg := range cfg.Hosts {
		sb.WriteString(fmt.Sprintf("  %q = {\n", hostName))
		sb.WriteString(fmt.Sprintf("    ip = %q\n", hostCfg.IP))
		sb.WriteString(fmt.Sprintf("    nat_subnet = %q\n", hostCfg.NATSubnet))
		sb.WriteString(fmt.Sprintf("    wireguard_endpoint = %t\n", hostCfg.WireGuardEndpoint))
		sb.WriteString("  }\n")
	}
	sb.WriteString("}\n\n")

	// VPN configuration
	sb.WriteString(fmt.Sprintf("vpn_subnet = %q\n", cfg.VPN.Subnet))
	sb.WriteString(fmt.Sprintf("vpn_endpoint = %q\n", cfg.VPN.Endpoint))
	sb.WriteString(fmt.Sprintf("vpn_port = %d\n", cfg.VPN.Port))
	if len(cfg.VPN.LANRoutes) > 0 {
		sb.WriteString("vpn_lan_routes = [\n")
		for _, route := range cfg.VPN.LANRoutes {
			sb.WriteString(fmt.Sprintf("  %q,\n", route))
		}
		sb.WriteString("]\n")
	} else {
		sb.WriteString("vpn_lan_routes = []\n")
	}
	sb.WriteString("\n")

	// Storage configuration
	sb.WriteString(fmt.Sprintf("media_path = %q\n", cfg.Storage.MediaPath))
	sb.WriteString(fmt.Sprintf("downloads_path = %q\n", cfg.Storage.DownloadsPath))
	sb.WriteString(fmt.Sprintf("config_path = %q\n", cfg.Storage.ConfigPath))
	sb.WriteString("\n")

	// Apps configuration
	sb.WriteString("apps = {\n")
	for appName, appCfg := range cfg.Apps {
		sb.WriteString(fmt.Sprintf("  %q = {\n", appName))
		sb.WriteString(fmt.Sprintf("    enabled = %t\n", appCfg.Enabled))
		sb.WriteString(fmt.Sprintf("    image = %q\n", appCfg.Image))
		sb.WriteString(fmt.Sprintf("    port = %d\n", appCfg.Port))
		sb.WriteString("  }\n")
	}
	sb.WriteString("}\n\n")

	// Auto-update configuration
	sb.WriteString(fmt.Sprintf("auto_update_enabled = %t\n", cfg.AutoUpdate.Enabled))
	sb.WriteString(fmt.Sprintf("auto_update_schedule = %q\n", cfg.AutoUpdate.Schedule))
	sb.WriteString(fmt.Sprintf("auto_update_auto_revert = %t\n", cfg.AutoUpdate.AutoRevert))
	sb.WriteString(fmt.Sprintf("auto_update_health_check_timeout = %q\n", cfg.AutoUpdate.HealthCheckTimeout))

	// Write to file
	tfvarsPath := filepath.Join(outputDir, "terraform.tfvars")
	if err := os.WriteFile(tfvarsPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing terraform.tfvars: %w", err)
	}

	return nil
}

// getFirstHostIP returns the IP of the first host in the map.
// Map iteration order is random, but we need a consistent value,
// so we could enhance this to sort by name first if needed.
func getFirstHostIP(cfg *config.Config) string {
	for _, hostCfg := range cfg.Hosts {
		return hostCfg.IP
	}
	return ""
}
