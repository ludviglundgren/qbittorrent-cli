package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/utils"

	"github.com/autobrr/go-qbittorrent"
	"github.com/pkg/errors"
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
		Use:     "set",
		Short:   "Set torrent category",
		Long:    `Set category for torrents via hashes`,
		Example: `  qbt torrent category set test-category --hashes hash1,hash2`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a category as first argument")
			}

			return nil
		},
	}

	var (
		dry    bool
		hashes []string
	)

	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().StringSliceVar(&hashes, "hashes", []string{}, "Torrent hashes, as comma separated list")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		if len(hashes) == 0 {
			return errors.New("no hashes supplied!")
		}

		err := utils.ValidateHash(hashes)
		if err != nil {
			return errors.Wrap(err, "invalid hashes supplied")
		}

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
			return errors.Wrap(err, "could not login to qbit")
		}

		// args
		// first arg is path to torrent file
		category := args[0]

		if dry {
			log.Printf("dry-run: successfully set category %s on torrents: %v\n", category, hashes)

			return nil

		} else {
			if err := qb.SetCategoryCtx(ctx, hashes, category); err != nil {
				return errors.Wrapf(err, "could not set category %s on torrents %v", category, hashes)
			}

			log.Printf("successfully set category %s on torrents: %v\n", category, hashes)
		}

		return nil
	}

	return command
}

// RunTorrentCategoryUnSet cmd for torrent category operations
func RunTorrentCategoryUnSet() *cobra.Command {
	var command = &cobra.Command{
		Use:     "unset",
		Short:   "Unset torrent category",
		Long:    `Unset category for torrents via hashes`,
		Example: `  qbt torrent category unset --hashes hash1,hash2`,
	}

	var (
		dry    bool
		hashes []string
	)

	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().StringSliceVar(&hashes, "hashes", []string{}, "Torrent hashes, as comma separated list")

	command.RunE = func(cmd *cobra.Command, args []string) error {
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
			return errors.Wrap(err, "could not login to qbit")
		}

		if dry {
			log.Printf("dry-run: successfully unset category on torrents: %v\n", hashes)

			return nil

		} else {
			if err := qb.SetCategoryCtx(ctx, hashes, ""); err != nil {
				return errors.Wrapf(err, "could not unset category on torrents %v", hashes)
			}

			log.Printf("successfully unset category on torrents: %v\n", hashes)
		}

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

	command.RunE = func(cmd *cobra.Command, args []string) error {
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
			return errors.Wrap(err, "could not login to qbit")
		}

		var hashes []string

		for _, cat := range fromCategories {
			torrents, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Category: cat})
			if err != nil {
				return errors.Wrapf(err, "could not get torrents by category: %s", cat)
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
						log.Printf("ignoring torrent %s %s containng tag: %s of tags: %s", torrent.Name, torrent.Hash, tag, excludeTags)
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
			log.Printf("Could not find any matching torrents to move from (%s) to (%s) with tags (%s) and min-seed-time %d minutes\n", strings.Join(fromCategories, ","), targetCategory, strings.Join(includeTags, ","), minSeedTime)
			return nil
		}

		if dry {
			log.Printf("dry-run: Found %d matching torrents to move from (%s) to (%s)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)
			log.Printf("dry-run: Successfully moved %d torrents from (%s) to (%s)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)
		} else {
			log.Printf("Found %d matching torrents to move from (%s) to (%s)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)

			if err := qb.SetCategoryCtx(ctx, hashes, targetCategory); err != nil {
				return errors.Wrapf(err, "could not pause torrents: %v", hashes)
			}

			log.Printf("Successfully moved %d torrents from (%s) to (%s)\n", len(hashes), strings.Join(fromCategories, ","), targetCategory)
		}

		return nil
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
