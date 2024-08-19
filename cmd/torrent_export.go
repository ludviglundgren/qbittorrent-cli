package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	fsutil "github.com/ludviglundgren/qbittorrent-cli/internal/fs"

	"github.com/autobrr/go-qbittorrent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func RunTorrentExport() *cobra.Command {
	var command = &cobra.Command{
		Use:     "export",
		Short:   "Export torrents",
		Long:    "Export torrents and fastresume by category",
		Example: `  qbt torrent export --source ~/.local/share/data/qBittorrent/BT_backup --export-dir ~/qbt-backup --include-category=movies,tv`,
	}

	f := export{
		dry:             false,
		verbose:         false,
		sourceDir:       "",
		exportDir:       "",
		includeCategory: nil,
		excludeCategory: nil,
		includeTag:      nil,
		excludeTag:      nil,
		tags:            map[string]struct{}{},
		category:        map[string]qbittorrent.Category{},
		hashes:          map[string]qbittorrent.Torrent{},
	}
	var skipManifest bool

	command.Flags().BoolVar(&f.dry, "dry-run", false, "dry run")
	command.Flags().BoolVarP(&f.verbose, "verbose", "v", false, "verbose output")
	command.Flags().BoolVar(&skipManifest, "skip-manifest", false, "Do not export all used tags and categories into manifest")

	command.Flags().StringVar(&f.sourceDir, "source", "", "Dir with torrent and fast-resume files (required)")
	command.Flags().StringVar(&f.exportDir, "export-dir", "", "Dir to export files to (required)")

	command.Flags().StringSliceVar(&f.includeCategory, "include-category", []string{}, "Export torrents from these categories. Comma separated")
	command.Flags().StringSliceVar(&f.excludeCategory, "exclude-category", []string{}, "Exclude categories. Comma separated")

	command.Flags().StringSliceVar(&f.includeTag, "include-tag", []string{}, "Include tags. Comma separated")
	command.Flags().StringSliceVar(&f.excludeTag, "exclude-tag", []string{}, "Exclude tags. Comma separated")

	command.MarkFlagRequired("source")
	command.MarkFlagRequired("export-dir")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		// get torrents from client by categories
		config.InitConfig()

		if _, err := os.Stat(f.sourceDir); err != nil {
			if os.IsNotExist(err) {
				return errors.Wrapf(err, "source dir %s does not exist", f.sourceDir)
			}

			return err
		}

		qbtSettings := qbittorrent.Config{
			Host:      config.Qbit.Addr,
			Username:  config.Qbit.Login,
			Password:  config.Qbit.Password,
			BasicUser: config.Qbit.BasicUser,
			BasicPass: config.Qbit.BasicPass,
		}

		qb := qbittorrent.NewClient(qbtSettings)

		ctx := cmd.Context()

		if err := qb.LoginCtx(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		if len(f.includeCategory) > 0 {
			for _, category := range f.includeCategory {
				torrents, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Category: category})
				if err != nil {
					return errors.Wrapf(err, "could not get torrents for category: %s", category)
				}

				for _, tor := range torrents {
					// only grab completed torrents
					//if tor.Progress != 1 {
					//	continue
					//}

					if tor.Tags != "" {
						tags := strings.Split(tor.Tags, ", ")

						// check tags and exclude categories
						if len(f.includeTag) > 0 && !containsTag(f.includeTag, tags) {
							continue
						}

						if len(f.excludeTag) > 0 && containsTag(f.excludeTag, tags) {
							continue
						}

						for _, tag := range tags {
							_, ok := f.tags[tag]
							if !ok {
								f.tags[tag] = struct{}{}
							}
						}

					}

					if tor.Category != "" {
						f.category[tor.Category] = qbittorrent.Category{
							Name:     tor.Category,
							SavePath: "",
						}
					}

					// append hash to map of hashes to gather
					f.hashes[strings.ToLower(tor.Hash)] = tor
				}
			}

		} else {
			torrents, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{})
			if err != nil {
				return errors.Wrap(err, "could not get torrents")
			}

			for _, tor := range torrents {
				// only grab completed torrents
				//if tor.Progress != 1 {
				//	continue
				//}

				if len(f.excludeCategory) > 0 && containsCategory(f.excludeCategory, tor.Category) {
					continue
				}

				if tor.Tags != "" {
					tags := strings.Split(tor.Tags, ", ")

					// check tags and exclude categories
					if len(f.includeTag) > 0 && !containsTag(f.includeTag, tags) {
						continue
					}

					if len(f.excludeTag) > 0 && containsTag(f.excludeTag, tags) {
						continue
					}

					for _, tag := range tags {
						_, ok := f.tags[tag]
						if !ok {
							f.tags[tag] = struct{}{}
						}
					}
				}

				if tor.Category != "" {
					f.category[tor.Category] = qbittorrent.Category{
						Name:     tor.Category,
						SavePath: "",
					}
				}

				// append hash to map of hashes to gather
				f.hashes[strings.ToLower(tor.Hash)] = tor
			}
		}

		if len(f.hashes) == 0 {
			fmt.Printf("Could not find any matching torrents to export from (%s)\n", strings.Join(f.includeCategory, ","))
			os.Exit(1)
		}

		fmt.Printf("Found '%d' matching torrents\n", len(f.hashes))

		if err := processExport(f.sourceDir, f.exportDir, f.hashes, f.dry, f.verbose); err != nil {
			return errors.Wrapf(err, "could not process torrents")
		}

		// write export manifest with categories and tags
		// can be used for import later on
		if !skipManifest {
			// get categories
			if len(f.category) > 0 {
				cats, err := qb.GetCategoriesCtx(ctx)
				if err != nil {
					return errors.Wrapf(err, "could not get categories from qbit")
				}

				for name, category := range cats {
					_, ok := f.category[name]
					if !ok {
						continue
					}

					f.category[name] = category
				}
			}

			if f.dry {
				fmt.Println("dry-run: successfully wrote manifest to file")
			} else {
				if err := exportManifest(f.hashes, f.tags, f.category); err != nil {
					fmt.Printf("could not export manifest: %q\n", err)
					os.Exit(1)
				}

				fmt.Println("successfully wrote manifest to file")
			}
		}

		fmt.Println("Successfully exported torrents!")

		return nil
	}

	return command
}

func exportManifest(hashes map[string]qbittorrent.Torrent, tags map[string]struct{}, categories map[string]qbittorrent.Category) error {
	data := Manifest{
		Tags:       make([]string, 0),
		Categories: []qbittorrent.Category{},
		Torrents:   make([]basicTorrent, 0),
	}

	for tag, _ := range tags {
		data.Tags = append(data.Tags, tag)
	}

	for _, category := range categories {
		data.Categories = append(data.Categories, category)
	}

	for _, torrent := range hashes {
		data.Torrents = append(data.Torrents, basicTorrent{
			Hash:     torrent.Hash,
			Name:     torrent.Name,
			Tags:     torrent.Tags,
			Category: torrent.Category,
			Tracker:  torrent.Tracker,
		})
	}

	res, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "could not marshal manifest to json")
	}

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}

	// Create a new file in the current working directory.
	fileName := "export-manifest.json"

	file, err := os.Create(filepath.Join(currentWorkingDirectory, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the string to the file.
	_, err = file.WriteString(string(res))
	if err != nil {
		return err
	}

	return nil
}

func processExport(sourceDir, exportDir string, hashes map[string]qbittorrent.Torrent, dry, verbose bool) error {
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
				fmt.Printf("dry-run: (%d/%d) exported: %s\n", exportTorrentCount, len(hashes), fileName)
			case ".fastresume":
				exportFastresumeCount++
				fmt.Printf("dry-run: (%d/%d) exported: %s\n", exportFastresumeCount, len(hashes), fileName)
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
				fmt.Printf("(%d/%d) exported: %s\n", exportTorrentCount, len(hashes), fileName)
			case ".fastresume":
				exportFastresumeCount++
				fmt.Printf("(%d/%d) exported: %s\n", exportFastresumeCount, len(hashes), fileName)
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
	includeCategory []string
	excludeCategory []string
	includeTag      []string
	excludeTag      []string

	tags     map[string]struct{}
	category map[string]qbittorrent.Category

	hashes map[string]qbittorrent.Torrent
}

func containsTag(contains []string, tags []string) bool {
	for _, s := range tags {
		s = strings.ToLower(s)
		for _, contain := range contains {
			if s == strings.ToLower(contain) {
				return true
			}
		}
	}

	return false
}

func containsCategory(contains []string, category string) bool {
	for _, cat := range contains {
		if strings.ToLower(category) == strings.ToLower(cat) {
			return true
		}
	}

	return false
}

type basicTorrent struct {
	Hash     string `json:"hash"`
	Name     string `json:"name"`
	Tags     string `json:"tags"`
	Category string `json:"category"`
	Tracker  string `json:"tracker"`
}

type Manifest struct {
	Tags       []string               `json:"tags"`
	Categories []qbittorrent.Category `json:"categories"`
	Torrents   []basicTorrent         `json:"torrents"`
}
