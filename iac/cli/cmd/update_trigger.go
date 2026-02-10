package cmd

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/homelab-vibe/iac/config"
	"github.com/homelab-vibe/iac/generators"
)

func updateTrigger(args []string) error {
	fs := flag.NewFlagSet("update trigger", flag.ExitOnError)
	appName := fs.String("name", "", "App to trigger update for (or 'all')")
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *appName == "" {
		return fmt.Errorf("--name is required. Usage: iac update trigger --name <app_name|all>")
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

	if *appName == "all" {
		fmt.Println("Triggering update check for all apps...")

		// Regenerate and submit all app jobs
		if err := generators.GenerateNomadJobs(cfg, filepath.Join(generatedDir, "nomad"), templateDir); err != nil {
			return fmt.Errorf("generating Nomad jobs: %w", err)
		}

		for name, app := range cfg.EnabledApps() {
			jobFile := filepath.Join(generatedDir, "nomad", name+".hcl")
			fmt.Printf("  Updating %s (image: %s)...\n", name, app.Image)
			if err := runCommand(repoRoot, "nomad", "job", "run", jobFile); err != nil {
				fmt.Printf("  Warning: failed to update %s: %v\n", name, err)
			}
		}
	} else {
		app, ok := cfg.Apps[*appName]
		if !ok {
			return fmt.Errorf("app %q not found in homelab.yml", *appName)
		}
		if !app.Enabled {
			return fmt.Errorf("app %q is disabled in homelab.yml", *appName)
		}

		fmt.Printf("Triggering update for %s...\n", *appName)

		// Regenerate and submit the specific app job
		if err := generators.GenerateNomadJobs(cfg, filepath.Join(generatedDir, "nomad"), templateDir); err != nil {
			return fmt.Errorf("generating Nomad jobs: %w", err)
		}

		jobFile := filepath.Join(generatedDir, "nomad", *appName+".hcl")
		if err := runCommand(repoRoot, "nomad", "job", "run", jobFile); err != nil {
			return fmt.Errorf("nomad job run %s: %w", *appName, err)
		}

		fmt.Printf("\n%s update triggered. Nomad will perform a health-checked deployment.\n", *appName)
		if cfg.AutoUpdate.AutoRevert {
			fmt.Println("  Auto-revert is enabled — will rollback if health checks fail.")
		}
	}

	return nil
}
