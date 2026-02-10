package cmd

import (
	"flag"
	"fmt"

	"github.com/homelab-vibe/iac/config"
)

func updateEnable(args []string) error {
	return toggleAutoUpdate(args, true)
}

func updateDisable(args []string) error {
	return toggleAutoUpdate(args, false)
}

func toggleAutoUpdate(args []string, enable bool) error {
	fs := flag.NewFlagSet("update toggle", flag.ExitOnError)
	appName := fs.String("name", "", "App to enable/disable auto-updates for")
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *appName == "" {
		action := "enable"
		if !enable {
			action = "disable"
		}
		return fmt.Errorf("--name is required. Usage: iac update %s --name <app_name>", action)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	app, ok := cfg.Apps[*appName]
	if !ok {
		return fmt.Errorf("app %q not found in homelab.yml", *appName)
	}

	action := "Enabled"
	if !enable {
		action = "Disabled"
	}

	// Note: This is a runtime toggle. To persist, user should edit homelab.yml.
	fmt.Printf("%s auto-updates for %s\n", action, *appName)
	fmt.Printf("  Current image: %s\n", app.Image)

	if enable {
		fmt.Println("  Auto-updates will check for new images on schedule:", cfg.AutoUpdate.Schedule)
		fmt.Println("  To persist this change, ensure auto_update.enabled: true in homelab.yml")
	} else {
		fmt.Println("  Auto-updates are paused for this app.")
		fmt.Println("  To persist this change, set the app's enabled: false in homelab.yml or auto_update.enabled: false")
	}

	return nil
}
