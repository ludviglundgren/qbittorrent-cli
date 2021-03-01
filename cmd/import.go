package cmd

import (
	"log"

	"github.com/ludviglundgren/qbittorrent-cli/internal/importer"

	"github.com/spf13/cobra"
)

// RunImport cmd import torrents
func RunImport() *cobra.Command {
	var from string

	var command = &cobra.Command{
		Use:   "import",
		Short: "import torrents",
		Long:  `Import torrents from other client`,
	}
	command.Flags().StringVar(&from, "from", "", "from client")

	command.Run = func(cmd *cobra.Command, args []string) {
		log.Println("Importing torrents..")

		// TODO check if program is running, if true exit

		d := importer.NewDelugeImporter()
		d.Import("./test/import/deluge/torrents.fastresume")
	}

	return command
}
