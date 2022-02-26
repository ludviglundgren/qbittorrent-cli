package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/spf13/cobra"
)

// RunMove cmd to move torrents from some category to another
func RunMove() *cobra.Command {
	var (
		dry            bool
		fromCategories []string
		targetCategory string
		includeTags    []string
		minSeedTime    int
	)

	var command = &cobra.Command{
		Use:   "move",
		Short: "move torrents",
		Long:  `Move torrents from some category to another`,
	}

	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().StringSliceVar(&fromCategories, "from", []string{}, "Move from categories")
	command.Flags().StringVar(&targetCategory, "to", "", "Move to the specified category")
	command.Flags().StringSliceVar(&includeTags, "include-tags", []string{}, "Move from tags")
	command.Flags().IntVar(&minSeedTime, "min-seed-time", 0, "Minimum seed time in MINUTES before moving.")
	command.MarkFlagRequired("from")
	command.MarkFlagRequired("to")

	command.Run = func(cmd *cobra.Command, args []string) {
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

		var hashes []string

		for _, cat := range fromCategories {
			torrents, err := qb.GetTorrentsByCategory(cat)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not get torrents by category %v\n", err)
				os.Exit(1)
			}

			for _, torrent := range torrents {
				// only grab completed torrents since we don't specify filter state
				if torrent.Progress != 1 {
					continue
				}

				if len(includeTags) > 0 {
					validTag := validateTag(includeTags, torrent.Tags)
					if !validTag {
						continue
					}
				}

				// check TimeActive (seconds), CompletionOn (epoch) SeenComplete
				if minSeedTime > 0 {
					completedTime := time.Unix(int64(torrent.CompletionOn), 0)
					completedTimePlusMinSeedTime := completedTime.Add(time.Duration(minSeedTime) * time.Minute)
					currentTime := time.Now()

					diff := currentTime.After(completedTimePlusMinSeedTime)
					if !diff {
						continue
					}
				}

				hashes = append(hashes, torrent.Hash)
			}
		}

		if len(hashes) == 0 {
			fmt.Printf("Could not find any matching torrents to move from (%v) to (%v) with tags (%v) and min-seed-time %d minutes \n", strings.Join(fromCategories, ","), targetCategory, strings.Join(includeTags, ","), minSeedTime)
			os.Exit(0)
		}

		if !dry {
			fmt.Printf("Found %d matching torrents to move from (%v) to (%v)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)
			err = qb.SetCategory(hashes, targetCategory)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not pause torrents %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Successfully moved %d torrents from (%v) to (%v)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)
		} else {
			fmt.Printf("DRY-RUN: Found %d matching torrents to move from (%v) to (%v)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)
			fmt.Printf("DRY-RUN: Successfully moved %d torrents from (%v) to (%v)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)
			return
		}
	}

	return command
}

func validateTag(includeTags []string, torrentTags string) bool {
	tagList := strings.Split(torrentTags, ", ")

	for _, includeTag := range includeTags {
		for _, tag := range tagList {
			if tag == includeTag {
				fmt.Printf("Tag match %v\n", tag)
				return true
			}
		}
	}

	return false
}
