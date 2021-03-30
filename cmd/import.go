package cmd

import (
	"fmt"

	"github.com/ludviglundgren/qbittorrent-cli/internal/importer"

	"github.com/spf13/cobra"
)

// RunImport cmd import torrents
func RunImport() *cobra.Command {
	var (
		source    string
		sourceDir string
		qbitDir   string
		dryRun    bool
	)

	var command = &cobra.Command{
		Use:   "import",
		Short: "import torrents",
		Long:  `Import torrents which state from other clients [rtorrent, deluge]`,
	}
	command.Flags().StringVar(&source, "source", "", "source client [deluge, rtorrent]")
	command.Flags().StringVar(&sourceDir, "source-dir", "", "source state dir")
	command.Flags().StringVar(&qbitDir, "qbit-dir", "", "qbit dir")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Run without doing anything")

	command.Run = func(cmd *cobra.Command, args []string) {

		// TODO check if program is running, if true exit

		opts := importer.Options{
			SourceDir: sourceDir,
			QbitDir:   qbitDir,
			DryRun:    dryRun,
		}

		switch source {
		case "deluge":
			d := importer.NewDelugeImporter()

			d.Import(opts)

		case "rtorrent":
			r := importer.NewRTorrentImporter()

			r.Import(opts)

		default:
			fmt.Println("WARNING: Unsupported client!")
			break
		}
	}

	return command
}
