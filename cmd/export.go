package cmd

import (
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
		sourceDir  string
		exportDir  string
		categories []string
	)

	command.Flags().BoolVar(&dry, "dry-run", false, "dry run")
	command.Flags().StringVar(&sourceDir, "source", "", "Dir with torrent and fast-resume files")
	command.Flags().StringVar(&exportDir, "export-dir", "", "Dir to export files to")
	command.Flags().StringSliceVar(&categories, "categories", []string{}, "Export torrents from categories")
	command.MarkFlagRequired("categories")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		// get torrents from client by categories
		config.InitConfig()

		qbtSettings := qbittorrent.Settings{
			Hostname: config.Qbit.Host,
			Port:     config.Qbit.Port,
			Username: config.Qbit.Login,
			Password: config.Qbit.Password,
			SSL:      config.Qbit.SSL,
		}

		qb := qbittorrent.NewClient(qbtSettings)
		if err := qb.Login(); err != nil {
			return errors.Wrapf(err, "connection failed")
		}

		hashes := map[string]struct{}{}

		for _, category := range categories {
			torrents, err := qb.GetTorrentsByCategory(category)
			if err != nil {
				return errors.Wrapf(err, "could not get torrents for category: %s", category)
			}

			for _, t := range torrents {
				// only grab completed torrents
				if t.Progress != 1 {
					continue
				}

				// append hash to map of hashes to gather
				hashes[t.Hash] = struct{}{}
			}
		}

		if len(hashes) == 0 {
			fmt.Printf("Could not find any matching torrents to export from (%s)\n", strings.Join(categories, ","))
			os.Exit(0)
		}

		fmt.Printf("Found '%d' matching torrents\n", len(hashes))

		if err := processHashes(sourceDir, exportDir, hashes, dry); err != nil {
			return errors.Wrapf(err, "could not process torrents")
		}

		fmt.Println("Successfully exported torrents!")

		return nil
	}

	return command
}
func processHashes(sourceDir, exportDir string, hashes map[string]struct{}, dry bool) error {
	exportCount := 0
	// check BT_backup dir, pick torrent and fastresume files by id
	err := filepath.Walk(sourceDir, func(dirPath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !!info.IsDir() {
			return nil //
		}

		fileName := info.Name()

		if filepath.Ext(fileName) != ".torrent" || filepath.Ext(fileName) != ".fastresume" {
			return nil
		}

		torrentHash := fileNameTrimExt(fileName)

		// if filename not in hashes return and check next
		_, ok := hashes[torrentHash]
		if !ok {
			return nil
		}

		fmt.Printf("processing: %s\n", fileName)

		if !dry {
			outFile := filepath.Join(exportDir, fileName)
			if err := torrent.CopyFile(dirPath, outFile); err != nil {
				return errors.Wrapf(err, "could not copy file: %s to %s", dirPath, outFile)
			}
		}

		exportCount++

		return nil
	})
	if err != nil {
		log.Printf("error reading files: %q", err)
		return err
	}

	fmt.Printf("Found and exported '%d' torrents\n", exportCount)

	return nil
}

func fileNameTrimExt(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
