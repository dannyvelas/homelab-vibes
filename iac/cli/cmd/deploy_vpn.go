package cmd

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/homelab-vibe/iac/config"
	"github.com/homelab-vibe/iac/generators"
)

// Deploy handles the "iac deploy" command and routes to subcommands.
func Deploy(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing subcommand. Usage: iac deploy [vpn|proxy|app]")
	}

	switch args[0] {
	case "vpn":
		return deployVPN(args[1:])
	case "proxy":
		return deployProxy(args[1:])
	case "app":
		return deployApp(args[1:])
	default:
		return fmt.Errorf("unknown deploy subcommand: %s", args[0])
	}
}

func deployVPN(args []string) error {
	fs := flag.NewFlagSet("deploy vpn", flag.ExitOnError)
	hostName := fs.String("host", "", "Target host for WireGuard deployment")
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Default to the designated WireGuard endpoint host
	target := *hostName
	if target == "" {
		target = cfg.WireGuardHost()
		if target == "" {
			return fmt.Errorf("no host has wireguard_endpoint: true in homelab.yml, and --host not specified")
		}
	}

	if _, ok := cfg.Hosts[target]; !ok {
		return fmt.Errorf("host %q not found in homelab.yml", target)
	}

	fmt.Printf("Deploying WireGuard VPN on host: %s\n", target)

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	// Generate configs
	generatedDir := filepath.Join(repoRoot, ".generated")
	if err := generators.GenerateAnsibleInventory(cfg, filepath.Join(generatedDir, "ansible", "inventory")); err != nil {
		return fmt.Errorf("generating Ansible inventory: %w", err)
	}

	// Run VPN deployment playbook
	inventoryFile := filepath.Join(generatedDir, "ansible", "inventory", "hosts.yml")
	playbookFile := filepath.Join(repoRoot, "iac", "ansible", "playbooks", "deploy-vpn.yml")

	if err := runCommand(repoRoot, "ansible-playbook",
		"-i", inventoryFile,
		playbookFile,
		"--limit", target,
		"-e", fmt.Sprintf("vpn_subnet=%s", cfg.VPN.Subnet),
		"-e", fmt.Sprintf("vpn_endpoint=%s", cfg.VPN.Endpoint),
		"-e", fmt.Sprintf("vpn_port=%d", cfg.VPN.Port),
	); err != nil {
		return fmt.Errorf("ansible-playbook: %w", err)
	}

	host := cfg.Hosts[target]
	fmt.Printf("\nWireGuard VPN deployed on %s (%s)\n", target, host.IP)
	fmt.Printf("  Endpoint: %s:%d\n", cfg.VPN.Endpoint, cfg.VPN.Port)
	fmt.Printf("  VPN subnet: %s\n", cfg.VPN.Subnet)
	fmt.Println("  Generate client configs with: iac generate vpn-client --name <peer_name>")

	return nil
}
