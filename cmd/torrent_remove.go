package cmd

import (
	"log"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/utils"

	"github.com/autobrr/go-qbittorrent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RunTorrentRemove cmd to remove torrents
func RunTorrentRemove() *cobra.Command {
	var (
		dryRun          bool
		removeAll       bool
		deleteFiles     bool
		hashes          []string
		includeCategory []string
		includeTags     []string
		excludeTags     []string
		filter          string
	)

	var command = &cobra.Command{
		Use:   "remove",
		Short: "Removes specified torrent(s)",
		Long:  `Removes torrents indicated by hash, name or a prefix of either. Whitespace indicates next prefix unless argument is surrounded by quotes`,
	}

	command.Flags().BoolVar(&dryRun, "dry-run", false, "Display what would be done without actually doing it")
	command.Flags().BoolVar(&removeAll, "all", false, "Removes all torrents")
	command.Flags().BoolVar(&deleteFiles, "delete-files", false, "Also delete downloaded files from torrent(s)")
	command.Flags().StringVarP(&filter, "filter", "f", "", "Filter by state: all, active, paused, completed, stalled, errored")
	command.Flags().StringSliceVar(&hashes, "hashes", []string{}, "Add hashes as comma separated list")
	command.Flags().StringSliceVarP(&includeCategory, "include-category", "c", []string{}, "Remove torrents from these categories. Comma separated")
	command.Flags().StringSliceVar(&includeTags, "include-tags", []string{}, "Include torrents with provided tags")
	command.Flags().StringSliceVar(&excludeTags, "exclude-tags", []string{}, "Exclude torrents with provided tags")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		if len(hashes) > 0 {
			if err := utils.ValidateHash(hashes); err != nil {
				return errors.Wrap(err, "invalid hashes supplied")
			}
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

		if removeAll {
			hashes = []string{"all"}
		}

		options := qbittorrent.TorrentFilterOptions{}
		if filter != "" {
			options.Filter = qbittorrent.TorrentFilter(filter)
			if len(includeCategory) == 0 {
				includeCategory = []string{""}
			}
		}

		if len(includeCategory) > 0 {
			for _, category := range includeCategory {
				options.Category = category

				torrents, err := qb.GetTorrentsCtx(ctx, options)
				if err != nil {
					return errors.Wrapf(err, "could not get torrents for category: %s", category)
				}

				for _, torrent := range torrents {
					if len(includeTags) > 0 {
						if _, validTag := validateTag(includeTags, torrent.Tags); !validTag {
							continue
						}
					}

					if len(excludeTags) > 0 {
						if _, found := validateTag(excludeTags, torrent.Tags); found {
							continue
						}
					}

					hashes = append(hashes, torrent.Hash)
				}
			}
		}

		if len(hashes) == 0 {
			log.Println("No torrents found to remove")
			return nil
		}

		if dryRun {
			if hashes[0] == "all" {
				log.Println("dry-run: all torrents to be removed")
			} else {
				log.Printf("dry-run: (%d) torrents to be removed\n", len(hashes))
			}
		} else {
			if hashes[0] == "all" {
				log.Println("all torrents to be removed")
			} else {
				log.Printf("(%d) torrents to be removed\n", len(hashes))
			}

			err := batchRequests(hashes, func(start, end int) error {
				return qb.DeleteTorrentsCtx(ctx, hashes[start:end], deleteFiles)
			})
			if err != nil {
				return errors.Wrap(err, "could not delete torrents")
			}

			if hashes[0] == "all" {
				log.Println("successfully removed all torrents")
			} else {
				log.Printf("successfully removed (%d) torrents\n", len(hashes))
			}
		}

		log.Printf("torrent(s) successfully deleted\n")

		return nil
	}

	return command
}
