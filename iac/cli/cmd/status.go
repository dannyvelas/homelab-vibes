package cmd

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/homelab-vibe/iac/config"
)

// Status handles the "iac status" command.
func Status(args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Printf("Cluster: %s\n\n", cfg.Cluster.Name)

	// Hosts table
	printHostsTable(cfg)

	// Docker containers on each VM
	printContainerStatus(cfg)

	// WireGuard status
	if wgHost := cfg.WireGuardHost(); wgHost != "" {
		printWireGuardStatus(cfg, wgHost)
	}

	return nil
}

func printHostsTable(cfg *config.Config) {
	fmt.Println("Hosts:")
	fmt.Printf("  %-25s %-16s %-22s %-10s\n", "NAME", "IP", "NAT SUBNET", "VPN")
	fmt.Printf("  %-25s %-16s %-22s %-10s\n", "----", "--", "----------", "---")

	for name, host := range cfg.Hosts {
		vpn := ""
		if host.WireGuardEndpoint {
			vpn = "endpoint"
		}
		fmt.Printf("  %-25s %-16s %-22s %-10s\n", name, host.IP, host.NATSubnet, vpn)
	}
	fmt.Println()
}

func printContainerStatus(cfg *config.Config) {
	fmt.Println("Containers:")

	for hostName, hostCfg := range cfg.Hosts {
		vmIP := hostCfg.IP // Will query via SSH to the VM
		fmt.Printf("  Host: %s (%s)\n", hostName, vmIP)

		// Query Docker containers on the VM via SSH
		out, err := exec.Command("ssh", "-o", "ConnectTimeout=3", "-o", "StrictHostKeyChecking=no",
			fmt.Sprintf("%s@%s", cfg.Secrets.SSHUser, vmIP),
			"docker ps --format '{{.Names}}\\t{{.Image}}\\t{{.Status}}\\t{{.Ports}}'",
		).Output()
		if err != nil {
			fmt.Println("    (unable to query containers)")
			continue
		}

		lines := strings.TrimSpace(string(out))
		if lines == "" {
			fmt.Println("    (no containers running)")
			continue
		}

		fmt.Printf("    %-20s %-35s %-25s %s\n", "NAME", "IMAGE", "STATUS", "PORTS")
		fmt.Printf("    %-20s %-35s %-25s %s\n", "----", "-----", "------", "-----")
		for _, line := range strings.Split(lines, "\n") {
			parts := strings.SplitN(line, "\t", 4)
			if len(parts) >= 3 {
				name := parts[0]
				image := parts[1]
				status := parts[2]
				ports := ""
				if len(parts) == 4 {
					ports = parts[3]
				}
				fmt.Printf("    %-20s %-35s %-25s %s\n", name, image, status, ports)
			}
		}
	}
	fmt.Println()
}

func printWireGuardStatus(cfg *config.Config, wgHost string) {
	host := cfg.Hosts[wgHost]
	fmt.Println("WireGuard VPN:")
	fmt.Printf("  Endpoint host: %s (%s)\n", wgHost, host.IP)
	fmt.Printf("  VPN subnet: %s\n", cfg.VPN.Subnet)
	fmt.Printf("  Listen port: %d\n", cfg.VPN.Port)

	// Try to get peer count via SSH
	out, err := exec.Command("ssh", "-o", "ConnectTimeout=3",
		fmt.Sprintf("%s@%s", cfg.Secrets.SSHUser, host.IP),
		"wg show wg0 peers | wc -l",
	).Output()
	if err == nil {
		peers := strings.TrimSpace(string(out))
		fmt.Printf("  Connected peers: %s\n", peers)
	} else {
		fmt.Println("  Connected peers: (unable to query)")
	}
	fmt.Println()
}
