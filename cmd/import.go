package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/importer"

	"github.com/mholt/archiver/v3"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// RunImport cmd import torrents
func RunImport() *cobra.Command {
	var (
		source     string
		sourceDir  string
		qbitDir    string
		dryRun     bool
		skipBackup bool
	)

	var command = &cobra.Command{
		Use:   "import",
		Short: "import torrents",
		Long:  `Import torrents with state from other clients [rtorrent, deluge]`,
	}
	command.Flags().StringVar(&source, "source", "", "source client [deluge, rtorrent]")
	command.Flags().StringVar(&sourceDir, "source-dir", "", "source client state dir")
	command.Flags().StringVar(&qbitDir, "qbit-dir", "", "qbit dir")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Run without doing anything")
	command.Flags().BoolVar(&skipBackup, "skip-backup", false, "Skip backup before import. Not advised")

	command.MarkFlagRequired("source")
	command.MarkFlagRequired("source-dir")
	command.MarkFlagRequired("qbit-dir")

	command.Run = func(cmd *cobra.Command, args []string) {

		// TODO check if program is running, if true exit

		// Backup data before running
		if skipBackup != true {
			fmt.Print("Prepare to backup torrent data before import..\n")
			t := time.Now().Format("20060102150405")

			homeDir, err := homedir.Dir()
			if err != nil {
				fmt.Printf("could not find home directory: %v", err)
			}

			sourceBackupArchive := fmt.Sprintf(homeDir + "/qbt_backup/" + source + "_backup_" + t + ".tar.gz")
			qbitBackupArchive := fmt.Sprintf(homeDir + "/qbt_backup/qBittorrent_backup_" + t + ".tar.gz")

			err = archiver.Archive([]string{sourceDir}, sourceBackupArchive)
			if err != nil {
				log.Fatalf("could not backup directory: %v", err)
			}
			fmt.Printf("Backup %v directory: %v to %v\n", source, sourceDir, sourceBackupArchive)

			err = archiver.Archive([]string{qbitDir}, qbitBackupArchive)
			if err != nil {
				log.Fatalf("could not backup directory: %v", err)
			}
			fmt.Printf("Backup qBittorrent directory: %v to %v\n", qbitDir, qbitBackupArchive)

			fmt.Print("Backup completed!\n")
		}

		opts := importer.Options{
			SourceDir: sourceDir,
			QbitDir:   qbitDir,
			DryRun:    dryRun,
		}

		start := time.Now()

		switch source {
		case "deluge":
			d := importer.NewDelugeImporter()

			err := d.Import(opts)
			if err != nil {
				fmt.Printf("import error: %v", err)
			}

		case "rtorrent":
			r := importer.NewRTorrentImporter()

			err := r.Import(opts)
			if err != nil {
				fmt.Printf("import error: %v", err)
			}

		default:
			fmt.Println("WARNING: Unsupported client!")
			break
		}

		elapsed := time.Since(start)
		fmt.Printf("\nImport finished in: %s\n", elapsed)

	}

	return command
}
