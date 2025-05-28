package registry

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// InstallGenericContextMenu installs the generic "Open with Zed" context menu entries
func InstallGenericContextMenu(config *RegistryConfig) error {
	// 1. All files context menu (*/shell/Zed)
	if err := createContextMenuEntry("*", config); err != nil {
		return fmt.Errorf("failed to create context menu for all files: %w", err)
	}

	// 2. Directory context menu (Directory/shell/Zed)
	if err := createContextMenuEntry("Directory", config); err != nil {
		return fmt.Errorf("failed to create context menu for directories: %w", err)
	}

	// 3. Directory background context menu (Directory/Background/shell/Zed)
	if err := createDirectoryBackgroundContextMenu(config); err != nil {
		return fmt.Errorf("failed to create directory background context menu: %w", err)
	}

	return nil
}

// createContextMenuEntry creates a context menu entry for a given file type
func createContextMenuEntry(fileType string, config *RegistryConfig) error {
	shellKeyPath := filepath.Join("Software", "Classes", fileType, "shell", config.AppName)

	// Create the shell key
	shellKey, _, err := ensureKey(registry.CURRENT_USER, shellKeyPath, registry.WRITE)
	if err != nil {
		return fmt.Errorf("shell key for %s: %w", fileType, err)
	}
	defer shellKey.Close()

	if err := setStringValue(shellKey, "", config.GenericMenuText); err != nil {
		return fmt.Errorf("shell key text for %s: %w", fileType, err)
	}

	iconPath := fmt.Sprintf(`"%s"`, config.ExecutablePath)
	if err := setStringValue(shellKey, "Icon", iconPath); err != nil {
		log.Printf("Warning: failed to set icon for %s: %v", fileType, err)
	}

	// Create the command subkey
	commandKeyPath := filepath.Join(shellKeyPath, "command")
	commandKey, _, err := ensureKey(registry.CURRENT_USER, commandKeyPath, registry.WRITE)
	if err != nil {
		return fmt.Errorf("command key for %s: %w", fileType, err)
	}
	defer commandKey.Close()

	commandValue := fmt.Sprintf(`"%s" "%%1"`, config.ExecutablePath)
	if err := setStringValue(commandKey, "", commandValue); err != nil {
		return fmt.Errorf("command value for %s: %w", fileType, err)
	}

	return nil
}

// createDirectoryBackgroundContextMenu creates context menu for directory background
func createDirectoryBackgroundContextMenu(config *RegistryConfig) error {
	shellKeyPath := filepath.Join("Software", "Classes", "Directory", "Background", "shell", config.AppName)

	// Create the shell key
	shellKey, _, err := ensureKey(registry.CURRENT_USER, shellKeyPath, registry.WRITE)
	if err != nil {
		return fmt.Errorf("directory background shell key: %w", err)
	}
	defer shellKey.Close()

	if err := setStringValue(shellKey, "", config.GenericMenuText); err != nil {
		return fmt.Errorf("directory background shell key text: %w", err)
	}

	iconPath := fmt.Sprintf(`"%s"`, config.ExecutablePath)
	if err := setStringValue(shellKey, "Icon", iconPath); err != nil {
		fmt.Printf("Warning: failed to set icon for directory background: %v\n", err)
	}

	// Create the command subkey
	commandKeyPath := filepath.Join(shellKeyPath, "command")
	commandKey, _, err := ensureKey(registry.CURRENT_USER, commandKeyPath, registry.WRITE)
	if err != nil {
		return fmt.Errorf("directory background command key: %w", err)
	}
	defer commandKey.Close()

	// Set the command - for directory background, use %V% which represents the current directory
	commandValue := fmt.Sprintf(`"%s" "%%V"`, config.ExecutablePath)
	if err := setStringValue(commandKey, "", commandValue); err != nil {
		return fmt.Errorf("directory background command value: %w", err)
	}

	return nil
}

// UninstallAllContextMenus removes all Zed context menu entries
func UninstallAllContextMenus(config *RegistryConfig) error {
	// Remove all file types context menu
	DeleteKeyRecursively(registry.CURRENT_USER, filepath.Join("Software", "Classes", "*", "shell", config.AppName))

	// Remove directory context menu
	DeleteKeyRecursively(registry.CURRENT_USER, filepath.Join("Software", "Classes", "Directory", "shell", config.AppName))

	// Remove directory background context menu
	DeleteKeyRecursively(registry.CURRENT_USER, filepath.Join("Software", "Classes", "Directory", "Background", "shell", config.AppName))

	// Remove ProgIDs for each file extension
	for _, ext := range config.FileExtensions {
		if !strings.HasPrefix(ext, ".") {
			continue
		}
		progID := fmt.Sprintf("%s%s", config.AppName, ext)
		DeleteKeyRecursively(registry.CURRENT_USER, filepath.Join("Software", "Classes", progID))
		DeleteValueSilently(registry.CURRENT_USER, filepath.Join("Software", "Classes", ext), "")
	}

	return nil
}
