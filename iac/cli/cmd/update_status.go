package cmd

import (
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
	fmt.Println()

	// Query systemd timer status on each VM via SSH
	for hostName, hostCfg := range cfg.Hosts {
		fmt.Printf("  Host: %s (%s)\n", hostName, hostCfg.IP)

		// Check timer status
		out, err := exec.Command("ssh", "-o", "ConnectTimeout=3", "-o", "StrictHostKeyChecking=no",
			fmt.Sprintf("%s@%s", cfg.Secrets.SSHUser, hostCfg.IP),
			"systemctl is-active homelab-updater.timer 2>/dev/null && systemctl show homelab-updater.timer --property=LastTriggerUSec --value 2>/dev/null",
		).Output()
		if err != nil {
			fmt.Println("    Timer: (unable to query)")
			continue
		}

		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) >= 1 {
			fmt.Printf("    Timer: %s\n", lines[0])
		}
		if len(lines) >= 2 {
			fmt.Printf("    Last run: %s\n", lines[1])
		}

		// Get recent journal output
		logOut, err := exec.Command("ssh", "-o", "ConnectTimeout=3", "-o", "StrictHostKeyChecking=no",
			fmt.Sprintf("%s@%s", cfg.Secrets.SSHUser, hostCfg.IP),
			"journalctl -u homelab-updater.service --no-pager -n 5 --output=short-iso 2>/dev/null",
		).Output()
		if err == nil && strings.TrimSpace(string(logOut)) != "" {
			fmt.Println("    Recent logs:")
			for _, line := range strings.Split(strings.TrimSpace(string(logOut)), "\n") {
				fmt.Printf("      %s\n", line)
			}
		}
	}

	fmt.Println()
	return nil
}

func boolStatus(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
