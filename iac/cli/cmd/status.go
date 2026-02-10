package cmd

import (
	"encoding/json"
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

	fmt.Printf("Cluster: %s (datacenter: %s)\n\n", cfg.Cluster.Name, cfg.Cluster.Datacenter)

	// Hosts table
	printHostsTable(cfg)

	// Nomad jobs
	printNomadJobs()

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

func printNomadJobs() {
	fmt.Println("Nomad Jobs:")

	// Query Nomad API for job status
	out, err := exec.Command("nomad", "job", "status", "-json").Output()
	if err != nil {
		fmt.Println("  (Nomad not reachable or no jobs running)")
		fmt.Println()
		return
	}

	var jobs []struct {
		ID     string `json:"ID"`
		Status string `json:"Status"`
		Type   string `json:"Type"`
	}

	if err := json.Unmarshal(out, &jobs); err != nil {
		fmt.Printf("  (Error parsing Nomad output: %v)\n\n", err)
		return
	}

	if len(jobs) == 0 {
		fmt.Println("  (No jobs registered)")
		fmt.Println()
		return
	}

	fmt.Printf("  %-20s %-12s %-10s\n", "JOB", "STATUS", "TYPE")
	fmt.Printf("  %-20s %-12s %-10s\n", "---", "------", "----")
	for _, job := range jobs {
		fmt.Printf("  %-20s %-12s %-10s\n", job.ID, job.Status, job.Type)
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
