package registry

import (
	"fmt"
	"path/filepath"
	"strings"
	"zed-cli-win-unofficial/internal/utils"

	"golang.org/x/sys/windows/registry"
)

// ensureKey creates or opens a registry key
func ensureKey(baseKey registry.Key, path string, access uint32) (registry.Key, bool, error) {
	key, alreadyExists, err := registry.CreateKey(baseKey, path, access)

	if err != nil {
		return registry.Key(0), false, fmt.Errorf("unable to access registry entry: %w", err)
	}

	return key, alreadyExists, nil
}

// setStringValue sets a string value in the registry
func setStringValue(key registry.Key, name string, value string) error {
	err := key.SetStringValue(name, value)

	if err != nil {
		return fmt.Errorf("unable to save registry value: %w", err)
	}

	return nil
}

// CreateProgID creates a ProgID registry entry for a file extension
func CreateProgID(registryConfig *RegistryConfig, ext string) error {
	// 1. Create root level ProgID (for eg: Zed.json)
	progID := fmt.Sprintf("%s%s", registryConfig.AppName, ext)
	progPath := filepath.Join("Software", "Classes", progID)

	key, _, err := ensureKey(registry.CURRENT_USER, progPath, registry.WRITE)
	if err != nil {
		return err
	}
	defer key.Close()

	fileTypeDescription := fmt.Sprintf(registryConfig.PerFileTypeDescriptionText, strings.ToUpper(strings.TrimPrefix(ext, ".")))

	if err := setStringValue(key, "", fileTypeDescription); err != nil {
		return fmt.Errorf("failed to register %s file type: %w", ext, err)
	}

	if err := setStringValue(key, "AppUserModelID", registryConfig.AppUserModelId); err != nil {
		return fmt.Errorf("failed to configure %s file type: %w", ext, err)
	}

	// 2. Add DefaultIcon Key with its value
	defaultIconPath := filepath.Join(progPath, "DefaultIcon")
	defaultIconKey, _, err := ensureKey(registry.CURRENT_USER, defaultIconPath, registry.WRITE)

	if err != nil {
		return fmt.Errorf("failed to set icon for %s files: %w", ext, err)

	}
	defer defaultIconKey.Close()

	defaultIconValue := fmt.Sprintf(`"%s"`, registryConfig.ExecutablePath)
	if err := setStringValue(defaultIconKey, "", defaultIconValue); err != nil {
		return fmt.Errorf("failed to set icon for %s files: %w", ext, err)
	}

	// 3. Adding Shell > Open Key with Icon key/value entry
	openKeyPath := filepath.Join(progPath, "shell", "open")
	openKey, _, err := ensureKey(registry.CURRENT_USER, openKeyPath, registry.WRITE)
	if err != nil {
		return fmt.Errorf("failed to configure %s file opening: %w", ext, err)
	}
	defer openKey.Close()

	if err := setStringValue(openKey, "Icon", defaultIconValue); err != nil {
		return fmt.Errorf("failed to set icon for %s files: %w", ext, err)
	}

	// 4. Adding Shell > Open > Command entry with DefaultValue of exe path
	commandKeyPath := filepath.Join(openKeyPath, "command")
	commandKey, _, err := ensureKey(registry.CURRENT_USER, commandKeyPath, registry.WRITE)

	if err != nil {
		return fmt.Errorf("failed to configure %s file opening: %w", ext, err)
	}

	defer commandKey.Close()
	commandKeyValue := fmt.Sprintf(`"%s" "%%1"`, registryConfig.ExecutablePath)
	if err := setStringValue(commandKey, "", commandKeyValue); err != nil {
		return fmt.Errorf("failed to configure %s file opening: %w", ext, err)
	}

	return nil
}

// AssociateExtensionWithProgID: associates a file extension with its ProgID
func AssociateExtensionWithProgID(ext string, progID string) error {
	extKeyPath := filepath.Join("Software", "Classes", ext, "OpenWithProgids")
	extKey, _, err := ensureKey(registry.CURRENT_USER, extKeyPath, registry.WRITE)

	if err != nil {
		return fmt.Errorf("failed to access %s file type settings: %w", ext, err)
	}

	defer extKey.Close()

	if err := setStringValue(extKey, progID, ""); err != nil {
		return fmt.Errorf("failed to associate %s files with Zed: %w", ext, err)
	}

	utils.Debug("File type %s associated with Zed\n", ext)
	return nil
}

// DeleteKeyRecursivly deletes a registry key and all its subkey
func DeleteKeyRecursively(baseKey registry.Key, path string) {
	// Step 1: Open the Key
	key, err := registry.OpenKey(baseKey, path,
		registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)

	if err != nil {
		if err == registry.ErrNotExist {
			utils.Debugln("Registry entry not found (already removed)")
		} else {
			utils.Debugln("Unable to access registry entry")
		}

		errDelete := registry.DeleteKey(baseKey, path)
		if errDelete != nil && errDelete != registry.ErrNotExist {
			utils.Debugln("Failed to remove registry entry")
		} else if errDelete == nil {
			utils.Debugln("Registry entry removed")
		}
		return
	}

	defer key.Close()

	subKeyNames, err := key.ReadSubKeyNames(0)
	if err != nil {
		utils.Debugln("Unable to read registry subentries")
	}

	for _, subKeyName := range subKeyNames {
		fullSubKeyPath := filepath.Join(path, subKeyName)
		DeleteKeyRecursively(baseKey, fullSubKeyPath)
	}

	key.Close()
	deleteErr := registry.DeleteKey(baseKey, path)

	if deleteErr != nil {
		if deleteErr == registry.ErrNotExist {
			utils.Debugln("Registry entry already removed")
		} else {
			utils.Debugln("Failed to remove registry entry")
		}
	} else {
		utils.Debugln("Registry entry removed successfully")
	}
}

// DeleteValueSilently deletes a registry value without throwing errors
func DeleteValueSilently(baseKey registry.Key, keyPath string, valueName string) {
	key, err := registry.OpenKey(baseKey, keyPath, registry.WRITE)

	if err != nil {
		if err == registry.ErrNotExist {
			utils.Debugln("Registry entry not found")
		} else {
			utils.Debugln("Unable to access registry entry")
		}
		return
	}

	defer key.Close()

	key.DeleteValue(valueName)
}
