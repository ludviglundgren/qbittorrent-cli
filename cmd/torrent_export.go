package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	fsutil "github.com/ludviglundgren/qbittorrent-cli/internal/fs"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func RunExport() *cobra.Command {
	var command = &cobra.Command{
		Use:   "export",
		Short: "export torrents",
		Long:  "Export torrents and fastresume by category",
	}

	f := export{
		dry:             false,
		verbose:         false,
		sourceDir:       "",
		exportDir:       "",
		categories:      nil,
		includeCategory: nil,
		excludeCategory: nil,
		includeTag:      nil,
		excludeTag:      nil,
		hashes:          map[string]struct{}{},
	}

	command.Flags().BoolVar(&f.dry, "dry-run", false, "dry run")
	command.Flags().BoolVarP(&f.verbose, "verbose", "v", false, "verbose output")
	command.Flags().StringVar(&f.sourceDir, "source", "", "Dir with torrent and fast-resume files")
	command.Flags().StringVar(&f.exportDir, "export-dir", "", "Dir to export files to")
	command.Flags().StringSliceVar(&f.categories, "categories", []string{}, "Export torrents from categories. Comma separated")
	command.Flags().StringSliceVar(&f.includeCategory, "include-category", []string{}, "Include categories. Comma separated")

	command.MarkFlagRequired("categories")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		// get torrents from client by categories
		config.InitConfig()

		if _, err := os.Stat(f.sourceDir); err != nil {
			if os.IsNotExist(err) {
				return errors.Wrapf(err, "source dir %s does not exist", f.sourceDir)
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

		if err := qb.Login(cmd.Context()); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		if len(f.includeCategory) > 0 {
			for _, category := range f.categories {
				torrents, err := qb.GetTorrentsWithFilters(cmd.Context(), &qbittorrent.GetTorrentsRequest{Category: category})
				if err != nil {
					return errors.Wrapf(err, "could not get torrents for category: %s", category)
				}

				for _, t := range torrents {
					// only grab completed torrents
					if t.Progress != 1 {
						continue
					}

					// todo check tags

					// append hash to map of hashes to gather
					f.hashes[strings.ToLower(t.Hash)] = struct{}{}
				}
			}

		} else {
			torrents, err := qb.GetTorrents(cmd.Context())
			if err != nil {
				return errors.Wrap(err, "could not get torrents")
			}

			for _, t := range torrents {
				// only grab completed torrents
				if t.Progress != 1 {
					continue
				}

				// todo check tags and exclude categories

				// append hash to map of hashes to gather
				f.hashes[strings.ToLower(t.Hash)] = struct{}{}
			}

		}

		if len(f.hashes) == 0 {
			fmt.Printf("Could not find any matching torrents to export from (%s)\n", strings.Join(f.categories, ","))
			os.Exit(1)
		}

		fmt.Printf("Found '%d' matching torrents\n", len(f.hashes))

		if err := processExport(f.sourceDir, f.exportDir, f.hashes, f.dry, f.verbose); err != nil {
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

	// check if export dir exists, if not then lets create it
	if err := createDirIfNotExists(exportDir); err != nil {
		fmt.Printf("could not check if dir %s exists. err: %q\n", exportDir, err)
		return errors.Wrapf(err, "could not check if dir exists: %s", exportDir)
	}

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

			switch ext {
			case ".torrent":
				exportTorrentCount++
				fmt.Printf("dry-run: (%d/%d) torrent exported: %s '%s'\n", exportTorrentCount, len(hashes), torrentHash, fileName)
			case ".fastresume":
				exportFastresumeCount++
				fmt.Printf("dry-run: (%d/%d) fastresume exported: %s '%s'\n", exportFastresumeCount, len(hashes), torrentHash, fileName)
			}

		} else {
			if verbose {
				fmt.Printf("processing: %s\n", fileName)
			}

			outFile := filepath.Join(exportDir, fileName)
			if err := fsutil.CopyFile(dirPath, outFile); err != nil {
				return errors.Wrapf(err, "could not copy file: %s to %s", dirPath, outFile)
			}

			exportCount++

			switch ext {
			case ".torrent":
				exportTorrentCount++
				fmt.Printf("(%d/%d) torrent exported: %s '%s'\n", exportTorrentCount, len(hashes), torrentHash, fileName)
			case ".fastresume":
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

type export struct {
	dry             bool
	verbose         bool
	sourceDir       string
	exportDir       string
	categories      []string
	includeCategory []string
	excludeCategory []string
	includeTag      []string
	excludeTag      []string

	hashes map[string]struct{}
}
