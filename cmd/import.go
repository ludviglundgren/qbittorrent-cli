package cmd

import (
	"log"

	"github.com/ludviglundgren/qbittorrent-cli/internal/importer"

	"github.com/spf13/cobra"
)

// RunImport cmd import torrents
func RunImport() *cobra.Command {
	var (
		from      string
		stateDir  string
		delugeDir string
		qbitDir   string
	)

	var command = &cobra.Command{
		Use:   "import",
		Short: "import torrents",
		Long:  `Import torrents from other client`,
	}
	command.Flags().StringVar(&from, "from", "", "from client")
	command.Flags().StringVar(&stateDir, "state-dir", "", "deluge dir")
	command.Flags().StringVar(&delugeDir, "deluge-dir", "", "deluge dir")
	command.Flags().StringVar(&qbitDir, "qbit-dir", "", "qbit dir")

	command.Run = func(cmd *cobra.Command, args []string) {
		log.Println("Importing torrents..")

		// TODO check if program is running, if true exit

		d := importer.NewDelugeImporter()
		opts := importer.Options{
			StateDir:  stateDir,
			DelugeDir: delugeDir,
			QbitDir:   qbitDir,
		}

		d.Import(opts)
	}

	return command
}
