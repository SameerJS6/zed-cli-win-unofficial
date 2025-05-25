package cmd

import (
	"context"
	"fmt"
	"zed-cli-win-unofficial/internal/config"

	"github.com/urfave/cli/v3"
)

func configCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Configure the CLI's path for everything",
		Commands: []*cli.Command{
			{
				Name:  "set",
				Usage: "Set the executable's path",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					path := cmd.Args().First()

					if path == "" {
						fmt.Println("❌ No path provided.")
						fmt.Println("Usage: zed config set <path-to-zed-executable>")
						return nil
					}

					validatedPath, err := config.ValidatePath(path)
					if err != nil {
						fmt.Printf("❌ %v\n", err)
						return nil
					}

					configStruct := &config.Config{
						ZedPath: validatedPath,
					}

					if err := config.SaveConfig(configStruct); err != nil {
						fmt.Printf("%v\n", err)
						return nil
					}

					fmt.Printf("Getting the path to the Zed executable: %s\n", path)
					fmt.Println("You might want to update the context menu and file association for the current user.")
					fmt.Println("You can do that by running: zed context install")
					return nil
				},
			},
			{
				Name:  "get",
				Usage: "Get the currently set executable's path",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := config.LoadConfig()

					if err != nil {
						fmt.Printf("❌ Error loading config: %v\n", err)
						fmt.Println("👉 Tip: Run `zed config set <path>` to create a config file.")
						return nil
					}

					if !config.FileExists(cfg.ZedPath) {
						fmt.Printf("❌ Configured Zed path does not exist: %s\n", cfg.ZedPath)
						fmt.Println("👉 Tip: Run `zed config set <path>` to update the path.")
						return nil
					}

					fmt.Printf("✅ Zed is configured to run from: %s\n", cfg.ZedPath)
					return nil
				},
			},
		},
	}
}
