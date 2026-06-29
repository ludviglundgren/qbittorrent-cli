package cmd

import (
	"io"
	"log"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/spf13/cobra"
)

// NewRootCmd builds the root qbt command with all subcommands attached.
//
// version, commit and date are injected at build time and surfaced by
// `qbt version`. It is used both by the CLI entrypoint (cmd/qbt) and by the
// documentation generator (tools/gen-docs), so the generated docs always match
// the real command tree.
func NewRootCmd(version, commit, date string) *cobra.Command {
	var silentOutput bool

	rootCmd := &cobra.Command{
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

	rootCmd.AddCommand(RunApp())
	rootCmd.AddCommand(RunBencode())
	rootCmd.AddCommand(RunTorrent())
	rootCmd.AddCommand(RunTransfer())
	rootCmd.AddCommand(RunCategory())
	rootCmd.AddCommand(RunTag())
	rootCmd.AddCommand(RunVersion(version, commit, date))
	rootCmd.AddCommand(RunUpdate(version))

	return rootCmd
}
