package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/homelab-vibe/iac/config"
)

func updateStatus(args []string) error {
	fs := flag.NewFlagSet("update status", flag.ExitOnError)
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Println("Auto-Update Status:")
	fmt.Printf("  Global: %s\n", boolStatus(cfg.AutoUpdate.Enabled))
	fmt.Printf("  Schedule: %s\n", cfg.AutoUpdate.Schedule)
	fmt.Printf("  Auto-revert: %s\n", boolStatus(cfg.AutoUpdate.AutoRevert))
	fmt.Printf("  Health check timeout: %s\n", cfg.AutoUpdate.HealthCheckTimeout)
	fmt.Println()

	// Query Nomad for app deployment status
	enabledApps := cfg.EnabledApps()
	if len(enabledApps) == 0 {
		fmt.Println("  No enabled apps.")
		return nil
	}

	fmt.Printf("  %-15s %-15s %-20s %-15s\n", "APP", "STATUS", "IMAGE", "VERSION")
	fmt.Printf("  %-15s %-15s %-20s %-15s\n", "---", "------", "-----", "-------")

	for appName := range enabledApps {
		status, image, version := getAppDeploymentInfo(appName)
		fmt.Printf("  %-15s %-15s %-20s %-15s\n", appName, status, image, version)
	}

	fmt.Println()

	// Check auto-updater job status
	out, err := exec.Command("nomad", "job", "status", "-json", "auto-updater").Output()
	if err != nil {
		fmt.Println("  Auto-updater job: not registered")
	} else {
		var job struct {
			Status string `json:"Status"`
		}
		if json.Unmarshal(out, &job) == nil {
			fmt.Printf("  Auto-updater job: %s\n", job.Status)
		}
	}

	return nil
}

func getAppDeploymentInfo(appName string) (status, image, version string) {
	out, err := exec.Command("nomad", "job", "status", "-json", appName).Output()
	if err != nil {
		return "(not deployed)", "-", "-"
	}

	var job struct {
		Status     string `json:"Status"`
		TaskGroups []struct {
			Tasks []struct {
				Config map[string]interface{} `json:"Config"`
			} `json:"Tasks"`
		} `json:"TaskGroups"`
	}

	if err := json.Unmarshal(out, &job); err != nil {
		return "(error)", "-", "-"
	}

	status = job.Status
	image = "-"
	version = "-"

	if len(job.TaskGroups) > 0 && len(job.TaskGroups[0].Tasks) > 0 {
		if img, ok := job.TaskGroups[0].Tasks[0].Config["image"].(string); ok {
			image = img
			// Extract version tag
			parts := strings.SplitN(img, ":", 2)
			if len(parts) == 2 {
				version = parts[1]
			}
		}
	}

	return status, image, version
}

func boolStatus(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
