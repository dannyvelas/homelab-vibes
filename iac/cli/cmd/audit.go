package cmd

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/homelab-vibe/iac/config"
	"github.com/homelab-vibe/iac/generators"
)

// Audit handles the "iac audit" command and routes to subcommands.
func Audit(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing subcommand. Usage: iac audit [security]")
	}

	switch args[0] {
	case "security":
		return auditSecurity(args[1:])
	default:
		return fmt.Errorf("unknown audit subcommand: %s", args[0])
	}
}

func auditSecurity(args []string) error {
	fs := flag.NewFlagSet("audit security", flag.ExitOnError)
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Println("Running security audit...")

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	generatedDir := filepath.Join(repoRoot, ".generated")

	// Ensure inventory is up to date
	if err := generators.GenerateAnsibleInventory(cfg, filepath.Join(generatedDir, "ansible", "inventory")); err != nil {
		return fmt.Errorf("generating Ansible inventory: %w", err)
	}

	// Run security audit playbook
	inventoryFile := filepath.Join(generatedDir, "ansible", "inventory", "hosts.yml")
	playbookFile := filepath.Join(repoRoot, "iac", "ansible", "playbooks", "security-audit.yml")

	if err := runCommand(repoRoot, "ansible-playbook",
		"-i", inventoryFile,
		playbookFile,
	); err != nil {
		return fmt.Errorf("security audit playbook: %w", err)
	}

	fmt.Println("\nSecurity audit complete.")
	return nil
}
