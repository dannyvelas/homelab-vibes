package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/homelab-vibe/iac/config"
	"github.com/homelab-vibe/iac/generators"
)

// Provision handles the "iac provision" command.
func Provision(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing subcommand. Usage: iac provision host --name <host_name>")
	}

	switch args[0] {
	case "host":
		return provisionHost(args[1:])
	default:
		return fmt.Errorf("unknown provision subcommand: %s", args[0])
	}
}

func provisionHost(args []string) error {
	fs := flag.NewFlagSet("provision host", flag.ExitOnError)
	name := fs.String("name", "", "Host name to provision (must match homelab.yml)")
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("--name is required. Usage: iac provision host --name <host_name>")
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Validate host exists
	if _, ok := cfg.Hosts[*name]; !ok {
		return fmt.Errorf("host %q not found in homelab.yml. Available hosts: %v", *name, hostNames(cfg))
	}

	fmt.Printf("Provisioning host: %s\n", *name)

	// Determine paths
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}
	generatedDir := filepath.Join(repoRoot, ".generated")
	templateDir := filepath.Join(repoRoot, "iac", "templates")

	// Step 1: Generate all tool-specific configs
	fmt.Println("  [1/3] Generating tool-specific configs from homelab.yml...")

	if err := generators.GenerateAnsibleInventory(cfg, filepath.Join(generatedDir, "ansible", "inventory")); err != nil {
		return fmt.Errorf("generating Ansible inventory: %w", err)
	}

	if err := generators.GenerateTerraformVars(cfg, filepath.Join(generatedDir, "terraform")); err != nil {
		return fmt.Errorf("generating Terraform vars: %w", err)
	}

	if err := generators.GenerateNomadJobs(cfg, filepath.Join(generatedDir, "nomad"), templateDir); err != nil {
		return fmt.Errorf("generating Nomad jobs: %w", err)
	}

	fmt.Println("  Configs generated into .generated/")

	// Step 2: Run Terraform apply for VM lifecycle
	fmt.Println("  [2/3] Running Terraform apply (VM lifecycle)...")
	tfDir := filepath.Join(repoRoot, "iac", "terraform")
	tfVarsFile := filepath.Join(generatedDir, "terraform", "terraform.tfvars")

	if err := runCommand(tfDir, "terraform", "init", "-input=false"); err != nil {
		return fmt.Errorf("terraform init: %w", err)
	}
	if err := runCommand(tfDir, "terraform", "apply", "-auto-approve", "-var-file="+tfVarsFile); err != nil {
		return fmt.Errorf("terraform apply: %w", err)
	}

	// Step 3: Run Ansible provisioning playbook
	fmt.Println("  [3/3] Running Ansible provisioning playbook...")
	inventoryFile := filepath.Join(generatedDir, "ansible", "inventory", "hosts.yml")
	playbookFile := filepath.Join(repoRoot, "iac", "ansible", "playbooks", "provision-host.yml")

	if err := runCommand(repoRoot, "ansible-playbook",
		"-i", inventoryFile,
		playbookFile,
		"--limit", *name,
	); err != nil {
		return fmt.Errorf("ansible-playbook: %w", err)
	}

	host := cfg.Hosts[*name]
	fmt.Printf("\nHost %s provisioned successfully!\n", *name)
	fmt.Printf("  Host IP: %s\n", host.IP)
	fmt.Printf("  NAT Subnet: %s\n", host.NATSubnet)
	fmt.Println("  Nomad cluster status: run 'iac status' to check")

	return nil
}

// runCommand executes a command with stdout/stderr forwarded.
func runCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// findRepoRoot walks up from CWD to find the repo root (directory containing homelab.yml or .git).
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "homelab.yml")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repo root (no homelab.yml or .git found)")
		}
		dir = parent
	}
}

// hostNames returns a slice of host names from the config.
func hostNames(cfg *config.Config) []string {
	names := make([]string, 0, len(cfg.Hosts))
	for name := range cfg.Hosts {
		names = append(names, name)
	}
	return names
}
