package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"zed-cli-win-unofficial/internal/telementry"
	"zed-cli-win-unofficial/internal/utils"

	"github.com/bi-zone/go-fileversion"
	"github.com/hashicorp/go-version"
)

const MIN_ZED_VERSION string = "0.177.0"

// GetZedVersion retrieves the version of the Zed executable
func GetZedVersion(zedPath string) (*version.Version, error) {
	info, err := fileversion.New(zedPath)
	if err != nil {
		return nil, fmt.Errorf("could not get version info for %s: %w", zedPath, err)
	}

	v, err := version.NewVersion(info.FixedInfo().ProductVersion.String())
	if err != nil {
		return nil, fmt.Errorf("could not parse version string: %w", err)
	}
	return v, nil
}

// isZedRunning checks if the Zed is currently running.
func isZedRunning() (bool, error) {
	cmd := exec.Command("powershell", "-NoLogo", "-NoProfile", "-Command", "Get-Process Zed -ErrorAction SilentlyContinue")
	output, err := cmd.Output()

	if err == nil && len(output) > 0 {
		return true, nil
	}

	return false, nil
}

// LaunchZed launches zed with optional project path
func LaunchZed(zedPath string, projectPath string) error {
	var cmd *exec.Cmd
	isRunning, _ := isZedRunning()
	zedVersion, err := GetZedVersion(zedPath)

	if err != nil {
		telementry.TrackEvent("zed_version_error", map[string]any{
			"error_type": "version_read_failed",
		})
		utils.Warning("Could not determine Zed version: " + err.Error())
	} else {
		telementry.TrackEvent("zed_version_found", map[string]any{
			"zed_version": zedVersion.String(),
		})
		utils.Debugln(fmt.Sprintf("Current Zed version: %s", zedVersion.String()))
	}

	if isRunning {
		constraint, _ := version.NewConstraint("< " + MIN_ZED_VERSION)
		utils.Debugln(fmt.Sprintf("Constraint Checking is %t\n", constraint.Check(zedVersion)))

		// We only block if we could successfully get the version and it matches the constraint.
		if constraint.Check(zedVersion) {
			utils.PrintUpgradeRequiredBanner(MIN_ZED_VERSION)
			fmt.Printf("Your current Zed version: v%s\n", zedVersion.String())
			utils.Info("This CLI feature requires Zed v%s or newer when Zed is already running.\n", MIN_ZED_VERSION)
			utils.Info("Solutions:\n")
			utils.Info(" 1. Update Zed to the latest version (recommended)\n")
			utils.Info(" 2. Close the existing Zed window and try again\n")
			telementry.TrackEvent("zed_minimum_version_error", map[string]any{
				"error_type":       "version_too_old",
				"current_version":  zedVersion.String(),
				"required_version": MIN_ZED_VERSION,
			})
			return nil
		}
	}

	if projectPath != "" {
		if _, err := os.Stat(projectPath); os.IsNotExist(err) {
			if err := os.MkdirAll(projectPath, 0755); err != nil {
				telementry.TrackEvent("project_folder_creation_error", map[string]any{
					"error_type": "folder_creation_failed",
				})
				return fmt.Errorf("unable to create project folder: %w", err)
			}

			telementry.TrackEvent("project_folder_created", map[string]any{
				"created_new_folder": true,
			})
			utils.Error("Path doesn't exists")
			utils.Info("📁 Created new folder: %s\n", filepath.Clean(projectPath))
		}

		cmd = exec.Command(zedPath, projectPath)
	} else {
		cmd = exec.Command(zedPath)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start Zed: %w", err)
	}

	utils.Success("Zed opened successfully!!")
	telementry.TrackEvent("zed_launched", map[string]any{
		"has_project_path": projectPath != "",
		"zed_was_running":  isRunning,
		"zed_version":      zedVersion.String(),
	})
	return nil
}
