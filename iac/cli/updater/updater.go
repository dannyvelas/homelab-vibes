package updater

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// AppUpdateResult represents the result of checking one app for updates.
type AppUpdateResult struct {
	AppName       string
	CurrentDigest string
	LatestDigest  string
	NeedsUpdate   bool
	Error         error
}

// CheckForUpdates checks if a Docker image has a newer version available.
func CheckForUpdates(appName, currentImage string) AppUpdateResult {
	result := AppUpdateResult{AppName: appName}

	// Pull the latest image to get its digest
	if err := exec.Command("docker", "pull", currentImage).Run(); err != nil {
		result.Error = fmt.Errorf("pulling image %s: %w", currentImage, err)
		return result
	}

	// Get the digest of the pulled image
	latestDigest, err := getImageDigest(currentImage)
	if err != nil {
		result.Error = fmt.Errorf("getting latest digest: %w", err)
		return result
	}
	result.LatestDigest = latestDigest

	// Get the currently running container's image digest
	runningDigest, err := getRunningDigest(appName)
	if err != nil {
		result.Error = fmt.Errorf("getting running digest: %w", err)
		return result
	}
	result.CurrentDigest = runningDigest

	result.NeedsUpdate = result.CurrentDigest != result.LatestDigest
	return result
}

// TriggerDeployment triggers a Nomad job redeployment for the given app.
func TriggerDeployment(appName, jobFile string) error {
	cmd := exec.Command("nomad", "job", "run", jobFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nomad job run %s: %s: %w", appName, string(output), err)
	}
	return nil
}

// GetCurrentImage returns the Docker image for a running Nomad job.
func GetCurrentImage(appName string) (string, error) {
	out, err := exec.Command("nomad", "job", "inspect", appName).Output()
	if err != nil {
		return "", fmt.Errorf("inspecting nomad job %s: %w", appName, err)
	}

	var job struct {
		Job struct {
			TaskGroups []struct {
				Tasks []struct {
					Config map[string]interface{} `json:"Config"`
				} `json:"Tasks"`
			} `json:"TaskGroups"`
		} `json:"Job"`
	}

	if err := json.Unmarshal(out, &job); err != nil {
		return "", fmt.Errorf("parsing job inspect output: %w", err)
	}

	if len(job.Job.TaskGroups) > 0 && len(job.Job.TaskGroups[0].Tasks) > 0 {
		if img, ok := job.Job.TaskGroups[0].Tasks[0].Config["image"].(string); ok {
			return img, nil
		}
	}

	return "", fmt.Errorf("could not find image in job %s", appName)
}

// getImageDigest returns the digest of a local Docker image.
func getImageDigest(image string) (string, error) {
	out, err := exec.Command("docker", "inspect", "--format", "{{.Id}}", image).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// getRunningDigest returns the image digest of a running container.
func getRunningDigest(appName string) (string, error) {
	// Find container by name filter
	out, err := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", appName), "-q").Output()
	if err != nil {
		return "", err
	}

	containerID := strings.TrimSpace(string(out))
	if containerID == "" {
		return "", fmt.Errorf("no running container found for %s", appName)
	}

	// Get the image digest of the running container
	digest, err := exec.Command("docker", "inspect", "--format", "{{.Image}}", containerID).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(digest)), nil
}
