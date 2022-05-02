package cmd

import (
	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/spf13/cobra"

	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func RunExport() *cobra.Command {
	var (
		dry        bool
		sourceDir  string
		exportDir  string
		categories []string
		replace    []string
	)

	var command = &cobra.Command{
		Use:   "export",
		Short: "export torrents",
		Long:  "Export torrents and fastresume by category",
	}

	command.Flags().BoolVar(&dry, "dry-run", false, "dry run")
	command.Flags().StringVar(&sourceDir, "source", "", "Dir with torrent and fast-resume files")
	command.Flags().StringVar(&exportDir, "export", "", "Dir to export files to")
	command.Flags().StringSliceVar(&categories, "categories", []string{}, "Export torrents from categories")
	command.Flags().StringSliceVar(&replace, "replace", []string{}, "Replace pattern. old,new")
	command.MarkFlagRequired("categories")

	command.Run = func(cmd *cobra.Command, args []string) {
		// get torrents from client by categories
		config.InitConfig()
		qbtSettings := qbittorrent.Settings{
			Hostname: config.Qbit.Host,
			Port:     config.Qbit.Port,
			Username: config.Qbit.Login,
			Password: config.Qbit.Password,
		}
		qb := qbittorrent.NewClient(qbtSettings)

		err := qb.Login()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		hashes := map[string]struct{}{}

		for _, category := range categories {
			torrents, err := qb.GetTorrentsByCategory(category)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not get torrents by category %v\n", err)
				os.Exit(1)
			}

			for _, torrent := range torrents {
				// only grab completed torrents
				if torrent.Progress != 1 {
					continue
				}

				// append hash to map of hashes to gather
				hashes[torrent.Hash] = struct{}{}
			}
		}

		if len(hashes) == 0 {
			fmt.Printf("Could not find any matching torrents to export from (%v)\n", strings.Join(categories, ","))
			os.Exit(0)
		}

		if err := processHashes(sourceDir, exportDir, hashes, replace, dry); err != nil {
			fmt.Printf("Could not find process files\n")
			os.Exit(0)
		}

	}

	return command
}
func processHashes(sourceDir, exportDir string, hashes map[string]struct{}, replace []string, dry bool) error {
	matchedFiles := 0
	// check BT_backup dir, pick torrent and fastresume files by id
	err := filepath.Walk(sourceDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !!info.IsDir() {
			return nil //
		}

		fileName := info.Name()

		matchedTorrent, err := filepath.Match("*.torrent", fileName)
		if err != nil {
			log.Fatalf("error matching files: %v", err)
		}

		if !matchedTorrent {
			return nil
		}

		// if filename not in hashes return and check next
		_, ok := hashes[fileNameTrimExt(fileName)]
		if !ok {
			return nil
		}

		if !dry {
			err := copyFile(path, filepath.Join(exportDir, fileName), replace)
			if err != nil {
				return err
			}
		}

		matchedFiles++

		return nil
	})
	if err != nil {
		log.Fatalf("error reading files: %v", err)
	}

	fmt.Printf("Found, matched and replaced in '%d' files\n", matchedFiles)

	return nil
}

func fileNameTrimExt(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func copyFile(source, dest string, replace []string) error {
	read, err := ioutil.ReadFile(source)
	if err != nil {
		log.Fatalf("error reading file: %v - %v", source, err)
		return err
	}

	var v metainfo.MetaInfo
	if err := bencode.Unmarshal(read, &v); err != nil {
		log.Printf("could not decode fastresume %v", source)
		return err
	}

	// replace content if needed
	for _, r := range replace {
		// split replace string pattern,replace
		rep := strings.Split(r, ",")

		if v.Announce != "" {
			v.Announce = strings.Replace(v.Announce, rep[0], rep[1], -1)
		}
		for annIdx, a := range v.AnnounceList {
			for i, s := range a {
				v.AnnounceList[annIdx][i] = strings.Replace(s, rep[0], rep[1], -1)
			}
		}
	}

	edited, err := bencode.Marshal(&v)
	if err != nil {
		log.Printf("could not decode fastresume %v", source)
		return err
	}

	// Create new file
	newFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer newFile.Close()

	// save files to new dir
	err = ioutil.WriteFile(dest, edited, 0644)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
