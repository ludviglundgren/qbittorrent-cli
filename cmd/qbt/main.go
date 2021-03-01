package main

import (
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/cmd"
	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/spf13/cobra"
)

func main() {
	cobra.OnInitialize(config.InitConfig)

	var rootCmd = &cobra.Command{
		Use:   "qbt",
		Short: "Manage Qbittorrent with cli",
		Long: `Manage Qbittorrent from command line.

Documentation is available at http://github.com/ludviglundgren/qbittorrent-cli`,
	}

	// override config
	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is $HOME/.config/qbt/.qbt.toml)")

	rootCmd.AddCommand(cmd.RunVersion())
	rootCmd.AddCommand(cmd.RunList())
	rootCmd.AddCommand(cmd.RunAdd())
	rootCmd.AddCommand(cmd.RunImport())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
