package cmd

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ludviglundgren/qbittorrent-cli/internal/fs"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zeebo/bencode"
)

// RunBencode cmd for bencode operations
func RunBencode() *cobra.Command {
	var command = &cobra.Command{
		Use:   "bencode",
		Short: "Bencode subcommand",
		Long:  `Do various bencode operations`,
	}

	command.AddCommand(RunBencodeEdit())

	return command
}

func RunBencodeEdit() *cobra.Command {
	var command = &cobra.Command{
		Use:     "edit",
		Short:   "edit bencode data",
		Long:    "Edit bencode files like .fastresume. Shut down client and make a backup of data before.",
		Example: `  qbt bencode edit /home/user/.local/share/qBittorrent/BT_backup/*.fastresume --save-path="/home/user01/torrents|/home/test/torrents"`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a folder or glob (./*.fastresume) as first argument")
			}

			return nil
		},
	}

	var (
		dry             bool
		verbose         bool
		removeTags      bool
		removeCategory  bool
		exportDir       string
		savePath        string
		replaceCategory string
		replaceTags     []string
	)

	command.Flags().BoolVar(&dry, "dry-run", false, "Dry run, don't write changes")
	command.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	command.Flags().StringVar(&exportDir, "export-dir", "", "Export dir")
	command.Flags().StringVar(&savePath, "save-path", "", "Update save-path. Total or partial replace.")
	command.Flags().BoolVar(&removeTags, "remove-tags", false, "Remove tags")
	command.Flags().StringSliceVar(&replaceTags, "replace-tags", []string{}, "Replace tags. Comma separated")
	command.Flags().BoolVar(&removeCategory, "remove-category", false, "Remove category")
	command.Flags().StringVar(&replaceCategory, "replace-category", "", "Replace category")

	command.RunE = func(cmd *cobra.Command, args []string) error {

		// first arg is dir or glob
		dir := args[0]

		// make sure dir exists before walk
		_, err := os.Stat(filepath.Dir(dir))
		if err != nil {
			if os.IsNotExist(err) {
				return errors.Wrapf(err, "Directory does not exist: %s\n", dir)
			}

			return errors.Wrapf(err, "Directory error: %s\n", dir)
		}

		if exportDir != "" {
			dir := filepath.Dir(exportDir)

			if err := fs.MkDirIfNotExists(dir); err != nil {
				log.Printf("create dir error: %q\n", err)
				return err
			}
		}

		matchedFiles := 0

		files, err := filepath.Glob(dir)
		if err != nil {
			log.Fatal("could not open files\n")
		}

		for _, file := range files {
			fileName := filepath.Base(file)
			path := file

			matchedFiles++

			if dry {
				if verbose {
					log.Printf("dry-run: modified file %s\n", path)

					continue
				}
			} else {
				read, err := os.ReadFile(file)
				if err != nil {
					log.Fatalf("error reading file: %s err: %q\n", path, err)
				}

				var fastResume qbittorrent.Fastresume
				if err := bencode.DecodeBytes(read, &fastResume); err != nil {
					log.Fatalf("could not decode fastresume %s\n", path)
				}

				if savePath != "" {
					if !strings.Contains(savePath, "|") {
						log.Fatalf("save path bad format: (%s) must be old|new\n", savePath)
					}
					parts := strings.Split(savePath, "|")

					if len(parts) == 0 || len(parts) > 2 {
						continue
					}

					fastResume.SavePath = strings.Replace(fastResume.SavePath, parts[0], parts[1], -1)

					if verbose {
						log.Printf("replaced save-path: '%s' with '%s' in %s\n", parts[0], parts[1], path)
					}
				}

				if removeTags {
					fastResume.QbtTags = []string{}
				}

				if len(replaceTags) > 0 {
					fastResume.QbtTags = replaceTags
				}

				if removeCategory {
					fastResume.QbtCategory = ""
				}

				if replaceCategory != "" {
					fastResume.QbtCategory = replaceCategory
				}

				if exportDir != "" {
					path = filepath.Join(exportDir, fileName)
				}

				if err := fastResume.Encode(path); err != nil {
					log.Printf("could not create qBittorrent fastresume file %s error: %q", path, err)
					return err
				}

				if verbose {
					log.Printf("successfully modified file: %s\n", fileName)
				}

			}
		}

		log.Printf("Successfully modified (%d) files\n", matchedFiles)

		return nil
	}

	return command
}
