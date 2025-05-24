package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"golang.org/x/sys/windows/registry"
)

func isZedRunning() (bool, error) {
	var cmd *exec.Cmd = exec.Command("powershell", "-NoLogo", "-NoProfile", "-Command", "Get-Process Zed -ErrorAction SilentlyContinue")

	output, err := cmd.Output()
	// outputStr := strings.TrimSpace(string(output))
	// fmt.Printf("Raw output from PowerShell (trimmed):\n'%s'\n", outputStr)
	fmt.Println("Output from PowerShell Command:", output)

	// if err != nil {
	// 	return false, fmt.Errorf("powershell command failed: %w", err)
	// }

	// if strings.EqualFold(outputStr, "Zed") {
	// 	return true, nil
	// }

	// if strings.Contains(strings.ToLower(outputStr), "zed") && !strings.Contains(strings.ToLower(outputStr), "processname") {
	// 	return true, nil
	// }

	if err == nil && len(output) > 0 {
		return true, nil
	}

	return false, nil
	// procs, err := process.Processes()

	// if err != nil {
	// 	return false, err
	// }

	// for _, p := range procs {
	// 	name, err := p.Name()
	// 	if err == nil && strings.EqualFold(name, "zed.exe") {
	// 		return true, nil
	// 	}
	// }

	// return false, nil
}

const CLI_CONFIG_PATH string = `%APPDATA%\ZedUNOFFICIALCLI\config.json`

type Config struct {
	ZedPath string `json:"zedPath"`
}

func saveConfig(path string) error {
	appData := os.Getenv("APPDATA")
	fullAppDataPath := filepath.Join(appData, `Zed-Unofficial-Cli`)

	if err := os.MkdirAll(fullAppDataPath, 0755); err != nil {
		return err
	}

	config := Config{ZedPath: path}
	configPath := filepath.Join(fullAppDataPath, `config.json`)

	file, err := os.Create(configPath)

	if err != nil {
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return err
	}

	fmt.Println("✅ Config file saved at: ", configPath)
	return nil
}

func loadConfig() (Config, error) {
	appData := os.Getenv("APPDATA")
	fullAppDataPath := filepath.Join(appData, "Zed-Unofficial-Cli", "config.json")

	if _, err := os.Stat(fullAppDataPath); os.IsNotExist(err) {
		fmt.Println("❌ Config file does not exist at: ", fullAppDataPath)
		fmt.Println("👉 Tip: Run `zed --config-path <path>` to create a config file.")
		return Config{}, err
	}

	file, err := os.Open(fullAppDataPath)
	if err != nil {
		return Config{}, err
	}

	defer file.Close()
	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func resolvePath(path string) (string, error) {
	if strings.HasPrefix(path, "%") && strings.Contains(path, "%") {
		end := strings.Index(path[1:], "%") + 1
		envName := path[1:end]
		envValue := os.Getenv(envName)

		if envValue == "" {
			fmt.Println("❌ Environment variable is not set or in correct.", envName)
			return "", fmt.Errorf("environment variable is not set or in correct: %s", envName)
		}

		restOfPath := path[end+1:]
		resolvedPath := filepath.Join(envValue, restOfPath)
		fmt.Println("✅ Using resolved path: ", resolvedPath)
		return resolvedPath, nil
	}

	fmt.Println("✅ No Environment variable is used, using the path directly: ", path)
	return path, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type RegistryConfig struct {
	AppName             string   // e.g Zed
	ExecutablePath      string   // Full path to zed.exe
	AppUserModelID      string   // e.g In our case, its same as executable path
	GenericMenuText     string   // e.g Open w%th Zed
	FileExtensions      []string // e.g .txt, .md, .rs, etc.
	PerFileTypeDescTmpl string   // "%s File (Zea)
}

func ensureKey(baseKey registry.Key, path string, access uint32) (registry.Key, bool, error) {
	key, created, err := registry.CreateKey(baseKey, path, access)
	if err != nil {
		return registry.Key(0), false, fmt.Errorf("failed to create/open key %s: %w", path, err)
	}

	return key, created, nil
}

func setStringValue(key registry.Key, name string, value string) error {
	err := key.SetStringValue(name, value)

	if err != nil {
		return fmt.Errorf("failed to set string value for %s: %w", name, value, err)
	}

	return nil
}

// Creates root level key for file extensions e.g. Zed.json
func registryProgId(registryConfig *RegistryConfig, ext string) error { // ext -> e.g .txt, .md, .rs, etc.
	progId := fmt.Sprintf("%s%s", registryConfig.AppName, ext) // Zed.json
	progIdPath := filepath.Join("Software", "Classes", progId)
	const DEFAULT_KEY_VALUE string = ""

	// 1. Create/Open the ProgID key  (eg: HKEY_CURRENT_USER\Software\Classes\Zed.json)
	key, _, err := ensureKey(registry.CURRENT_USER, progIdPath, registry.WRITE)
	if err != nil {
		return err
	}

	defer key.Close()

	// 2. Set the (Default) value:  file type description
	fileTypeDescription := fmt.Sprintf(registryConfig.PerFileTypeDescTmpl, strings.ToUpper(strings.TrimPrefix(ext, ".")))
	stringValueErr := setStringValue(key, "", fileTypeDescription)

	if stringValueErr != nil {
		return fmt.Errorf("ProgId %s: %w", progId, stringValueErr)
	}

	// 3. Add AppUserModelID key (Optional)
	if registryConfig.AppUserModelID != "" {
		if err := setStringValue(key, "AppUserModelID", registryConfig.AppUserModelID); err != nil {
			log.Printf("Warning: ProgID %s: failed to set AppUserModelID: %v", progId, err)
		}
	}

	// 4. Create DefaultIcon Subkey
	defaultIconPath := filepath.Join(progIdPath, "DefaultIcon")
	iconKey, _, iconKeyErr := ensureKey(registry.CURRENT_USER, defaultIconPath, registry.WRITE)
	if iconKeyErr != nil {
		return fmt.Errorf("ProgID %s DefaultIcon: %w", progId, iconKeyErr)
	}

	defer iconKey.Close()

	iconValue := DEFAULT_KEY_VALUE // TODO: Add the correct icon path here.
	if err := setStringValue(iconKey, "", iconValue); err != nil {
		return fmt.Errorf("ProgID %s DefaultIcon value: %w", progId, err)
	}

	// 5. Create shell subkey
	shellPath := filepath.Join(progIdPath, "shell")
	shellKey, _, shellKeyErr := ensureKey(registry.CURRENT_USER, shellPath, registry.WRITE)

	if shellKeyErr != nil {
		return fmt.Errorf("ProgID %s shell %w", progId, shellKeyErr)
	}

	defer shellKey.Close()

	if err := setStringValue(shellKey, "", DEFAULT_KEY_VALUE); err != nil {
		return fmt.Errorf("ProgID %s shell: %w", progId, err)
	}

	// 6. Create open subkey under shell
	openPath := filepath.Join(shellPath, "open")
	openKey, _, openKeyErr := ensureKey(registry.CURRENT_USER, openPath, registry.WRITE)

	if openKeyErr != nil {
		return fmt.Errorf("ProgID %s open: %w", progId, openKeyErr)
	}

	defer openKey.Close()

	if err := setStringValue(openKey, "", DEFAULT_KEY_VALUE); err != nil {
		return fmt.Errorf("ProgID %s open: %w", progId, err)
	}

	if err := setStringValue(openKey, "Icon", registryConfig.ExecutablePath); err != nil {
		return fmt.Errorf("ProgID %s open Icon: %w", progId, err)
	}

	// 7. Create command subkey under open
	commandPath := filepath.Join(openPath, "command")
	commandKey, _, commandKeyErr := ensureKey(registry.CURRENT_USER, commandPath, registry.WRITE)

	if commandKeyErr != nil {
		return fmt.Errorf("ProgID %s command: %w", progId, commandKeyErr)
	}

	defer commandKey.Close()

	commandValue := fmt.Sprintf("\"%s\" \"%%1\"", registryConfig.ExecutablePath)
	if err := setStringValue(commandKey, "", commandValue); err != nil {
		return fmt.Errorf("ProgID %s command: %w", progId, err)
	}

	log.Printf("Successfully registered ProgID: %s", progId)
	return nil
}

// associateExtensionWithProgID -> This function will associate a file extension with a ProgID eg:- .json -> Zed.json
func associateExtensionWithProgID(registryConfig *RegistryConfig, ext string, progID string) error {

	// Add ProgID to OpenWithProgids subkey
	extPath := filepath.Join("Software", "Classes", ext, "OpenWithProgids")

	key, _, err := ensureKey(registry.CURRENT_USER, extPath, registry.WRITE)
	if err != nil {
		return fmt.Errorf("association for %s OpenWithProgids key: %w", ext, err)
	}

	defer key.Close()

	// prodIdKey := fmt.Sprintf("%s%s", registryConfig.AppName, ext)
	if err := setStringValue(key, progID, ""); err != nil {
		return fmt.Errorf("failed to set string value for %s: %w", progID, err)
	}

	log.Printf("Successfully associated extension %s with ProgID %s", ext, progID)
	return nil

}

func registryGenericContextMenu(config *RegistryConfig) error {
	basePaths := map[string]string{
		"all_files":            filepath.Join("Software", "Classes", "*", "shell"),
		"directory":            filepath.Join("Software", "Classes", "Directory", "shell"),
		"directory_background": filepath.Join("Software", "Classes", "Directory", "Background", "shell"),
	}

	contextMenuKeyName := config.AppName + "ContextMenuHandler"

	for contextType, basePath := range basePaths {
		menuPath := filepath.Join(basePath, contextMenuKeyName)

		key, _, err := ensureKey(registry.CURRENT_USER, menuPath, registry.WRITE)
		if err != nil {
			return fmt.Errorf("generic menu %s: %w", menuPath, err)
		}

		if err := setStringValue(key, "", config.GenericMenuText); err != nil {
			key.Close()
			return fmt.Errorf("generic menu %s (Default) text: %w", contextType, err)
		}

		if err := setStringValue(key, "Icon", config.ExecutablePath); err != nil {
			log.Printf("Warning: generic menu %s: failed to set Icon: %v", contextType, err)
		}

		commandPath := filepath.Join(menuPath, "command")
		commandKey, _, commandKeyErr := ensureKey(registry.CURRENT_USER, commandPath, registry.WRITE)

		if commandKeyErr != nil {
			return fmt.Errorf("generic menu %s command: %w", contextType, commandKeyErr)
		}

		var commandValue string
		if contextType == "all_files" {
			commandValue = fmt.Sprintf("\"%s\" \"%%1\"", config.ExecutablePath)
		} else {
			commandValue = fmt.Sprintf("\"%s\" \"%%V\"", config.ExecutablePath)
		}

		if err := setStringValue(commandKey, "", commandValue); err != nil {
			commandKey.Close()
			return fmt.Errorf("generic menu %s command: %w", contextType, err)
		}

		commandKey.Close()

		log.Printf("Successfully registered generic context menu for: %s", contextType)
	}

	return nil
}

// Deleting Keys from Registry
// deleteRegistryKeyRecursively attempts to delete a key and all its subkeys.
// It's designed for cleanup, so errors are logged but don't stop the overall process.
func deleteRegistryKeyRecursively(baseKey registry.Key, path string) {
	// Open the key to read its subkeys
	key, err := registry.OpenKey(baseKey, path, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			// Key doesn't exist, nothing to do
			// log.Printf("ℹ️ Key not found for recursive deletion (already removed or never existed): %s", path)
		} else {
			log.Printf("⚠️ Failed to open key %s for reading subkeys: %v", path, err)
		}
		// Still attempt to delete the key itself, in case it's an empty key that couldn't be opened for enumeration
		// or if the error was something other than NotExist.
		errDelete := registry.DeleteKey(baseKey, path)
		if errDelete != nil && errDelete != registry.ErrNotExist {
			log.Printf("⚠️ Failed to delete key %s (after failing to open for subkey enumeration): %v", path, errDelete)
		} else if errDelete == nil {
			log.Printf("✅ Successfully deleted key: %s (was likely empty or became empty)", path)
		}
		return
	}
	defer key.Close()

	// Enumerate and recursively delete all subkeys
	subKeyNames, err := key.ReadSubKeyNames(0) // 0 means read all
	if err == nil {
		for _, subKeyName := range subKeyNames {
			fullSubKeyPath := filepath.Join(path, subKeyName)
			deleteRegistryKeyRecursively(baseKey, fullSubKeyPath) // Recursive call
		}
	} else {
		log.Printf("⚠️ Failed to read subkey names for %s: %v", path, err)
		// Continue to attempt to delete the current key anyway
	}

	// After all subkeys are (attempted to be) deleted, try to delete the current key.
	// We need to close the key handle before deleting it.
	key.Close() // Explicitly close before deletion attempt
	err = registry.DeleteKey(baseKey, path)
	if err != nil {
		if err == registry.ErrNotExist {
			// log.Printf("ℹ️ Key not found for final deletion (already removed): %s", path)
		} else {
			// This might happen if subkey deletion failed and the key is not empty,
			// or due to permission issues not caught earlier.
			log.Printf("⚠️ Failed to delete key %s (after attempting subkey deletion): %v", path, err)
		}
	} else {
		log.Printf("✅ Successfully deleted key: %s", path)
	}
}

// deleteRegistryValueQuietly attempts to delete a specific value from a key.
func deleteRegistryValueQuietly(baseKey registry.Key, keyPath string, valueName string) {
	key, err := registry.OpenKey(baseKey, keyPath, registry.WRITE)
	if err != nil {
		if err == registry.ErrNotExist {
			// log.Printf("ℹ️ Key not found for value deletion: %s", keyPath)
		} else {
			log.Printf("⚠️ Failed to open key %s to delete value '%s': %v", keyPath, valueName, err)
		}
		return
	}
	defer key.Close()

	err = key.DeleteValue(valueName)
	if err != nil {
		if err == registry.ErrNotExist { // Or the specific error for value not existing
			// log.Printf("ℹ️ Value '%s' not found in key %s (already removed or never existed)", valueName, keyPath)
		} else {
			log.Printf("⚠️ Failed to delete value '%s' from key %s: %v", valueName, keyPath, err)
		}
	} else {
		log.Printf("✅ Successfully deleted value '%s' from key: %s", valueName, keyPath)
	}
}

func main() {
	cmd := &cli.Command{
		Name:  "zed",
		Usage: "Zed's Unofficial Windows CLI",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:             "version",
				Aliases:          []string{"v"},
				ValidateDefaults: false,
				Usage:            "Print the version of Zed CLI",
				Action: func(ctx context.Context, cmd *cli.Command, b bool) error {
					if b {
						fmt.Println("v1.0.0")
						return nil
					}
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "Configure the CLI's Path & Setting",
				Commands: []*cli.Command{
					{
						Name:  "set",
						Usage: "Set the path to the Zed executable",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							path := cmd.Args().First()

							if path == "" {
								fmt.Println("❌ No path provided.")
								return nil
							}

							resolvedPath, resolvedPathErr := resolvePath(path)

							if !fileExists(resolvedPath) || resolvedPathErr != nil {
								fmt.Println("❌ Provided path does not exist.")
								return nil
							}

							err := saveConfig(resolvedPath)

							if err != nil {
								fmt.Println("Error occured during saving the path: ", err)
								return nil
							}

							fmt.Println("Getting the path to the Zed executable: ", cmd.Args().First())
							fmt.Println("You might want to update the context menu and file association for the current user.")
							fmt.Println("You can do that by running: zed context install")
							return nil
						},
					},
					{
						Name:  "get",
						Usage: "Get the current path to the Zed executable",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							loadedPath, loadedPathErr := loadConfig()

							if loadedPathErr != nil || !fileExists(loadedPath.ZedPath) {
								fmt.Println("Error occured during loading configured path: ", loadedPathErr)
								return nil
							}

							fmt.Println("✅ Zed is configured to run from: ", loadedPath.ZedPath)
							return nil
						}},
				},
			},
			{
				Name:  "context",
				Usage: "Configure the 'Open with Zed' in context menu",
				Commands: []*cli.Command{
					{
						Name:  "install",
						Usage: "Install the 'Open with Zed' Context menu option",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							// Loading the correct executable path
							loadedPath, loadedPathErr := loadConfig()

							if loadedPathErr != nil || !fileExists(loadedPath.ZedPath) {
								fmt.Println("Error occured during loading configured path: ", loadedPathErr)
								return nil
							}

							// Setting up the registry config struct with correct data
							registryConfig := &RegistryConfig{
								AppName:         "Zed",
								ExecutablePath:  loadedPath.ZedPath,
								AppUserModelID:  loadedPath.ZedPath,
								GenericMenuText: "Open w&ith Zed",
								FileExtensions: []string{
									".asp",
									".aspx",
									".bash",
									".bash_login",
									".bash_logout",
									".bash_profile",
									".bashrc",
									".bib",
									".bowerrc",
									".c",
									".c",
									".cc",
									".cfg",
									".cjs",
									".clj",
									".cljs",
									".cljx",
									".clojure",
									".cls",
									".cmake",
									".code",
									".code",
									".coffee",
									".config",
									".containerfile",
									".cpp",
									".cS",
									".cshtml",
									".csproj",
									".css",
									".csV",
									".csx",
									".ctp",
									".cxx",
									".dart",
									".diff",
									".dockerfile",
									".dot",
									".dtd",
									".editorconfig",
									".edn",
									".erb",
									".eyaml",
									".eyml",
									".fs",
									".fsi",
									".fsscript",
									".fsx",
									".gemspec",
									".gitattributes",
									".gitconfig",
									".gitignore",
									".go",
									".gradle",
									".groovy",
									".h",
									".h",
									".handlebars",
									".hbs",
									".hh",
									".hpp",
									".htm",
									".html",
									".hxx",
									".ini",
									".ipynb",
									".jade",
									".jav",
									".java",
									".js",
									".jscsrc",
									".jshintrc",
									".jshtm",
									".json",
									".jsp",
									".jsX",
									".less",
									".log",
									".lua",
									".m",
									".makefile",
									".markdown",
									".md",
									".mdoc",
									".mdown",
									".mdtext",
									".mdtxt",
									".mdwn",
									".mjs",
									".mk",
									".mkd",
									".mkdn",
									".ml",
									".mli",
									".npmignore",
									".php",
									".phtml",
									".pl",
									".pl6",
									".plist",
									".pm",
									".pm6",
									".pod",
									".pp",
									".profile",
									".properties",
									".ps1",
									".psd1",
									".psgi",
									".psm1",
									".py",
									".pyi",
									".r",
									".rb",
									".rhistory",
									".rprofile",
									".rs",
									".rst",
									".rt",
									".sass",
									".scss",
									".sh",
									".shtml",
									".sql",
									".svg",
									".svgz",
									".t",
									".tex",
									".toml",
									".ts",
									".tsx",
									".txt",
									".vb",
									".vue",
									".wxi",
									".wxl",
									".wxs",
									".xaml",
									".xhtml",
									".xml",
									".yaml",
									".yml",
									".zsh"},
								PerFileTypeDescTmpl: "%s File (Zed)",
							}

							fmt.Printf("Starting Zed context menu and file association setup...")

							// Registering the generic context menu (*/shell/Zed/command, Directory/shell, Directory/Background )
							if err := registryGenericContextMenu(registryConfig); err != nil {
								log.Fatalf("Error registering generic context menus: %v", err)
							}

							// Running a loop to registry ProgId for each file extension (Zed.json, Zed.md, Zed.txt)
							for _, ext := range registryConfig.FileExtensions {
								// Checking if the file extension is valid or not
								if !strings.HasPrefix(ext, ".") {
									log.Printf("Skipping non-extension file type: %s (must start with '.')", ext)
									return nil
								}

								// Creating the ProgId (key) for each file extension (eg: Zed.json)
								progId := fmt.Sprintf("%s%s", registryConfig.AppName, ext)

								// Creating/Registering the ProgId for each file extension (eg: Zed.json)
								if err := registryProgId(registryConfig, ext); err != nil {
									log.Printf("Error registering ProgID for %s (%s): %v. Skipping association.", ext, progId, err)
									return nil
								}

								// Registering/Associating the registered ProgId with the file extension (eg: Zed.json with .json)
								if err := associateExtensionWithProgID(registryConfig, ext, progId); err != nil {
									log.Printf("Error associating extension %s with ProgID %s: %v", ext, progId, err)
									return nil
								}
							}

							fmt.Println("--------------------------------------------------------------------")
							fmt.Println("Zea context menu and file association setup complete for current user.")
							fmt.Println("You might need to restart Windows Explorer or log out/in for all changes to take full effect.")
							fmt.Println("To remove these entries, you would need to delete the created registry keys.")
							fmt.Println("--------------------------------------------------------------------")
							// TODO: Come tomorrow and finish both of these registry stuff, now that I've figured it out entirely.
							return nil
						},
					},
					{
						Name:  "uninstall",
						Usage: "Uninstall the 'Open with Zed' Context menu option",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							// Setting up the registry config struct with correct data
							config := &RegistryConfig{
								AppName: "Zed",
								FileExtensions: []string{
									".asp",
									".aspx",
									".bash",
									".bash_login",
									".bash_logout",
									".bash_profile",
									".bashrc",
									".bib",
									".bowerrc",
									".c",
									".c",
									".cc",
									".cfg",
									".cjs",
									".clj",
									".cljs",
									".cljx",
									".clojure",
									".cls",
									".cmake",
									".code",
									".code",
									".coffee",
									".config",
									".containerfile",
									".cpp",
									".cS",
									".cshtml",
									".csproj",
									".css",
									".csV",
									".csx",
									".ctp",
									".cxx",
									".dart",
									".diff",
									".dockerfile",
									".dot",
									".dtd",
									".editorconfig",
									".edn",
									".erb",
									".eyaml",
									".eyml",
									".fs",
									".fsi",
									".fsscript",
									".fsx",
									".gemspec",
									".gitattributes",
									".gitconfig",
									".gitignore",
									".go",
									".gradle",
									".groovy",
									".h",
									".h",
									".handlebars",
									".hbs",
									".hh",
									".hpp",
									".htm",
									".html",
									".hxx",
									".ini",
									".ipynb",
									".jade",
									".jav",
									".java",
									".js",
									".jscsrc",
									".jshintrc",
									".jshtm",
									".json",
									".jsp",
									".jsX",
									".less",
									".log",
									".lua",
									".m",
									".makefile",
									".markdown",
									".md",
									".mdoc",
									".mdown",
									".mdtext",
									".mdtxt",
									".mdwn",
									".mjs",
									".mk",
									".mkd",
									".mkdn",
									".ml",
									".mli",
									".npmignore",
									".php",
									".phtml",
									".pl",
									".pl6",
									".plist",
									".pm",
									".pm6",
									".pod",
									".pp",
									".profile",
									".properties",
									".ps1",
									".psd1",
									".psgi",
									".psm1",
									".py",
									".pyi",
									".r",
									".rb",
									".rhistory",
									".rprofile",
									".rs",
									".rst",
									".rt",
									".sass",
									".scss",
									".sh",
									".shtml",
									".sql",
									".svg",
									".svgz",
									".t",
									".tex",
									".toml",
									".ts",
									".tsx",
									".txt",
									".vb",
									".vue",
									".wxi",
									".wxl",
									".wxs",
									".xaml",
									".xhtml",
									".xml",
									".yaml",
									".yml",
									".zsh"},
							}

							fmt.Printf("🔍 Starting Zed context menu and file association cleanup...")

							// Quick check (optional)
							hasAnyEntryBeenFound := false
							quickCheckPath := filepath.Join("Software", "Classes", "*", "shell", config.AppName+"ContextMenuHandler")
							if key, err := registry.OpenKey(registry.CURRENT_USER, quickCheckPath, registry.QUERY_VALUE); err == nil {
								hasAnyEntryBeenFound = true
								key.Close()
							}
							if !hasAnyEntryBeenFound && len(config.FileExtensions) > 0 {
								progIDCheck := fmt.Sprintf("%s%s", config.AppName, config.FileExtensions[0])
								progIDPathCheck := filepath.Join("Software", "Classes", progIDCheck)
								if key, err := registry.OpenKey(registry.CURRENT_USER, progIDPathCheck, registry.QUERY_VALUE); err == nil {
									hasAnyEntryBeenFound = true
									key.Close()
								}
							}
							if !hasAnyEntryBeenFound {
								fmt.Printf("🤔 No obvious Zed registry entries found by quick check. Proceeding with cleanup attempt anyway.")
							} else {
								fmt.Printf("📝 Found indications of existing Zed registry entries. Starting thorough cleanup...")
							}

							// 1. Remove generic context menu entries
							fmt.Printf("--- Removing Generic Context Menu Entries ---")
							genericContextMenuKeyName := config.AppName + "ContextMenuHandler"
							genericBasePaths := map[string]string{
								"all_files":    filepath.Join("Software", "Classes", "*", "shell", genericContextMenuKeyName),
								"directory":    filepath.Join("Software", "Classes", "Directory", "shell", genericContextMenuKeyName),
								"directory_bg": filepath.Join("Software", "Classes", "Directory", "Background", "shell", genericContextMenuKeyName),
							}

							for contextType, menuPath := range genericBasePaths {
								log.Printf("Attempting to remove generic context menu for: %s (Path: %s)", contextType, menuPath)
								// The menuPath itself is the key to delete. It might have a 'command' subkey.
								// deleteRegistryKeyRecursively will handle deleting 'command' first, then 'menuPath'.
								deleteRegistryKeyRecursively(registry.CURRENT_USER, menuPath)
							}

							// 2. Remove file extension associations and ProgIDs
							fmt.Printf("--- Removing File Extension Associations and ProgIDs ---")
							for _, ext := range config.FileExtensions {
								if !strings.HasPrefix(ext, ".") {
									fmt.Printf("⚠️ Skipping invalid extension format: %s (must start with '.')", ext)
									continue
								}

								progID := fmt.Sprintf("%s%s", config.AppName, ext)
								progIDPath := filepath.Join("Software", "Classes", progID)
								extensionKeyPath := filepath.Join("Software", "Classes", ext)
								openWithProgidsPath := filepath.Join(extensionKeyPath, "OpenWithProgids")

								fmt.Printf("Cleaning up for extension: %s (ProgID: %s)", ext, progID)

								// A. Remove ProgID from the extension's OpenWithProgids
								fmt.Printf("Attempting to remove '%s' from OpenWithProgids for %s", progID, ext)
								deleteRegistryValueQuietly(registry.CURRENT_USER, openWithProgidsPath, progID)

								// B. Remove the ProgID key itself (and all its subkeys like shell, DefaultIcon, etc.)
								fmt.Printf("Attempting to remove ProgID key recursively: %s", progIDPath)
								deleteRegistryKeyRecursively(registry.CURRENT_USER, progIDPath)
							}

							// 3. Remove Application Registration (if created)
							appExeName := config.AppName + ".exe"
							appRegPath := filepath.Join("Software", "Classes", "Applications", appExeName)
							fmt.Printf("--- Removing Application Registration (if it exists for %s) ---", appExeName)
							deleteRegistryKeyRecursively(registry.CURRENT_USER, appRegPath)

							fmt.Println("--------------------------------------------------------------------")
							fmt.Println("✨ Zed context menu and file association cleanup attempt complete!")
							fmt.Println("ℹ️ Review the log above for any warnings. You might need to restart Windows Explorer or log out/in for all changes to take full effect.")
							fmt.Println("--------------------------------------------------------------------")

							// for _, ext := range registryConfig.FileExtensions {
							// 	if !strings.HasPrefix(ext, ".") {
							// 		fmt.Printf("⚠️  Skipping non-extension file type: %s (must start with '.')", ext)
							// 		continue
							// 	}

							// 	progId := fmt.Sprintf("%s%s", registryConfig.AppName, ext)
							// 	rootPath := filepath.Join("Software", "Classes", progId)
							// 	defaultIconPath := filepath.Join(rootPath, "DefaultIcon")
							// 	shellPath := filepath.Join(rootPath, "shell")
							// 	openPath := filepath.Join(shellPath, "open")
							// 	commandPath := filepath.Join(openPath, "command")

							// 	if err := registry.DeleteKey(registry.CURRENT_USER, commandPath); err != nil {
							// 		fmt.Printf("⚠️  Error deleting command key for %s: %v", progId, err)
							// 		continue
							// 	}

							// 	if err := registry.DeleteKey(registry.CURRENT_USER, openPath); err != nil {
							// 		fmt.Printf("⚠️  Error deleting open key for %s: %v", progId, err)
							// 		continue
							// 	}

							// 	if err := registry.DeleteKey(registry.CURRENT_USER, shellPath); err != nil {
							// 		fmt.Printf("⚠️  Error deleting shell key for %s: %v", progId, err)
							// 		continue
							// 	}

							// 	if err := registry.DeleteKey(registry.CURRENT_USER, defaultIconPath); err != nil {
							// 		fmt.Printf("⚠️  Error deleting default icon key for %s: %v", progId, err)
							// 		continue
							// 	}

							// 	if err := registry.DeleteKey(registry.CURRENT_USER, rootPath); err == nil {
							// 		fmt.Println("✅ Successfully deleted the key: ", rootPath)
							// 		return nil
							// 	}
							// }

							// fmt.Println("--------------------------------------------------------------------")
							// fmt.Println("✨ Zed context menu and file association cleanup complete!")
							// fmt.Println("ℹ️  You might need to restart Windows Explorer or log out/in for all changes to take full effect.")
							// fmt.Println("--------------------------------------------------------------------")

							// path := filepath.Join("Software", "Classes", "Zed.json")

							// First try to delete the key directly

							// err :=

							// fmt.Printf("🔍 Starting Zed context menu and file association cleanup...\n")

							// // Check if any registry entries exist
							// hasEntries := false
							// basePaths := map[string]string{
							// 	"all_files":            filepath.Join("Software", "Classes", "*", "shell"),
							// 	"directory":            filepath.Join("Software", "Classes", "Directory", "shell"),
							// 	"directory_background": filepath.Join("Software", "Classes", "Directory", "Background", "shell"),
							// }

							// contextMenuKeyName := registryConfig.AppName + "ContextMenuHandler"

							// // First check if any entries exist
							// for _, basePath := range basePaths {
							// 	menuPath := filepath.Join(basePath, contextMenuKeyName)
							// 	key, err := registry.OpenKey(registry.CURRENT_USER, menuPath, registry.QUERY_VALUE)
							// 	if err == nil {
							// 		hasEntries = true
							// 		key.Close()
							// 		break
							// 	}
							// }

							// // Check file extensions
							// if !hasEntries {
							// 	for _, ext := range registryConfig.FileExtensions {
							// 		if !strings.HasPrefix(ext, ".") {
							// 			continue
							// 		}
							// 		progId := fmt.Sprintf("%s%s", registryConfig.AppName, ext)
							// 		progIdPath := filepath.Join("Software", "Classes", progId)
							// 		key, err := registry.OpenKey(registry.CURRENT_USER, progIdPath, registry.QUERY_VALUE)
							// 		if err == nil {
							// 			hasEntries = true
							// 			key.Close()
							// 			break
							// 		}
							// 	}
							// }

							// if !hasEntries {
							// 	fmt.Println("ℹ️  No Zed registry entries found. Nothing to uninstall.")
							// 	return nil
							// }

							// fmt.Println("📝 Found existing Zed registry entries. Starting cleanup...")

							// // Remove generic context menu entries
							// for contextType, basePath := range basePaths {
							// 	menuPath := filepath.Join(basePath, contextMenuKeyName)
							// 	err := registry.DeleteKey(registry.CURRENT_USER, menuPath)
							// 	if err != nil {
							// 		fmt.Printf("⚠️  Failed to remove context menu for %s: %v\n", contextType, err)
							// 	} else {
							// 		fmt.Printf("✅ Successfully removed context menu for: %s\n", contextType)
							// 	}
							// }

							// // Remove file extension associations
							// for _, ext := range registryConfig.FileExtensions {
							// 	if !strings.HasPrefix(ext, ".") {
							// 		fmt.Printf("⚠️  Skipping non-extension file type: %s (must start with '.')\n", ext)
							// 		continue
							// 	}

							// 	// Remove ProgID association
							// 	progId := fmt.Sprintf("%s%s", registryConfig.AppName, ext)
							// 	progIdPath := filepath.Join("Software", "Classes", progId)

							// 	// Remove the OpenWithProgids entry
							// 	extPath := filepath.Join("Software", "Classes", ext, "OpenWithProgids")
							// 	key, err := registry.OpenKey(registry.CURRENT_USER, extPath, registry.WRITE)
							// 	if err == nil {
							// 		err = key.DeleteValue(progId)
							// 		key.Close()
							// 		if err != nil {
							// 			fmt.Printf("⚠️  Failed to remove OpenWithProgids entry for %s: %v\n", ext, err)
							// 		} else {
							// 			fmt.Printf("✅ Successfully removed OpenWithProgids entry for: %s\n", ext)
							// 		}
							// 	}

							// 	// Remove the ProgID key
							// 	err = registry.DeleteKey(registry.CURRENT_USER, progIdPath)
							// 	if err != nil {
							// 		fmt.Printf("⚠️  Failed to remove ProgID for %s: %v\n", ext, err)
							// 	} else {
							// 		fmt.Printf("✅ Successfully removed ProgID for: %s\n", ext)
							// 	}
							// }

							// fmt.Println("--------------------------------------------------------------------")
							// fmt.Println("✨ Zed context menu and file association cleanup complete!")
							// fmt.Println("ℹ️  You might need to restart Windows Explorer or log out/in for all changes to take full effect.")
							// fmt.Println("--------------------------------------------------------------------")
							return nil
						},
					},
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("version") {
				return nil // Early Exit, if version flag is being used
			}

			isRunning, isRunningErr := isZedRunning()

			if isRunning {
				fmt.Println("⚠️  Zed is Already running in another instance!!")
				fmt.Println("This CLI cannot launch a second instance due to Zed limitations.")
				return nil
			}

			if isRunningErr != nil {
				fmt.Println("⚠️  Failed to check running processes: ", isRunningErr)
				fmt.Println("Not you're mistake!")
				return nil
			}

			loadedPath, loadedPathErr := loadConfig()
			if loadedPathErr != nil {
				fmt.Println("Something went wrong during loading Config Path: ", loadedPathErr)
				return nil
			}

			projectPath := cmd.Args().First()
			fmt.Println("Argument passed to root Command: ", projectPath)

			var command *exec.Cmd

			if projectPath != "" {
				if _, err := os.Stat(projectPath); os.IsNotExist(err) {
					err := os.MkdirAll(projectPath, 0755)

					if err != nil {
						fmt.Println("❌ Failed to create folder: ", err)
						os.Exit(1)
					}
					fmt.Println("⚠️  Path doesn't exist")
					fmt.Println("📝 Creating a new folder on: ", filepath.Clean(projectPath))
					fmt.Println("📁 Created new folder: ", filepath.Clean(projectPath))
				}

				command = exec.Command(loadedPath.ZedPath, projectPath)
			} else {
				command = exec.Command(loadedPath.ZedPath)
			}

			command.Stdin = os.Stdin
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr

			commandErr := command.Start()

			if commandErr != nil {
				fmt.Println("❌ Error opening Zed: ", commandErr)
				return nil
			}

			fmt.Println("✅ Zed opened successfully.")
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

	// configPathFlag := flag.String("config-path", "", "The path to the Zed executable")
	// getConfigPathFlag := flag.String("current-path", "", "The current path to the Zed executable")

	// flag.Parse()
	// var actualConfigPath string = ""

	// if *configPathFlag != "" {
	// 	fmt.Println("Received config path: ", *configPathFlag)

	// 	resolvedPath, resolvedPathErr := resolvePath(*configPathFlag)
	// 	fmt.Println("Resolved Path: ", resolvedPath)
	// 	if !fileExists(resolvedPath) || resolvedPathErr != nil {
	// 		fmt.Println("❌ Provided path does not exist.")
	// 		os.Exit(1)
	// 	}

	// test := resolvePath(*configPath)

	// if resolvedPathErr != nil {
	// 	fmt.Println("Something went wrong during resolving the path: ", resolvedPathErr)
	// 	os.Exit(1)
	// }

	// saveConfig(resolvedPath)
	// os.Exit(1)

	// if _, err := os.Stat(*configPath); os.IsNotExist(err) {
	// 	fmt.Println("❌ Provided path does not exist.")
	// 	os.Exit(1)
	// }
	// }

	// saveConfig(*configPath)
	// path, loadConfigErr := loadConfig()

	// if loadConfigErr != nil {
	// 	fmt.Println("❌ Failed to load config.")
	// 	os.Exit(1)
	// }

	// var zedPath string = path.ZedPath
	// fmt.Println(zedPath)

	// const actualConfigPath string = `%LOCALAPPDATA%/Programs/Zed/zed.exe`
	// const examplePath2 string = `C:\Users\Samee\AppData\Local\Programs\Zed\zed.exe`
	// var zedPath string = ""
	// if strings.HasPrefix(actualConfigPath, "%") && strings.Contains(actualConfigPath, "%") {
	// 	var end int = strings.Index(actualConfigPath[1:], "%") + 1
	// 	envName := actualConfigPath[1:end]
	// 	envValue := os.Getenv(envName)

	// 	if envValue == "" {
	// 		fmt.Println("❌ Environment variable is not set or in correct.", envName)
	// 		os.Exit(1)
	// 	}

	// 	restOfPath := actualConfigPath[end+1:]
	// 	zedPath = filepath.Join(envValue, restOfPath)
	// 	fmt.Println("✅ Using resolved path: ", zedPath)
	// } else {
	// 	zedPath = actualConfigPath
	// 	fmt.Println("✅ No Environment variable is used, using the path directly: ", zedPath)
	// }

	// isValidEnv := strings.HasPrefix(examplePath2, "%")
	// fmt.Println("Is Valid Env: ", isValidEnv)

	// if !isValidEnv {
	// 	fmt.Println("❌ Invalid environment variable: ", examplePath2)
	// 	os.Exit(1)
	// }

	// var extractedEnv string = examplePath2[1:13]
	// fmt.Println("Extracted env: ", extractedEnv)

	// extractedEnvPath := os.Getenv(extractedEnv)
	// zedPath := filepath.Join(extractedEnvPath, `Programs\Zed\zed.exe`)
	// var projectPath string = os.Args[1]

	// isRunning, isRunningErr := isZedRunning()
	// fmt.Println("Is Running Zed Returns: ", isRunning, isRunningErr)

	// if isRunningErr != nil {
	// 	fmt.Println("⚠️ Failed to check running processes: ", isRunningErr)
	// 	fmt.Println("Not you're mistake!")
	// } else if isRunning {
	// 	fmt.Println("⚠️  Zed is Already running in another instance!!")
	// 	fmt.Println("This CLI cannot launch a second instance due to Zed limitations.")
	// 	os.Exit(0)
	// }

	// if _, err := os.Stat(zedPath); os.IsNotExist(err) {
	// 	fmt.Printf("❌ Could not find Zed at %s\n", zedPath)
	// 	fmt.Println()
	// 	fmt.Println("👉 Tip: If Zed is installed somewhere else, you can set the path manually: ")
	// 	fmt.Println()
	// 	fmt.Println(`zed --config-path "C:\Users\YourUser\AppData\Local\Programs\Zed\zed.exe"`)
	// 	fmt.Println()
	// 	fmt.Println("ℹ️  This only needs to be done once. The path will be saved for future runs.")
	// 	os.Exit(1)
	// }

	// if _, err := os.Stat(projectPath); os.IsNotExist(err) {
	// 	err := os.MkdirAll(projectPath, 0755)

	// 	if err != nil {
	// 		fmt.Println("❌ Failed to create folder: ", err)
	// 		os.Exit(1)
	// 	}
	// 	fmt.Println("Path doesn't exist")
	// 	fmt.Println("Creating a new folder on: ", filepath.Clean(projectPath))
	// 	fmt.Println("📁 Created new folder: ", filepath.Clean(projectPath))
	// }

	// cmd := exec.Command(zedPath, projectPath)
	// cmd.Stdout = os.Stdout
	// cmd.Stdin = os.Stdin
	// cmd.Stderr = os.Stderr

	// var err error = cmd.Start()

	// if err != nil {
	// 	fmt.Println("❌ Error opening project in Zed: ", err)
	// 	os.Exit(1)
	// }

	// fmt.Println("✅ Zed launched successfully.")
	// fmt.Println(projectPath)

}

// func isZedRunningV2() (bool, error) {
// 	var lockFilePath string = filepath.Join(os.TempDir(), "zed-test.lock")

// 	file, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
// 	if err != nil {
// 		return true, nil
// 	}

// 	fmt.Println(file, "%d", os.Getpid())

// 	go func() {
// 		c := make(chan os.Signal, 1)
// 		<-c
// 		os.Remove(lockFilePath)
// 	}()

// 	return false, nil
// }
