package cmd

import (
	"context"
	"fmt"
	"strings"
	"zed-cli-win-unofficial/internal/config"
	"zed-cli-win-unofficial/internal/fileext"
	"zed-cli-win-unofficial/internal/registry"
	"zed-cli-win-unofficial/internal/utils"

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
					utils.Debugln("Starting context menu installation...")

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

					utils.Debugln("🚀 Setting up Zed context menu and file associations...")

					if err := registry.InstallGenericContextMenu(registryCfg); err != nil {
						utils.Error(fmt.Sprintf("Failed to install context menu: %v", err))
						return nil
					}

					for _, ext := range registryCfg.FileExtensions {
						if !strings.HasPrefix(ext, ".") && !strings.Contains(ext, ".") {
							utils.Debug("Skipping invalid file type: %s\n", ext)
							continue
						}

						if err := registry.CreateProgID(registryCfg, ext); err != nil {
							utils.Debug("Failed to register %s files with Zed, skipping\n", ext)
							continue
						}

						progID := fmt.Sprintf("%s%s", registryCfg.AppName, ext)
						if err := registry.AssociateExtensionWithProgID(ext, progID); err != nil {
							utils.Error(fmt.Sprintf("Failed to associate %s files with Zed: %v", ext, err))
							return nil
						}
					}

					cfg.ContextMenuEnabled = true
					if err := config.SaveConfig(cfg); err != nil {
						utils.Error(fmt.Sprintf("Error saving config: %v", err))
						return nil
					}

					utils.Success("Zed context menu and file associations setup complete!")
					utils.Infoln("💡 Optional: Restart Explorer—rarely necessary for current user changes.")
					utils.Infoln("🔧 To remove these entries, run: zed context uninstall")

					return nil
				},
			},
			{
				Name:  "uninstall",
				Usage: "Uninstall the 'Open with Zed' context menu option",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					utils.Debugln("Starting context menu uninstallation...")

					cfg, err := config.LoadConfig()

					if err != nil {
						fmt.Printf("❌ Error loading config: %v\n", err)
						return nil
					}

					if !cfg.ContextMenuEnabled {
						utils.Infoln("ℹ️  Zed context menu is not installed. Nothing to remove.")
						return nil
					}

					registryConfig := registry.NewConfig("", fileext.SupportedExtensions())

					utils.Debugln("🧹 Removing Zed context menu and file associations...")

					if err := registry.UninstallAllContextMenus(registryConfig); err != nil {
						utils.Error(fmt.Sprintf("Failed to remove context menu: %v", err))
						return nil
					}

					cfg.ContextMenuEnabled = false
					if err := config.SaveConfig(cfg); err != nil {
						utils.Error(fmt.Sprintf("Error saving config: %v", err))
						return nil
					}

					utils.Success("Zed context menu and file associations removed successfully.")
					utils.Infoln("💡 Optional: Restart Explorer—rarely necessary for current user changes.")

					return nil
				},
			},
		},
	}
}
