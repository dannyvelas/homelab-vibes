package main

import (
	"fmt"
	"os"

	"github.com/homelab-vibe/iac/cmd"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "provision":
		if err := cmd.Provision(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "deploy":
		if err := cmd.Deploy(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "generate":
		if err := cmd.Generate(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "status":
		if err := cmd.Status(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "update":
		if err := cmd.Update(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "audit":
		if err := cmd.Audit(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "teardown":
		if err := cmd.Teardown(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Printf("iac version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`iac - Infrastructure as Code CLI for homelab management (v%s)

Usage: iac <command> [options]

Commands:
  provision host    Provision a physical host (hardening, hypervisor, Nomad, VM)
  deploy vpn        Deploy WireGuard VPN on designated host
  deploy proxy      Deploy reverse proxy into workload VMs
  deploy app        Deploy a media application (plex, sonarr, radarr, bazarr)
  generate vpn-client  Generate a WireGuard client config
  status            Show cluster status (hosts, VMs, jobs, VPN)
  update status     Show auto-update status for all apps
  update trigger    Manually trigger an update check
  update enable     Enable auto-updates for an app
  update disable    Disable auto-updates for an app
  audit security    Run security audit across all infrastructure
  teardown          Tear down all VMs and Nomad jobs

Configuration:
  All configuration is read from homelab.yml at the repository root.
  Tool-specific configs are generated into .generated/ automatically.

`, version)
}
