package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

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
	isRunning, err := isZedRunning()

	if err != nil {
		return fmt.Errorf("unable to check if Zed is running: %w", err)
	}

	if isRunning {
		fmt.Println("❌ Zed is already running in another instance!!")
		fmt.Println("⚠️  This CLI cannot launch a second instance due to Zed's limitation")
		return nil
	}

	if projectPath != "" {
		if _, err := os.Stat(projectPath); os.IsNotExist(err) {
			if err := os.MkdirAll(projectPath, 0755); err != nil {
				return fmt.Errorf("unable to create project folder: %w", err)
			}

			fmt.Println("❌  Path doesn't exists")
			fmt.Printf("📁  Created new folder: %s", filepath.Clean(projectPath))
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

	fmt.Println("✅  Zed opened successfully!!")
	return nil
}
