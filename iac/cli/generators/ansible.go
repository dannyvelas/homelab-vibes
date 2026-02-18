package generators

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/homelab-vibe/iac/config"
)

// GenerateAnsibleInventory creates Ansible inventory and group_vars files.
func GenerateAnsibleInventory(cfg *config.Config, outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", outputDir, err)
	}

	// Generate hosts.yml inventory
	if err := generateHostsYML(cfg, outputDir); err != nil {
		return fmt.Errorf("generating hosts.yml: %w", err)
	}

	// Generate group_vars/all.yml
	if err := generateGroupVars(cfg, outputDir); err != nil {
		return fmt.Errorf("generating group_vars/all.yml: %w", err)
	}

	return nil
}

// generateHostsYML creates the Ansible inventory file with hosts and VMs.
func generateHostsYML(cfg *config.Config, outputDir string) error {
	var sb strings.Builder

	// Write hosts group
	sb.WriteString("all:\n")
	sb.WriteString("  children:\n")
	sb.WriteString("    hosts:\n")
	sb.WriteString("      hosts:\n")

	for hostName, hostCfg := range cfg.Hosts {
		sb.WriteString(fmt.Sprintf("        %s:\n", hostName))
		sb.WriteString(fmt.Sprintf("          ansible_host: %s\n", hostCfg.IP))
		sb.WriteString(fmt.Sprintf("          nat_subnet: %s\n", hostCfg.NATSubnet))
	}

	// Write VMs group
	sb.WriteString("\n    vms:\n")
	sb.WriteString("      hosts:\n")

	for hostName, hostCfg := range cfg.Hosts {
		vmIP, err := calculateVMIP(hostCfg.NATSubnet)
		if err != nil {
			return fmt.Errorf("calculating VM IP for host %s: %w", hostName, err)
		}

		vmName := fmt.Sprintf("%s-vm", hostName)
		sb.WriteString(fmt.Sprintf("        %s:\n", vmName))
		sb.WriteString(fmt.Sprintf("          ansible_host: %s\n", vmIP))
		sb.WriteString(fmt.Sprintf("          ansible_ssh_common_args: '-o ProxyJump=%s@%s'\n", cfg.Secrets.SSHUser, hostCfg.IP))
		sb.WriteString(fmt.Sprintf("          parent_host: %s\n", hostName))
	}

	// Write to file
	inventoryPath := filepath.Join(outputDir, "hosts.yml")
	if err := os.WriteFile(inventoryPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing hosts.yml: %w", err)
	}

	return nil
}

// generateGroupVars creates the group_vars/all.yml with shared variables.
func generateGroupVars(cfg *config.Config, outputDir string) error {
	groupVarsDir := filepath.Join(outputDir, "group_vars")
	if err := os.MkdirAll(groupVarsDir, 0755); err != nil {
		return fmt.Errorf("creating group_vars directory: %w", err)
	}

	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("cluster_name: %s\n", cfg.Cluster.Name))
	sb.WriteString(fmt.Sprintf("ssh_user: %s\n", cfg.Secrets.SSHUser))
	sb.WriteString("\n")

	// Storage paths
	sb.WriteString("# Storage Configuration\n")
	sb.WriteString(fmt.Sprintf("media_path: %s\n", cfg.Storage.MediaPath))
	sb.WriteString(fmt.Sprintf("downloads_path: %s\n", cfg.Storage.DownloadsPath))
	sb.WriteString(fmt.Sprintf("config_path: %s\n", cfg.Storage.ConfigPath))
	sb.WriteString("\n")

	// VPN configuration
	sb.WriteString("# VPN Configuration\n")
	sb.WriteString(fmt.Sprintf("vpn_subnet: %s\n", cfg.VPN.Subnet))
	sb.WriteString(fmt.Sprintf("vpn_endpoint: %s\n", cfg.VPN.Endpoint))
	sb.WriteString(fmt.Sprintf("vpn_port: %d\n", cfg.VPN.Port))
	if len(cfg.VPN.LANRoutes) > 0 {
		sb.WriteString("vpn_lan_routes:\n")
		for _, route := range cfg.VPN.LANRoutes {
			sb.WriteString(fmt.Sprintf("  - %s\n", route))
		}
	}


	// Write to file
	groupVarsPath := filepath.Join(groupVarsDir, "all.yml")
	if err := os.WriteFile(groupVarsPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing group_vars/all.yml: %w", err)
	}

	return nil
}

// calculateVMIP computes the VM IP from a NAT subnet by adding 49 to the base IP.
// For example, 192.168.122.0/24 becomes 192.168.122.50
func calculateVMIP(subnet string) (string, error) {
	ip, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", fmt.Errorf("parsing subnet %s: %w", subnet, err)
	}

	// Convert IP to 4-byte representation
	ipv4 := ip.To4()
	if ipv4 == nil {
		return "", fmt.Errorf("subnet %s is not a valid IPv4 address", subnet)
	}

	// Add 49 to get .50 (since base is typically .1)
	// This handles the last octet properly
	ipv4[3] += 49

	return ipv4.String(), nil
}
