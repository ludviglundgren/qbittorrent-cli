package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/torrent"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func RunExport() *cobra.Command {
	var command = &cobra.Command{
		Use:   "export",
		Short: "export torrents",
		Long:  "Export torrents and fastresume by category",
	}

	var (
		dry        bool
		verbose    bool
		sourceDir  string
		exportDir  string
		categories []string
	)

	command.Flags().BoolVar(&dry, "dry-run", false, "dry run")
	command.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	command.Flags().StringVar(&sourceDir, "source", "", "Dir with torrent and fast-resume files")
	command.Flags().StringVar(&exportDir, "export-dir", "", "Dir to export files to")
	command.Flags().StringSliceVar(&categories, "categories", []string{}, "Export torrents from categories. Comma separated")
	command.MarkFlagRequired("categories")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		// get torrents from client by categories
		config.InitConfig()

		if _, err := os.Stat(sourceDir); err != nil {
			if os.IsNotExist(err) {
				return errors.Wrapf(err, "source dir %s does not exist", sourceDir)
			}

			return err
		}

		qbtSettings := qbittorrent.Settings{
			Addr:      config.Qbit.Addr,
			Hostname:  config.Qbit.Host,
			Port:      config.Qbit.Port,
			Username:  config.Qbit.Login,
			Password:  config.Qbit.Password,
			BasicUser: config.Qbit.BasicUser,
			BasicPass: config.Qbit.BasicPass,
		}

		qb := qbittorrent.NewClient(qbtSettings)

		ctx := context.Background()

		if err := qb.Login(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}
		hashes := map[string]struct{}{}

		for _, category := range categories {
			torrents, err := qb.GetTorrentsWithFilters(ctx, &qbittorrent.GetTorrentsRequest{Category: category})
			if err != nil {
				return errors.Wrapf(err, "could not get torrents for category: %s", category)
			}

			for _, t := range torrents {
				// only grab completed torrents
				if t.Progress != 1 {
					continue
				}

				// append hash to map of hashes to gather
				hashes[strings.ToLower(t.Hash)] = struct{}{}
			}
		}

		if len(hashes) == 0 {
			fmt.Printf("Could not find any matching torrents to export from (%s)\n", strings.Join(categories, ","))
			os.Exit(1)
		}

		fmt.Printf("Found '%d' matching torrents\n", len(hashes))

		// check if export dir exists, if not then lets create it
		if err := createDirIfNotExists(exportDir); err != nil {
			fmt.Printf("could not check if dir %s exists. err: %q\n", exportDir, err)
			return errors.Wrapf(err, "could not check if dir exists: %s", exportDir)
			//os.Exit(1)
		}

		if err := processExport(sourceDir, exportDir, hashes, dry, verbose); err != nil {
			return errors.Wrapf(err, "could not process torrents")
		}

		fmt.Println("Successfully exported torrents!")

		return nil
	}

	return command
}

func processExport(sourceDir, exportDir string, hashes map[string]struct{}, dry, verbose bool) error {
	exportCount := 0
	exportTorrentCount := 0
	exportFastresumeCount := 0

	// check BT_backup dir, pick torrent and fastresume files by id
	err := filepath.Walk(sourceDir, func(dirPath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileName := info.Name()
		ext := filepath.Ext(fileName)

		if !isValidExt(ext) {
			return nil
		}

		torrentHash := fileNameTrimExt(fileName)

		// if filename not in hashes return and check next
		_, ok := hashes[torrentHash]
		if !ok {
			return nil
		}

		if dry {
			if verbose {
				fmt.Printf("processing: %s\n", fileName)
			}

			exportCount++

			//fmt.Printf("dry-run: (%d/%d) exported: %s '%s'\n", exportCount, len(hashes), torrentHash, fileName)

			if ext == ".torrent" {
				exportTorrentCount++
				fmt.Printf("dry-run: (%d/%d) torrent exported: %s '%s'\n", exportTorrentCount, len(hashes), torrentHash, fileName)
			} else if ext == ".fastresume" {
				exportFastresumeCount++
				fmt.Printf("dry-run: (%d/%d) fastresume exported: %s '%s'\n", exportFastresumeCount, len(hashes), torrentHash, fileName)
			}

		} else {
			if verbose {
				fmt.Printf("processing: %s\n", fileName)
			}

			outFile := filepath.Join(exportDir, fileName)
			if err := torrent.CopyFile(dirPath, outFile); err != nil {
				return errors.Wrapf(err, "could not copy file: %s to %s", dirPath, outFile)
			}

			exportCount++
			if ext == ".torrent" {
				exportTorrentCount++
				fmt.Printf("(%d/%d) torrent exported: %s '%s'\n", exportTorrentCount, len(hashes), torrentHash, fileName)
			} else if ext == ".fastresume" {
				exportFastresumeCount++
				fmt.Printf("(%d/%d) fastresume exported: %s '%s'\n", exportFastresumeCount, len(hashes), torrentHash, fileName)
			}

			//fmt.Printf("(%d/%d) exported: %s '%s'\n", exportCount, len(hashes), torrentHash, fileName)
		}

		return nil
	})
	if err != nil {
		log.Printf("error reading files: %q\n", err)
		return err
	}

	fmt.Printf(`
found (%d) files in total
exported fastresume: %d
exported torrent %d
`, exportCount, exportFastresumeCount, exportTorrentCount)

	return nil
}

func fileNameTrimExt(fileName string) string {
	return strings.ToLower(strings.TrimSuffix(fileName, filepath.Ext(fileName)))
}

// isValidExt check if the input ext is one of the ext we want
func isValidExt(filename string) bool {
	valid := []string{".torrent", ".fastresume"}

	for _, s := range valid {
		if s == filename {
			return true
		}
	}

	return false
}

// createDirIfNotExists check if export dir exists, if not then lets create it
func createDirIfNotExists(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}

			return nil
		}

		return err
	}

	return nil
}
