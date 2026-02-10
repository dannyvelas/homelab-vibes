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
	templateDir := filepath.Join(repoRoot, "iac", "templates")

	// Generate Nomad job files
	if err := generators.GenerateNomadJobs(cfg, filepath.Join(generatedDir, "nomad"), templateDir); err != nil {
		return fmt.Errorf("generating Nomad jobs: %w", err)
	}

	// Submit proxy job to Nomad
	proxyJob := filepath.Join(generatedDir, "nomad", "proxy.hcl")
	if err := runCommand(repoRoot, "nomad", "job", "run", proxyJob); err != nil {
		return fmt.Errorf("nomad job run proxy: %w", err)
	}

	fmt.Println("\nReverse proxy deployed!")
	fmt.Println("  Listening on port 443 (HTTPS)")

	return nil
}
