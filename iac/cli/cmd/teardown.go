package cmd

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/homelab-vibe/iac/config"
)

// Teardown handles the "iac teardown" command.
func Teardown(args []string) error {
	fs := flag.NewFlagSet("teardown", flag.ExitOnError)
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	_, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	tfDir := filepath.Join(repoRoot, "iac", "terraform")
	generatedDir := filepath.Join(repoRoot, ".generated")
	tfVarsFile := filepath.Join(generatedDir, "terraform", "terraform.tfvars")

	fmt.Println("Tearing down all infrastructure...")

	if err := runCommand(tfDir, "terraform", "destroy", "-auto-approve", "-var-file="+tfVarsFile); err != nil {
		return fmt.Errorf("terraform destroy: %w", err)
	}

	fmt.Println("\nAll VMs and Nomad jobs destroyed.")
	return nil
}
