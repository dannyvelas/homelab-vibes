package cmd

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/homelab-vibe/iac/config"
	"github.com/homelab-vibe/iac/generators"
)

func deployProxy(args []string) error {
	fs := flag.NewFlagSet("deploy proxy", flag.ExitOnError)
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Println("Deploying reverse proxy...")

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	generatedDir := filepath.Join(repoRoot, ".generated")

	// Generate Ansible inventory
	if err := generators.GenerateAnsibleInventory(cfg, filepath.Join(generatedDir, "ansible", "inventory")); err != nil {
		return fmt.Errorf("generating Ansible inventory: %w", err)
	}

	// Run Ansible playbook to deploy the proxy container
	inventoryFile := filepath.Join(generatedDir, "ansible", "inventory", "hosts.yml")
	playbookFile := filepath.Join(repoRoot, "iac", "ansible", "playbooks", "deploy-proxy.yml")

	if err := runCommand(repoRoot, "ansible-playbook",
		"-i", inventoryFile,
		playbookFile,
	); err != nil {
		return fmt.Errorf("ansible-playbook deploy-proxy: %w", err)
	}

	fmt.Println("\nReverse proxy deployed!")
	fmt.Println("  Listening on port 443 (HTTPS)")

	return nil
}
