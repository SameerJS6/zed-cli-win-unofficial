package cmd

import (
	"context"
	"fmt"
	"os"
	"zed-cli-win-unofficial/internal/config"
	"zed-cli-win-unofficial/internal/process"

	"github.com/urfave/cli/v3"
)

func Execute(ctx context.Context) error {
	app := &cli.Command{
		Name:  "zed",
		Usage: "Zed's unofficial windows CLI",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Print the current version of cli",
				Action: func(ctx context.Context, cmd *cli.Command, b bool) error {
					if b {
						fmt.Println("v1.0.0")
					}
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			configCommand(),
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("version") {
				return nil
			}

			loadedPath, err := config.LoadConfig()

			if err != nil {
				fmt.Println("Failed to load the executable path from config")
				return nil
			}

			if !config.FileExists(loadedPath.ZedPath) {
				fmt.Println("Provided executable path from config doesn't exist")
				return nil
			}

			pathArgument := cmd.Args().First()

			if err := process.LaunchZed(loadedPath.ZedPath, pathArgument); err != nil {
				fmt.Printf("%v\n", err)
				return nil
			}

			return nil
		},
	}

	return app.Run(ctx, os.Args)
}
