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

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	tfDir := filepath.Join(repoRoot, "iac", "terraform")
	generatedDir := filepath.Join(repoRoot, ".generated")

	fmt.Println("Tearing down all infrastructure...")

	for hostName := range cfg.Hosts {
		fmt.Printf("  Destroying VMs on host: %s\n", hostName)
		tfVarsFile := filepath.Join(generatedDir, "terraform", hostName, "terraform.tfvars")
		tfStateFile := filepath.Join(generatedDir, "terraform", hostName, "terraform.tfstate")

		if err := runCommand(tfDir, "terraform", "destroy", "-auto-approve",
			"-var-file="+tfVarsFile, "-state="+tfStateFile); err != nil {
			return fmt.Errorf("terraform destroy for %s: %w", hostName, err)
		}
	}

	fmt.Println("\nAll VMs and Nomad jobs destroyed.")
	return nil
}
