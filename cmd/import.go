package cmd

import (
	"fmt"

	"github.com/ludviglundgren/qbittorrent-cli/internal/importer"

	"github.com/spf13/cobra"
)

// RunImport cmd import torrents
func RunImport() *cobra.Command {
	var (
		from      string
		delugeDir string
		qbitDir   string
	)

	var command = &cobra.Command{
		Use:   "import",
		Short: "import torrents",
		Long:  `Import torrents from other client`,
	}
	command.Flags().StringVar(&from, "from", "", "from client")
	command.Flags().StringVar(&delugeDir, "deluge-dir", "", "deluge dir")
	command.Flags().StringVar(&qbitDir, "qbit-dir", "", "qbit dir")

	command.Run = func(cmd *cobra.Command, args []string) {

		// TODO check if program is running, if true exit

		switch from {
		case "deluge":
			d := importer.NewDelugeImporter()
			opts := importer.Options{
				DelugeDir: delugeDir,
				QbitDir:   qbitDir,
			}

			d.Import(opts)
		default:
			fmt.Println("WARNING: Unsupported client!")
			break
		}
	}

	return command
}
