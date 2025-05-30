package cmd

import (
	"context"
	"fmt"
	"os"
	"zed-cli-win-unofficial/internal/config"
	"zed-cli-win-unofficial/internal/process"
	"zed-cli-win-unofficial/internal/utils"

	"github.com/urfave/cli/v3"
)

func Execute(ctx context.Context) error {
	app := &cli.Command{
		Name:  "zed",
		Usage: "Zed's Unofficial Windows CLI",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Print the version of Zed CLI",
				Action: func(ctx context.Context, cmd *cli.Command, value bool) error {
					if value {
						utils.Infoln("v1.0.0")
						return nil
					}
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			configCommand(),
			contextCommand(),
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("version") {
				return nil // Early exit for version flag
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				utils.Error(fmt.Sprintf("Error loading config: %v", err))
				utils.Infoln("👉 Tip: Run `zed config set <path>` to configure the Zed executable path.")
				return nil
			}

			if !config.FileExists(cfg.ZedPath) {
				utils.Error(fmt.Sprintf("Configured Zed path does not exist: %s", cfg.ZedPath))
				utils.Infoln("👉 Tip: Run `zed config set <path>` to update the path.")
				return nil
			}

			projectPath := cmd.Args().First()
			return process.LaunchZed(cfg.ZedPath, projectPath)
		},
	}

	return app.Run(ctx, os.Args)
}
