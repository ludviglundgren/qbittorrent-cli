package cmd

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/importer"

	"github.com/mholt/archiver/v3"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// RunImport cmd import torrents
func RunImport() *cobra.Command {
	var command = &cobra.Command{
		Use:   "import",
		Short: "import torrents",
		Long:  `Import torrents with state from other clients [rtorrent, deluge]`,
	}

	var (
		source     string
		sourceDir  string
		qbitDir    string
		dryRun     bool
		skipBackup bool
	)

	command.Flags().StringVar(&source, "source", "", "source client [deluge, rtorrent] (required)")
	command.Flags().StringVar(&sourceDir, "source-dir", "", "source client state dir (required)")
	command.Flags().StringVar(&qbitDir, "qbit-dir", "", "qbit dir (required)")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Run without doing anything")
	command.Flags().BoolVar(&skipBackup, "skip-backup", false, "Skip backup before import. Not advised")

	command.MarkFlagRequired("source")
	command.MarkFlagRequired("source-dir")
	command.MarkFlagRequired("qbit-dir")

	command.RunE = func(cmd *cobra.Command, args []string) error {

		// TODO check if program is running, if true exit

		// Backup data before running
		if !skipBackup {
			fmt.Print("Prepare to backup torrent data before import..\n")

			homeDir, err := homedir.Dir()
			if err != nil {
				fmt.Printf("could not find home directory: %q", err)
				return err
			}

			timeStamp := time.Now().Format("20060102150405")
			sourceBackupArchive := path.Join(homeDir, "/qbt_backup/", source+"_backup_"+timeStamp+".tar.gz")
			qbitBackupArchive := path.Join(homeDir, "/qbt_backup/", "qBittorrent_backup_"+timeStamp+".tar.gz")

			fmt.Printf("Creating %s backup of directory: %s to %s ...\n", source, sourceDir, sourceBackupArchive)
			if err = archiver.Archive([]string{sourceDir}, sourceBackupArchive); err != nil {
				log.Printf("could not backup directory: %q", err)
				return err
			}

			fmt.Printf("Creating qBittorrent backup of directory: %s to %s ...\n", qbitDir, qbitBackupArchive)
			if err = archiver.Archive([]string{qbitDir}, qbitBackupArchive); err != nil {
				log.Printf("could not backup directory: %q", err)
				return err
			}

			fmt.Print("Backup completed!\n")
		}

		opts := importer.Options{
			SourceDir: sourceDir,
			QbitDir:   qbitDir,
			DryRun:    dryRun,
		}

		start := time.Now()

		fmt.Printf("Preparing to import torrents from: %s dir: %s\n", source, sourceDir)
		if dryRun {
			fmt.Println("running with --dry-run, no data will be written")
		}

		switch source {
		case "deluge":
			d := importer.NewDelugeImporter()

			if err := d.Import(opts); err != nil {
				fmt.Printf("deluge import error: %q", err)
			}

		case "rtorrent":
			r := importer.NewRTorrentImporter()

			if err := r.Import(opts); err != nil {
				fmt.Printf("rtorrent import error: %q", err)
			}

		default:
			fmt.Println("WARNING: Unsupported client!")
		}

		elapsed := time.Since(start)
		fmt.Printf("\nImport finished in: %s\n", elapsed)

		return nil
	}

	return command
}
