package cmd

import (
	"fmt"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/fs"
	"github.com/ludviglundgren/qbittorrent-cli/internal/importer"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// RunImport cmd import torrents
func RunImport() *cobra.Command {
	var (
		source    string
		sourceDir string
		qbitDir   string
		dryRun    bool
		noBackup  bool
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
	command.Flags().BoolVar(&noBackup, "no-backup", false, "Don't backup before running")

	command.MarkFlagRequired("source")
	command.MarkFlagRequired("source-dir")
	command.MarkFlagRequired("qbit-dir")

	command.Run = func(cmd *cobra.Command, args []string) {

		// TODO check if program is running, if true exit

		// Backup data before running
		if noBackup != true {
			fmt.Print("Prepare to backup data..\n")
			t := time.Now().Format("2006-01-02_15-04-05")

			homeDir, err := homedir.Dir()
			if err != nil {
				fmt.Printf("could not find home directory: %v", err)
			}

			sourceBackupDir := homeDir + "/qbt_backup/" + source + "_backup_" + t
			qbitBackupDir := homeDir + "/qbt_backup/qBittorrent_backup_" + t

			fmt.Printf("Backup %v directory: %v ..\n", source, sourceBackupDir)
			err = fs.CopyDir(sourceDir, sourceBackupDir)
			if err != nil {
				fmt.Printf("could not backup directory: %v", err)
			}
			fmt.Print("Done!\n")

			fmt.Printf("Backup %v directory: %v .. \n", "qBittorrent", qbitBackupDir)
			err = fs.CopyDir(qbitDir, qbitBackupDir)
			if err != nil {
				fmt.Printf("could not backup directory: %v\n", err)
			}
			fmt.Print("Done!\n")

			fmt.Print("Backup done!\n")
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
