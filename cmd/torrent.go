package cmd

import (
	"github.com/spf13/cobra"
)

// RunTorrent cmd for torrent operations
func RunTorrent() *cobra.Command {
	var command = &cobra.Command{
		Use:   "torrent",
		Short: "torrent subcommand",
		Long:  `Do various torrent operations`,
	}

	command.AddCommand(RunTorrentAdd())
	command.AddCommand(RunTorrentCategory())
	command.AddCommand(RunTorrentCompare())
	command.AddCommand(RunTorrentExport())
	command.AddCommand(RunTorrentHash())
	command.AddCommand(RunTorrentImport())
	command.AddCommand(RunTorrentList())
	command.AddCommand(RunTorrentPause())
	command.AddCommand(RunTorrentReannounce())
	command.AddCommand(RunTorrentRemove())
	command.AddCommand(RunTorrentResume())

	return command
}
