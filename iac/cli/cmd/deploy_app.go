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

	// Generate Ansible inventory
	if err := generators.GenerateAnsibleInventory(cfg, filepath.Join(generatedDir, "ansible", "inventory")); err != nil {
		return fmt.Errorf("generating Ansible inventory: %w", err)
	}

	// Run Ansible playbook to deploy the app container
	inventoryFile := filepath.Join(generatedDir, "ansible", "inventory", "hosts.yml")
	playbookFile := filepath.Join(repoRoot, "iac", "ansible", "playbooks", "deploy-app.yml")

	if err := runCommand(repoRoot, "ansible-playbook",
		"-i", inventoryFile,
		playbookFile,
		"-e", fmt.Sprintf("app_name=%s", *appName),
		"-e", fmt.Sprintf("app_image=%s", app.Image),
		"-e", fmt.Sprintf("app_port=%d", app.Port),
		"-e", fmt.Sprintf("media_path=%s", cfg.Storage.MediaPath),
		"-e", fmt.Sprintf("downloads_path=%s", cfg.Storage.DownloadsPath),
		"-e", fmt.Sprintf("config_path=%s", cfg.Storage.ConfigPath),
	); err != nil {
		return fmt.Errorf("ansible-playbook deploy-app: %w", err)
	}

	// Deploy auto-updater if enabled
	if cfg.AutoUpdate.Enabled {
		updaterPlaybook := filepath.Join(repoRoot, "iac", "ansible", "playbooks", "deploy-updater.yml")
		if err := runCommand(repoRoot, "ansible-playbook",
			"-i", inventoryFile,
			updaterPlaybook,
		); err != nil {
			return fmt.Errorf("ansible-playbook deploy-updater: %w", err)
		}
	}

	fmt.Printf("\n%s deployed!\n", *appName)
	fmt.Printf("  Image: %s\n", app.Image)
	fmt.Printf("  Port: %d\n", app.Port)
	if cfg.AutoUpdate.Enabled {
		fmt.Printf("  Auto-update: enabled (schedule: %s)\n", cfg.AutoUpdate.Schedule)
	}

	return nil
}
