package cmd

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/homelab-vibe/iac/config"
	"github.com/homelab-vibe/iac/generators"
)

func deployApp(args []string) error {
	fs := flag.NewFlagSet("deploy app", flag.ExitOnError)
	appName := fs.String("name", "", "Application to deploy (plex, sonarr, radarr, bazarr)")
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *appName == "" {
		return fmt.Errorf("--name is required. Usage: iac deploy app --name <app_name>")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	app, ok := cfg.Apps[*appName]
	if !ok {
		return fmt.Errorf("app %q not found in homelab.yml", *appName)
	}
	if !app.Enabled {
		return fmt.Errorf("app %q is disabled in homelab.yml", *appName)
	}

	fmt.Printf("Deploying %s...\n", *appName)

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

	// Submit job to Nomad
	jobFile := filepath.Join(generatedDir, "nomad", *appName+".hcl")
	if err := runCommand(repoRoot, "nomad", "job", "run", jobFile); err != nil {
		return fmt.Errorf("nomad job run %s: %w", *appName, err)
	}

	fmt.Printf("\n%s deployed!\n", *appName)
	fmt.Printf("  Image: %s\n", app.Image)
	fmt.Printf("  Port: %d\n", app.Port)
	if cfg.AutoUpdate.Enabled {
		fmt.Printf("  Auto-update: enabled (schedule: %s)\n", cfg.AutoUpdate.Schedule)
	}

	return nil
}
