package registry

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// ensureKey creates or opens a registry key
func ensureKey(baseKey registry.Key, path string, access uint32) (registry.Key, bool, error) {
	key, alreadyExists, err := registry.CreateKey(baseKey, path, access)

	if err != nil {
		return registry.Key(0), false, fmt.Errorf("failed to create/open key %s: %w", path, err)
	}

	return key, alreadyExists, nil
}

// setStringValue sets a string value in the registry
func setStringValue(key registry.Key, name string, value string) error {
	err := key.SetStringValue(name, value)

	if err != nil {
		return fmt.Errorf("failed to set string value for %s: %w", name, err)
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
		return fmt.Errorf("ProgId %s: %w", progID, err)
	}

	if err := setStringValue(key, "AppUserModelID", registryConfig.AppUserModelId); err != nil {
		return fmt.Errorf("ProgID %s: failed to set AppUserModelID: %v", progID, err)
	}

	// 2. Add DefaultIcon Key with its value
	defaultIconPath := filepath.Join(progPath, "DefaultIcon")
	defaultIconKey, _, err := ensureKey(registry.CURRENT_USER, defaultIconPath, registry.WRITE)

	if err != nil {
		return fmt.Errorf("ProgID %s DefaultIcon: %w", progID, err)

	}
	defer defaultIconKey.Close()

	defaultIconValue := fmt.Sprintf(`"%s"`, registryConfig.ExecutablePath)
	if err := setStringValue(defaultIconKey, "", defaultIconValue); err != nil {
		return fmt.Errorf("ProgID %s DefaultIcon value: %w", progID, err)
	}

	// 3. Adding Shell > Open Key with Icon key/value entry
	openKeyPath := filepath.Join(progPath, "shell", "open")
	openKey, _, err := ensureKey(registry.CURRENT_USER, openKeyPath, registry.WRITE)
	if err != nil {
		return fmt.Errorf("ProgID %s open command: %w", progID, err)
	}
	defer openKey.Close()

	if err := setStringValue(openKey, "Icon", defaultIconValue); err != nil {
		return fmt.Errorf("ProgID %s Icon value: %w", progID, err)
	}

	// 4. Adding Shell > Open > Command entry with DefaultValue of exe path
	commandKeyPath := filepath.Join(openKeyPath, "command")
	commandKey, _, err := ensureKey(registry.CURRENT_USER, commandKeyPath, registry.WRITE)

	if err != nil {
		return fmt.Errorf("ProgID %s shell command: %w", progID, err)
	}

	defer commandKey.Close()
	commandKeyValue := fmt.Sprintf(`"%s" "%%1"`, registryConfig.ExecutablePath)
	if err := setStringValue(commandKey, "", commandKeyValue); err != nil {
		return fmt.Errorf("ProgID %s shell command: %w", progID, err)
	}

	return nil
}

// AssociateExtensionWithProgID: associates a file extension with its ProgID
func AssociateExtensionWithProgID(ext string, progID string) error {
	extKeyPath := filepath.Join("Software", "Classes", ext, "OpenWithProgids")
	extKey, _, err := ensureKey(registry.CURRENT_USER, extKeyPath, registry.WRITE)

	if err != nil {
		return fmt.Errorf("extension %s: %w", ext, err)
	}

	defer extKey.Close()

	if err := setStringValue(extKey, progID, ""); err != nil {
		return fmt.Errorf("extension %s association: %w", ext, err)
	}

	fmt.Printf("Successfully associated extension %s with ProgID %s\n", ext, progID)
	return nil
}

// DeleteKeyRecursivly deletes a registry key and all its subkey
func DeleteKeyRecursively(baseKey registry.Key, path string) {
	// Step 1: Open the Key
	key, err := registry.OpenKey(baseKey, path,
		registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)

	if err != nil {
		if err == registry.ErrNotExist {
			fmt.Printf("Key not found for recursive deletion (already removed or never existed): %s\n", path)
		} else {
			fmt.Printf("Failed to open key %s for reading subkeys: %v\n", path, err)
		}

		errDelete := registry.DeleteKey(baseKey, path)
		if errDelete != nil && errDelete != registry.ErrNotExist {
			fmt.Printf("⚠️ Failed to delete key %s (after failing to open for subkey enumeration): %v\n", path, errDelete)
		} else if errDelete == nil {
			fmt.Printf("✅  Successfully deleted key: %s\n", path)
		}
		return
	}

	defer key.Close()

	subKeyNames, err := key.ReadSubKeyNames(0)
	if err != nil {
		fmt.Printf("Failed to read subkey names for %s: %v\n", path, err)
	}

	for _, subKeyName := range subKeyNames {
		fullSubKeyPath := filepath.Join(path, subKeyName)
		DeleteKeyRecursively(baseKey, fullSubKeyPath)
	}

	key.Close()
	deleteErr := registry.DeleteKey(baseKey, path)

	if deleteErr != nil {
		if deleteErr == registry.ErrNotExist {
			fmt.Printf("ℹ️ Key not found for final deletion (already removed): %s\n", path)
		} else {
			fmt.Printf("⚠️ Failed to delete key %s (after attempting subkey deletion): %v\n", path, err)
		}
	} else {
		fmt.Printf("✅ Successfully deleted key: %s", path)
	}
}

// DeleteValueSilently deletes a registry value without throwing errors
func DeleteValueSilently(baseKey registry.Key, keyPath string, valueName string) {
	key, err := registry.OpenKey(baseKey, keyPath, registry.WRITE)

	if err != nil {
		if err == registry.ErrNotExist {
			fmt.Printf("ℹ️ Key not found for value deletion: %s\n", keyPath)
		} else {
			fmt.Printf("Failed to open key %s to deelte value '%s': %v", keyPath, valueName, err)
		}
		return
	}

	defer key.Close()

	key.DeleteValue(valueName)
}
