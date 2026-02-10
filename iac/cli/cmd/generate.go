package cmd

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/homelab-vibe/iac/config"
	"github.com/homelab-vibe/iac/generators"
)

// Generate handles the "iac generate" command and routes to subcommands.
func Generate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing subcommand. Usage: iac generate [vpn-client]")
	}

	switch args[0] {
	case "vpn-client":
		return generateVPNClient(args[1:])
	case "configs":
		return generateConfigs(args[1:])
	default:
		return fmt.Errorf("unknown generate subcommand: %s", args[0])
	}
}

func generateVPNClient(args []string) error {
	fs := flag.NewFlagSet("generate vpn-client", flag.ExitOnError)
	peerName := fs.String("name", "", "Name for the VPN client/peer")
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *peerName == "" {
		return fmt.Errorf("--name is required. Usage: iac generate vpn-client --name <peer_name>")
	}

	_, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// TODO: Implement WireGuard client config generation (Phase 4)
	fmt.Printf("Generating VPN client config for peer: %s\n", *peerName)
	fmt.Println("  (Not yet implemented — see Phase 4: US1 VPN)")

	return nil
}

func generateConfigs(args []string) error {
	fs := flag.NewFlagSet("generate configs", flag.ExitOnError)
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	generatedDir := filepath.Join(repoRoot, ".generated")
	templateDir := filepath.Join(repoRoot, "iac", "templates")

	fmt.Println("Generating all tool-specific configs...")

	if err := generators.GenerateAnsibleInventory(cfg, filepath.Join(generatedDir, "ansible", "inventory")); err != nil {
		return fmt.Errorf("generating Ansible inventory: %w", err)
	}

	if err := generators.GenerateTerraformVars(cfg, filepath.Join(generatedDir, "terraform")); err != nil {
		return fmt.Errorf("generating Terraform vars: %w", err)
	}

	if err := generators.GenerateNomadJobs(cfg, filepath.Join(generatedDir, "nomad"), templateDir); err != nil {
		return fmt.Errorf("generating Nomad jobs: %w", err)
	}

	fmt.Println("All configs generated into .generated/")
	return nil
}
