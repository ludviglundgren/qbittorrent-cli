package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/autobrr/go-qbittorrent"
	"github.com/spf13/cobra"
)

// RunTorrentCategory cmd for torrent category operations
func RunTorrentCategory() *cobra.Command {
	var command = &cobra.Command{
		Use:   "category",
		Short: "Torrent category subcommand",
		Long:  `Do various torrent category operations`,
	}

	command.AddCommand(RunTorrentCategoryChange())
	command.AddCommand(RunTorrentCategorySet())
	command.AddCommand(RunTorrentCategoryUnSet())

	return command
}

// RunTorrentCategorySet cmd for torrent category operations
func RunTorrentCategorySet() *cobra.Command {
	var command = &cobra.Command{
		Use:   "set",
		Short: "Set torrent category",
		Long:  `Do various torrent category operations`,
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	return command
}

// RunTorrentCategoryUnSet cmd for torrent category operations
func RunTorrentCategoryUnSet() *cobra.Command {
	var command = &cobra.Command{
		Use:   "unset",
		Short: "unset torrent category",
		Long:  `Do various torrent category operations`,
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	return command
}

// RunTorrentCategoryChange cmd to move torrents from some category to another
func RunTorrentCategoryChange() *cobra.Command {
	var (
		dry            bool
		fromCategories []string
		targetCategory string
		includeTags    []string
		excludeTags    []string
		minSeedTime    int
	)

	var command = &cobra.Command{
		Use:     "move",
		Short:   "move torrents between categories",
		Long:    `Move torrents from one category to another`,
		Example: `  qbt torrent category move --from cat1 --to cat2`,
	}

	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().StringSliceVar(&fromCategories, "from", []string{}, "Move from categories (required)")
	command.Flags().StringVar(&targetCategory, "to", "", "Move to the specified category (required)")
	command.Flags().StringSliceVar(&includeTags, "include-tags", []string{}, "Include torrents with provided tags")
	command.Flags().StringSliceVar(&excludeTags, "exclude-tags", []string{}, "Exclude torrents with provided tags")
	command.Flags().IntVar(&minSeedTime, "min-seed-time", 0, "Minimum seed time in MINUTES before moving.")

	command.MarkFlagRequired("from")
	command.MarkFlagRequired("to")

	command.Run = func(cmd *cobra.Command, args []string) {
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
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		var hashes []string

		for _, cat := range fromCategories {
			torrents, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Category: cat})
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
					if _, validTag := validateTag(includeTags, torrent.Tags); !validTag {
						continue
					}
				}

				if len(excludeTags) > 0 {
					if tag, found := validateTag(excludeTags, torrent.Tags); found {
						fmt.Printf("ignoring torrent %s %s containng tag: %s of tags: %s", torrent.Name, torrent.Hash, tag, excludeTags)
						continue
					}
				}

				// check TimeActive (seconds), CompletionOn (epoch) SeenComplete
				if minSeedTime > 0 {
					completedTime := time.Unix(torrent.CompletionOn, 0)
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
			fmt.Printf("Could not find any matching torrents to move from (%s) to (%s) with tags (%s) and min-seed-time %d minutes\n", strings.Join(fromCategories, ","), targetCategory, strings.Join(includeTags, ","), minSeedTime)
			os.Exit(0)
		}

		if dry {
			fmt.Printf("dry-run: Found %d matching torrents to move from (%s) to (%s)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)
			fmt.Printf("dry-run: Successfully moved %d torrents from (%s) to (%s)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)
		} else {
			fmt.Printf("Found %d matching torrents to move from (%s) to (%s)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)

			if err := qb.SetCategoryCtx(ctx, hashes, targetCategory); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not pause torrents %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Successfully moved %d torrents from (%s) to (%s)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)
		}
	}

	return command
}

func validateTag(includeTags []string, torrentTags string) (string, bool) {
	tagList := strings.Split(torrentTags, ", ")

	for _, includeTag := range includeTags {
		for _, tag := range tagList {
			if tag == includeTag {
				return tag, true
			}
		}
	}

	return "", false
}
