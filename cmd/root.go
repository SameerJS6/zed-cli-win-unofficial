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
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Println("v1.0.0")
	}

	cli.RootCommandHelpTemplate = fmt.Sprintf(`%s
	WEBSITE: https://zedcli.sameerjs.com,
	SUPPORT: https://github.com/SameerJS6/zed-cli-win-unofficial/issues`, cli.RootCommandHelpTemplate)

	app := &cli.Command{
		Name:        "zed",
		Usage:       "Zed's Unofficial Windows CLI",
		Description: "An unofficial Windows command-line interface for the Zed editor. Launch Zed projects, manage configuration, and install context menu integration.",
		Version:     "1.0.0",
		Authors: []any{
			"SameerJS6 <contact@sameerjs.com>",
		},
		Copyright:   "Copyright (c) 2025 SameerJS6",
		ArgsUsage:   "[project-path]",
		UsageText:   "zed [global options] [project-path]\n   zed [global options] command [command options] [arguments...]",
		Category:    "Development Tools",
		Suggest:     true,
		HideHelp:    false,
		HideVersion: false,
		Commands: []*cli.Command{
			configCommand(),
			contextCommand(),
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				utils.PrintZedNotFoundBanner("")
				utils.Error(fmt.Sprintf("Error loading config: %v", err))
				utils.Infoln("👉 Tip: Run `zed config set <path>` to configure the Zed executable path.")
				return nil
			}

			if !config.FileExists(cfg.ZedPath) {
				utils.PrintZedNotFoundBanner("")
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
