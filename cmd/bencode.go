package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

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
		Example: `  qbt bencode edit --dir /home/user/.local/share/qBittorrent/BT_backup --pattern '/home/user01/torrents' --replace '/home/test/torrents'`,
	}

	var (
		dry     bool
		verbose bool
		dir     string
		pattern string
		replace string
	)

	command.Flags().BoolVar(&dry, "dry-run", false, "Dry run, don't write changes")
	command.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	command.Flags().StringVar(&dir, "dir", "", "Dir with fast-resume files (required)")
	command.Flags().StringVar(&pattern, "pattern", "", "Pattern to change (required)")
	command.Flags().StringVar(&replace, "replace", "", "Text to replace pattern with (required)")

	command.MarkFlagRequired("dir")
	command.MarkFlagRequired("pattern")
	command.MarkFlagRequired("replace")

	command.RunE = func(cmd *cobra.Command, args []string) error {

		// make sure dir exists before walk
		_, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return errors.Wrapf(err, "Directory does not exist: %s\n", dir)
			}

			return errors.Wrapf(err, "Directory error: %s\n", dir)
		}

		matchedFiles := 0

		err = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			matched, err := filepath.Match("*.fastresume", info.Name())
			if err != nil {
				log.Fatalf("error matching files: %v", err)
			}

			if matched {
				matchedFiles++

				if err := processFastResume(path, pattern, replace, verbose, dry); err != nil {
					log.Fatalf("error processing file: %v", err)
				}
			}

			return nil
		})
		if err != nil {
			log.Fatalf("error reading files: %v", err)
		}

		fmt.Printf("Found, matched and replaced in '%d' files\n", matchedFiles)

		return nil
	}

	return command
}

func processFastResume(path, pattern, replace string, verbose, dry bool) error {
	if dry {
		if verbose {
			fmt.Printf("dry-run: replaced: '%s' with '%s' in %s\n", pattern, replace, path)
		}
	} else {
		read, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("error reading file: %v - %v", path, err)
		}

		var fastResume qbittorrent.Fastresume
		if err := bencode.DecodeString(string(read), &fastResume); err != nil {
			log.Printf("could not decode fastresume %v", path)
		}

		fastResume.SavePath = strings.Replace(fastResume.SavePath, pattern, replace, -1)

		if err = fastResume.Encode(path); err != nil {
			log.Printf("could not create qBittorrent fastresume file %s error: %q", path, err)
			return err
		}

		if verbose {
			fmt.Printf("Replaced: '%s' with '%s' in %s\n", pattern, replace, path)
		}
	}

	return nil
}
