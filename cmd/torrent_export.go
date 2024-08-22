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
	qbit "github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/autobrr/go-qbittorrent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zeebo/bencode"
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
		if len(f.includeCategory) > 0 && len(f.excludeCategory) > 0 {
			return fmt.Errorf("--include-category and --exclude-category cannot be used together")
		}

		if len(f.includeTag) > 0 && len(f.excludeTag) > 0 {
			return fmt.Errorf("--include-tag and --exclude-tag cannot be used together")
		}

		if _, err := os.Stat(f.sourceDir); err != nil {
			if os.IsNotExist(err) {
				return errors.Wrapf(err, "source dir %s does not exist", f.sourceDir)
			}

			return err
		}

		// get torrents from client by categories
		config.InitConfig()

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
			return errors.Wrapf(err, "failed to login")
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
			return errors.Errorf("Could not find any matching torrents to export from (%s)\n", strings.Join(f.includeCategory, ","))
		}

		log.Printf("Found (%d) matching torrents\n", len(f.hashes))

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
				log.Println("dry-run: successfully wrote manifest to file")
			} else {
				if err := exportManifest(f.hashes, f.tags, f.category); err != nil {
					return errors.Wrapf(err, "could not export manifest")
				}

				log.Println("successfully wrote manifest to file")
			}
		}

		log.Println("Successfully exported torrents!")

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

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "could not get current working directory")
	}

	// Create a new manifestFile in the current working directory.
	manifestFileName := "export-manifest.json"

	manifestFilePath := filepath.Join(currentWorkingDirectory, manifestFileName)

	manifestFile, err := os.Create(manifestFilePath)
	if err != nil {
		return errors.Wrapf(err, "could not create manifestFile: %s", manifestFilePath)
	}
	defer manifestFile.Close()

	if err := json.NewEncoder(manifestFile).Encode(&data); err != nil {
		return errors.Wrap(err, "could not encode manifest to json")
	}

	log.Printf("wrote export manifest to %s", manifestFilePath)

	return nil
}

func processExport(sourceDir, exportDir string, hashes map[string]qbittorrent.Torrent, dry, verbose bool) error {
	exportCount := 0
	exportTorrentCount := 0
	exportFastresumeCount := 0

	// check if export dir exists, if not then lets create it
	if err := createDirIfNotExists(exportDir); err != nil {
		return errors.Wrapf(err, "could not check if dir exists: %s", exportDir)
	}

	// qbittorrent from v4.5.x removes the announce-urls from the .torrent file so we need to add that back
	needTrackerFix := false

	// keep track of processed fastresume files
	processedFastResumeHashes := map[string]bool{}

	// check BT_backup dir, pick torrent and fastresume files by hash
	err := filepath.WalkDir(sourceDir, func(dirPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fileName := d.Name()
		ext := filepath.Ext(fileName)

		if ext != ".torrent" {
			return nil
		}

		torrentHash := fileNameTrimExt(fileName)

		// if filename not in hashes return and check next
		_, ok := hashes[torrentHash]
		if !ok {
			return nil
		}

		if verbose {
			log.Printf("processing: %s\n", fileName)
		}

		if dry {
			exportCount++

			exportTorrentCount++
			log.Printf("dry-run: (%d/%d) exported: %s\n", exportTorrentCount, len(hashes), torrentHash+".torrent")

			exportFastresumeCount++
			log.Printf("dry-run: (%d/%d) exported: %s\n", exportFastresumeCount, len(hashes), torrentHash+".fastresume")

			return nil
		}

		outFile := filepath.Join(exportDir, fileName)

		exportCount++

		// determine if this should be run on first run and the ones after
		if (exportTorrentCount == 0 && !needTrackerFix) || needTrackerFix {

			// open file and check if announce is in there. If it's not, open .fastresume and combine before output
			torrentFile, err := os.Open(dirPath)
			if err != nil {
				return errors.Wrapf(err, "could not open torrent file: %s", dirPath)
			}
			defer torrentFile.Close()

			torrentInfo, err := metainfo.Load(torrentFile)
			if err != nil {
				return errors.Wrapf(err, "could not open file: %s", dirPath)
			}

			if torrentInfo.Announce == "" {
				needTrackerFix = true

				sourceFastResumeFilePath := filepath.Join(sourceDir, torrentHash+".fastresume")
				fastResumeFile, err := os.Open(sourceFastResumeFilePath)
				if err != nil {
					return errors.Wrapf(err, "could not open fastresume file: %s", sourceFastResumeFilePath)
				}
				defer fastResumeFile.Close()

				// open fastresume and get announce then // open fastresume and get announce then <INSERT_CODE_HERE>
				var fastResume qbit.Fastresume
				if err := bencode.NewDecoder(fastResumeFile).Decode(&fastResume); err != nil {
					return errors.Wrapf(err, "could not open file: %s", sourceFastResumeFilePath)
				}

				if len(fastResume.Trackers) == 0 {
					return errors.New("no trackers found in fastresume")
				}

				torrentInfo.Announce = fastResume.Trackers[0][0]
				torrentInfo.AnnounceList = fastResume.Trackers

				if len(torrentInfo.UrlList) == 0 && len(fastResume.UrlList) > 0 {
					torrentInfo.UrlList = fastResume.UrlList
				}

				// copy .fastresume here already since we already have it open
				fastresumeFilePath := filepath.Join(exportDir, torrentHash+".fastresume")
				newFastResumeFile, err := os.Create(fastresumeFilePath)
				if err != nil {
					return errors.Wrapf(err, "could not create new fastresume file: %s", fastresumeFilePath)
				}
				defer newFastResumeFile.Close()

				if err := bencode.NewEncoder(newFastResumeFile).Encode(&fastResume); err != nil {
					return errors.Wrapf(err, "could not encode fastresume to file: %s", fastresumeFilePath)
				}

				// make sure the fastresume is only written once
				processedFastResumeHashes[torrentHash] = true

				exportFastresumeCount++
				log.Printf("[%d/%d] exported: %s\n", exportFastresumeCount, len(hashes), fileName)
			}

			// write new torrent file to destination path
			newTorrentFile, err := os.Create(outFile)
			if err != nil {
				return errors.Wrapf(err, "could not create new torrent file: %s", outFile)
			}
			defer newTorrentFile.Close()

			if err := torrentInfo.Write(newTorrentFile); err != nil {
				return errors.Wrapf(err, "could not write torrent info into file %s", outFile)
			}

			// all good lets return for this file
			exportTorrentCount++
			log.Printf("[%d/%d] exported: %s\n", exportTorrentCount, len(hashes), fileName)

			return nil
		}

		// only do this if !needTrackerFix
		if err := fsutil.CopyFile(dirPath, outFile); err != nil {
			return errors.Wrapf(err, "could not copy file: %s to %s", dirPath, outFile)
		}

		exportTorrentCount++
		log.Printf("[%d/%d] exported: %s\n", exportTorrentCount, len(hashes), fileName)

		// process if fastresume has not already been copied
		_, ok = processedFastResumeHashes[torrentHash]
		if !ok {
			fastResumeFilePath := filepath.Join(exportDir, torrentHash+".fastresume")

			if err := fsutil.CopyFile(dirPath, fastResumeFilePath); err != nil {
				return errors.Wrapf(err, "could not copy file: %s to %s", dirPath, fastResumeFilePath)
			}

			exportFastresumeCount++
			log.Printf("[%d/%d] exported: %s\n", exportFastresumeCount, len(hashes), torrentHash+".fastresume")
		}

		return nil
	})
	if err != nil {
		log.Printf("error reading files: %q\n", err)
		return err
	}

	log.Printf("found (%d) files in total. exported fastresume: %d exported torrent %d", exportCount, exportFastresumeCount, exportTorrentCount)

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
