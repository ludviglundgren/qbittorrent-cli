package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/importer"

	"github.com/mholt/archives"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RunTorrentImport cmd import torrents
func RunTorrentImport() *cobra.Command {
	var command = &cobra.Command{
		Use:   "import {rtorrent | deluge} --source-dir dir --qbit-dir dir2 [--skip-backup] [--dry-run]",
		Short: "Import torrents",
		Long:  `Import torrents with state from other clients [rtorrent, deluge]`,
		Example: `  qbt torrent import deluge --source-dir ~/.config/deluge/state/ --qbit-dir ~/.local/share/data/qBittorrent/BT_backup --dry-run
  qbt torrent import rtorrent --source-dir ~/.sessions --qbit-dir ~/.local/share/data/qBittorrent/BT_backup --dry-run`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a source client [rtorrent, deluge] as first argument")
			}

			return cobra.OnlyValidArgs(cmd, args)
		},
		ValidArgs: []string{"rtorrent", "deluge"},
	}

	var (
		sourceDir  string
		qbitDir    string
		dryRun     bool
		skipBackup bool
	)

	command.Flags().BoolVar(&dryRun, "dry-run", false, "Run without importing anything")
	command.Flags().StringVar(&sourceDir, "source-dir", "", "source client state dir (required)")
	command.Flags().StringVar(&qbitDir, "qbit-dir", "", "qBittorrent BT_backup dir. Commonly ~/.local/share/qBittorrent/BT_backup (required)")
	command.Flags().BoolVar(&skipBackup, "skip-backup", false, "Skip backup before import")

	command.MarkFlagRequired("source-dir")
	command.MarkFlagRequired("qbit-dir")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		source := args[0]

		var imp importer.Importer

		switch source {
		case "deluge":
			imp = importer.NewDelugeImporter()

		case "rtorrent":
			imp = importer.NewRTorrentImporter()

		default:
			return fmt.Errorf("error: unsupported client: %s", source)
		}

		// TODO check if program is running, if true exit

		// Backup data before running
		if !skipBackup {
			fmt.Print("prepare to backup torrent data before import..\n")

			homeDir, err := homedir.Dir()
			if err != nil {
				fmt.Printf("could not find home directory: %q", err)
				return err
			}

			timeStamp := time.Now().Format("20060102150405")

			sourceBackupArchive := filepath.Join(homeDir, "qbt_backup", source+"_backup_"+timeStamp+".tar.gz")
			qbitBackupArchive := filepath.Join(homeDir, "qbt_backup", "qBittorrent_backup_"+timeStamp+".tar.gz")

			if dryRun {
				fmt.Printf("dry-run: creating %s backup of directory: %s to %s ...\n", source, sourceDir, sourceBackupArchive)
			} else {
				fmt.Printf("creating %s backup of directory: %s to %s ...\n", source, sourceDir, sourceBackupArchive)

				// map files on disk to their paths in the archive using default settings (second arg)
				files, err := archives.FilesFromDisk(cmd.Context(), nil, map[string]string{
					sourceDir: "",
				})
				if err != nil {
					return err
				}

				// create the output file we'll write to
				out, err := os.Create(sourceBackupArchive)
				if err != nil {
					return err
				}
				defer out.Close()

				format := archives.CompressedArchive{
					Compression: archives.Gz{},
					Archival:    archives.Tar{},
				}

				// create the archive
				err = format.Archive(cmd.Context(), out, files)
				if err != nil {
					log.Printf("could not backup directory: %q", err)
					return err
				}
			}

			if dryRun {
				fmt.Printf("dry-run: creating qBittorrent backup of directory: %s to %s ...\n", qbitDir, qbitBackupArchive)
			} else {
				fmt.Printf("creating qBittorrent backup of directory: %s to %s ...\n", qbitDir, qbitBackupArchive)

				// map files on disk to their paths in the archive using default settings (second arg)
				files, err := archives.FilesFromDisk(cmd.Context(), nil, map[string]string{
					qbitDir: "",
				})
				if err != nil {
					return err
				}

				// create the output file we'll write to
				out, err := os.Create(qbitBackupArchive)
				if err != nil {
					return err
				}
				defer out.Close()

				format := archives.CompressedArchive{
					Compression: archives.Gz{},
					Archival:    archives.Tar{},
				}

				// create the archive
				err = format.Archive(cmd.Context(), out, files)
				if err != nil {
					log.Printf("could not backup directory: %q", err)
					return err
				}
			}

			fmt.Print("Backup completed!\n")
		}

		start := time.Now()

		if dryRun {
			fmt.Printf("dry-run: preparing to import torrents from: %s dir: %s\n", source, sourceDir)
			fmt.Println("dry-run: no data will be written")
		} else {
			fmt.Printf("preparing to import torrents from: %s dir: %s\n", source, sourceDir)
		}

		opts := importer.Options{
			SourceDir: sourceDir,
			QbitDir:   qbitDir,
			DryRun:    dryRun,
		}

		if err := imp.Import(opts); err != nil {
			fmt.Printf("%s import error: %q\n", source, err)
			os.Exit(1)
		}

		elapsed := time.Since(start)

		fmt.Printf("\nImport finished in: %s\n", elapsed)

		return nil
	}

	return command
}
