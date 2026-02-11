package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/homelab-vibe/iac/config"
)

// GenerateTerraformVars creates per-host terraform.tfvars files.
// Each host gets its own directory under outputDir/<hostname>/ with a
// terraform.tfvars that points the libvirt provider at that specific host.
func GenerateTerraformVars(cfg *config.Config, outputDir string) error {
	for hostName, hostCfg := range cfg.Hosts {
		hostDir := filepath.Join(outputDir, hostName)
		if err := os.MkdirAll(hostDir, 0755); err != nil {
			return fmt.Errorf("creating output directory %s: %w", hostDir, err)
		}

		content := generateHostTfvars(cfg, hostName, hostCfg)

		tfvarsPath := filepath.Join(hostDir, "terraform.tfvars")
		if err := os.WriteFile(tfvarsPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", tfvarsPath, err)
		}
	}

	return nil
}

// generateHostTfvars builds the terraform.tfvars content for a single host.
func generateHostTfvars(cfg *config.Config, hostName string, hostCfg config.HostConfig) string {
	var sb strings.Builder

	// Cluster configuration
	sb.WriteString(fmt.Sprintf("cluster_name = %q\n", cfg.Cluster.Name))
	sb.WriteString("\n")

	// Libvirt connection — scoped to this host
	sb.WriteString(fmt.Sprintf("libvirt_uri = \"qemu+ssh://%s@%s/system\"\n", cfg.Secrets.SSHUser, hostCfg.IP))
	sb.WriteString("\n")

	// SSH configuration
	sb.WriteString(fmt.Sprintf("ssh_user = %q\n", cfg.Secrets.SSHUser))
	sb.WriteString(fmt.Sprintf("ssh_key_path = %q\n", cfg.Secrets.SSHKeyPath))
	sb.WriteString("\n")

	// Hosts map — only this host
	sb.WriteString("hosts = {\n")
	sb.WriteString(fmt.Sprintf("  %q = {\n", hostName))
	sb.WriteString(fmt.Sprintf("    ip = %q\n", hostCfg.IP))
	sb.WriteString(fmt.Sprintf("    nat_subnet = %q\n", hostCfg.NATSubnet))
	sb.WriteString(fmt.Sprintf("    wireguard_endpoint = %t\n", hostCfg.WireGuardEndpoint))
	sb.WriteString("  }\n")
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

	return sb.String()
}
