package cmd

import (
	"context"
	"fmt"
	"zed-cli-win-unofficial/internal/config"
	"zed-cli-win-unofficial/internal/telementry"
	"zed-cli-win-unofficial/internal/utils"

	"github.com/urfave/cli/v3"
)

func configCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Configure the CLI's Path & Settings",
		Commands: []*cli.Command{
			{
				Name:  "set",
				Usage: "Set the path to the Zed executable",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					path := cmd.Args().First()

					if path == "" {
						telementry.TrackEvent("no_path_provided_error", map[string]any{
							"error_type": "no_path_provided",
						})
						utils.Error("No path provided.")
						return nil
					}

					resolvedPath, err := config.ValidatePath(path)
					if err != nil {
						utils.PrintInvalidPathBanner()
						telementry.TrackEvent("invalid_path_error", map[string]any{
							"error_type": "invalid_path",
						})
						utils.Error(fmt.Sprintf("Invalid path: %v", err))
						return nil
					}

					cfg := &config.Config{
						ZedPath:            resolvedPath,
						ContextMenuEnabled: false,
					}

					if err := config.SaveConfig(cfg); err != nil {
						telementry.TrackEvent("config_save_error", map[string]any{
							"error_type": "config_save_error",
						})
						utils.Error(fmt.Sprintf("Error saving config: %v", err))
						return nil
					}

					utils.Success(fmt.Sprintf("Zed path configured: %s", resolvedPath))
					utils.Infoln("💡 You may want to run `zed context install` to set up or update context menus.")
					telementry.TrackEvent("config_save_success", map[string]any{})
					return nil
				},
			},
			{
				Name:  "get",
				Usage: "Get the current path to the Zed executable",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := config.LoadConfig()
					if err != nil {
						utils.Error(fmt.Sprintf("Error loading config: %v", err))
						utils.Infoln("👉 Tip: Run `zed config set <path>` to configure the Zed executable path.")
						return nil
					}

					if !config.FileExists(cfg.ZedPath) {
						utils.Error("Configured Zed path no longer exists")
						utils.Infoln("👉 Tip: Run `zed config set <path>` to update the path.")
						return nil
					}

					telementry.TrackEvent("config_get_success", map[string]any{})
					utils.Success(fmt.Sprintf("Zed is configured at: %s", cfg.ZedPath))
					return nil
				},
			},
		},
	}
}
