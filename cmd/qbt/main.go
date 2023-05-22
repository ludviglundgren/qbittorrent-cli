package main

import (
	"log"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/cmd"
	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {

	log.SetFlags(0)

	var rootCmd = &cobra.Command{
		Use:   "qbt",
		Short: "Manage qBittorrent with cli",
		Long: `Manage qBittorrent from command line.

Documentation is available at https://github.com/ludviglundgren/qbittorrent-cli`,
	}

	// override config
	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is $HOME/.config/qbt/.qbt.toml)")

	rootCmd.AddCommand(cmd.RunVersion(version, commit, date))
	rootCmd.AddCommand(cmd.RunAdd())
	rootCmd.AddCommand(cmd.RunCompare())
	rootCmd.AddCommand(cmd.RunEdit())
	rootCmd.AddCommand(cmd.RunExport())
	rootCmd.AddCommand(cmd.RunHash())
	rootCmd.AddCommand(cmd.RunImport())
	rootCmd.AddCommand(cmd.RunList())
	rootCmd.AddCommand(cmd.RunMove())
	rootCmd.AddCommand(cmd.RunPause())
	rootCmd.AddCommand(cmd.RunReannounce())
	rootCmd.AddCommand(cmd.RunRemove())
	rootCmd.AddCommand(cmd.RunResume())
	rootCmd.AddCommand(cmd.RunTag())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
