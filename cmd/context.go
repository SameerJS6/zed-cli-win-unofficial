package cmd

import (
	"context"
	"fmt"
	"strings"
	"zed-cli-win-unofficial/internal/config"
	"zed-cli-win-unofficial/internal/fileext"
	"zed-cli-win-unofficial/internal/registry"

	"github.com/urfave/cli/v3"
)

func contextCommand() *cli.Command {
	return &cli.Command{
		Name:  "context",
		Usage: "Configure the `Open with Zed` in context menu",
		Commands: []*cli.Command{
			{
				Name:  "install",
				Usage: "To install Open with Zed in context menu feature",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := config.LoadConfig()
					if err != nil {
						fmt.Printf("❌ Error loading config: %v\n", err)
						fmt.Println("👉 Tip: Run `zed config set <path>` to configure the Zed executable path.")
						return nil
					}

					if !config.FileExists(cfg.ZedPath) {
						fmt.Printf("❌ Configured Zed path does not exist: %s\n", cfg.ZedPath)
						fmt.Println("👉 Tip: Run `zed config set <path>` to update the path.")
						return nil
					}

					registryCfg := registry.NewConfig(cfg.ZedPath, fileext.SupportedExtensions())

					fmt.Println("🚀 Setting up Zed context menu and file associations...")

					if err := registry.InstallGenericContextMenu(registryCfg); err != nil {
						fmt.Printf("❌ Failed to install context menu: %v\n", err)
						return nil
					}

					for _, ext := range registryCfg.FileExtensions {
						if !strings.HasPrefix(ext, ".") && !strings.Contains(ext, ".") {
							fmt.Printf("⚠️ Skipping invalid file type: %s\n", ext)
							continue
						}

						if err := registry.CreateProgID(registryCfg, ext); err != nil {
							fmt.Printf("⚠️ Failed to register %s files with Zed, skipping\n", ext)
							continue
						}

						progID := fmt.Sprintf("%s%s", registryCfg.AppName, ext)
						if err := registry.AssociateExtensionWithProgID(ext, progID); err != nil {
							fmt.Printf("❌ Failed to associate %s files with Zed: %v\n", ext, err)
							return nil
						}
					}

					cfg.ContextMenuEnabled = true
					if err := config.SaveConfig(cfg); err != nil {
						fmt.Printf("❌ Error saving config: %v\n", err)
						return nil
					}

					fmt.Println("--------------------------------------------------------------------")
					fmt.Println("✅ Zed context menu and file associations setup complete!")
					fmt.Println("💡 You may need to restart Windows Explorer for changes to take effect.")
					fmt.Println("🔧 To remove these entries, run: zed context uninstall")
					fmt.Println("--------------------------------------------------------------------")

					return nil
				},
			},
			{
				Name:  "uninstall",
				Usage: "Uninstall the 'Open with Zed' context menu option",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := config.LoadConfig()

					if err != nil {
						fmt.Printf("❌ Error loading config: %v\n", err)
						return nil
					}

					if !cfg.ContextMenuEnabled {
						fmt.Println("ℹ️  Zed context menu is not installed. Nothing to remove.")
						return nil
					}

					registryConfig := registry.NewConfig("", fileext.SupportedExtensions())

					fmt.Println("🧹 Removing Zed context menu and file associations...")

					if err := registry.UninstallAllContextMenus(registryConfig); err != nil {
						fmt.Printf("❌ Failed to remove context menu: %v\n", err)
						return nil
					}

					cfg.ContextMenuEnabled = false
					if err := config.SaveConfig(cfg); err != nil {
						fmt.Printf("❌ Error saving config: %v\n", err)
						return nil
					}

					fmt.Println("--------------------------------------------------------------------")
					fmt.Println("✅ Zed context menu and file associations removed successfully.")
					fmt.Println("💡 You may need to restart Windows Explorer for changes to take effect.")
					fmt.Println("--------------------------------------------------------------------")

					return nil
				},
			},
		},
	}
}
