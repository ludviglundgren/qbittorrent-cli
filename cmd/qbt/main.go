package main

import (
	"io"
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

	var silentOutput bool

	log.SetFlags(0)

	var rootCmd = &cobra.Command{
		Use:   "qbt",
		Short: "Manage qBittorrent with cli",
		Long: `Manage qBittorrent from command line.

Documentation is available at https://github.com/ludviglundgren/qbittorrent-cli`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if silentOutput {
				log.SetOutput(io.Discard)
				// Uncomment to also suppress standard output
				// os.Stdout = io.Discard
			}
		},
	}

	// override config
	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is $HOME/.config/qbt/.qbt.toml)")
	rootCmd.PersistentFlags().BoolVarP(&silentOutput, "quiet", "q", false, "suppress output")

	rootCmd.AddCommand(cmd.RunApp())
	rootCmd.AddCommand(cmd.RunBencode())
	rootCmd.AddCommand(cmd.RunTorrent())
	rootCmd.AddCommand(cmd.RunCategory())
	rootCmd.AddCommand(cmd.RunTag())
	rootCmd.AddCommand(cmd.RunVersion(version, commit, date))
	rootCmd.AddCommand(cmd.RunUpdate(version))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
