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
		fmt.Println("👉 Tip: Run `zed config set <path>` to create a config file.")
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
