package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/homelab-vibe/iac/config"
)

// GenerateNomadJobs generates Nomad job files from templates.
func GenerateNomadJobs(cfg *config.Config, outputDir string, templateDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", outputDir, err)
	}

	// Generate app jobs for enabled apps
	enabledApps := cfg.EnabledApps()
	for appName, appCfg := range enabledApps {
		if err := generateAppJob(appName, appCfg, cfg, outputDir, templateDir); err != nil {
			return fmt.Errorf("generating job for app %s: %w", appName, err)
		}
	}

	// Generate proxy job
	if err := generateProxyJob(cfg, outputDir, templateDir); err != nil {
		return fmt.Errorf("generating proxy job: %w", err)
	}

	// Generate auto-updater job
	if err := generateAutoUpdaterJob(cfg, outputDir, templateDir); err != nil {
		return fmt.Errorf("generating auto-updater job: %w", err)
	}

	return nil
}

// generateAppJob generates a Nomad job file for a specific app.
func generateAppJob(appName string, appCfg config.AppConfig, cfg *config.Config, outputDir, templateDir string) error {
	templatePath := filepath.Join(templateDir, "nomad-app.hcl.tmpl")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", templatePath, err)
	}

	// Prepare template data
	data := map[string]interface{}{
		"AppName":             appName,
		"Image":               appCfg.Image,
		"Port":                appCfg.Port,
		"MediaPath":           cfg.Storage.MediaPath,
		"DownloadsPath":       cfg.Storage.DownloadsPath,
		"ConfigPath":          cfg.Storage.ConfigPath,
		"ClusterName":         cfg.Cluster.Name,
		"Datacenter":          cfg.Cluster.Datacenter,
		"AutoUpdateEnabled":   cfg.AutoUpdate.Enabled,
		"AutoUpdateSchedule":  cfg.AutoUpdate.Schedule,
		"AutoRevert":          cfg.AutoUpdate.AutoRevert,
		"HealthCheckTimeout":  cfg.AutoUpdate.HealthCheckTimeout,
	}

	// Execute template
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.hcl", appName))
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outputPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template for %s: %w", appName, err)
	}

	return nil
}

// generateProxyJob generates the proxy Nomad job file.
func generateProxyJob(cfg *config.Config, outputDir, templateDir string) error {
	templatePath := filepath.Join(templateDir, "nomad-proxy.hcl.tmpl")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", templatePath, err)
	}

	// Build apps map (name -> port) for proxy routing
	enabledApps := cfg.EnabledApps()
	appPorts := make(map[string]int, len(enabledApps))
	for name, app := range enabledApps {
		appPorts[name] = app.Port
	}
	data := map[string]interface{}{
		"ClusterName": cfg.Cluster.Name,
		"Datacenter":  cfg.Cluster.Datacenter,
		"Apps":        appPorts,
		"ConfigPath":  cfg.Storage.ConfigPath,
	}

	// Execute template
	outputPath := filepath.Join(outputDir, "proxy.hcl")
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outputPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing proxy template: %w", err)
	}

	return nil
}

// generateAutoUpdaterJob generates the auto-updater Nomad job file.
func generateAutoUpdaterJob(cfg *config.Config, outputDir, templateDir string) error {
	templatePath := filepath.Join(templateDir, "nomad-updater.hcl.tmpl")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", templatePath, err)
	}

	// Build app name list for the updater
	enabledApps := cfg.EnabledApps()
	appNames := make([]string, 0, len(enabledApps))
	for name := range enabledApps {
		appNames = append(appNames, name)
	}
	data := map[string]interface{}{
		"ClusterName":        cfg.Cluster.Name,
		"Datacenter":         cfg.Cluster.Datacenter,
		"Enabled":            cfg.AutoUpdate.Enabled,
		"Schedule":           cfg.AutoUpdate.Schedule,
		"AutoRevert":         cfg.AutoUpdate.AutoRevert,
		"HealthCheckTimeout": cfg.AutoUpdate.HealthCheckTimeout,
		"Apps":               appNames,
	}

	// Execute template
	outputPath := filepath.Join(outputDir, "auto-updater.hcl")
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outputPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing auto-updater template: %w", err)
	}

	return nil
}
