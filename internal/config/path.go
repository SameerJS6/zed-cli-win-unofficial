package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileExists: Checks if the given path to a file exists or not
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// resolvePath: resolves the given path to normal, if there's ENV present else just return the normal path
func resolvePath(path string) (string, error) {
	if !strings.HasPrefix(path, "%") && !strings.Contains(path, "%") {
		fmt.Println("✅ No Environment variable is used, using the path directly: ", path)
		return path, nil
	}

	end := strings.Index(path[1:], "%") + 1 // Adding 1 to balance out the index, since we are indexing after 1 element.
	envName := path[1:end]
	envValue := os.Getenv(envName)

	if envValue == "" {
		return "", fmt.Errorf("❌ Environment variable is not set or in-correct: %s", envName)
	}

	restOfPath := path[end+1:]
	resolvedPath := filepath.Join(envValue, restOfPath)
	fmt.Println("✅ Using resolved path: ", resolvedPath)
	return resolvedPath, nil
}

// ValidatePath: validates that a path exists and resolves environment variables
func ValidatePath(path string) (string, error) {
	resolvedPath, err := resolvePath(path)

	if err != nil {
		return "", fmt.Errorf("Failed to resolve path: %w", err)
	}

	if !FileExists(resolvedPath) {
		return "", fmt.Errorf("Given path doesn't result to any existing file: %s", resolvedPath)
	}

	return resolvedPath, nil
}
