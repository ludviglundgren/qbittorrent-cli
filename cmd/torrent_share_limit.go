package cmd

import (
	"log"
	"strconv"
	"strings"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/utils"

	"github.com/autobrr/go-qbittorrent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RunTorrentShareLimit cmd for torrent share limit operations
func RunTorrentShareLimit() *cobra.Command {
	var command = &cobra.Command{
		Use:   "share-limit",
		Short: "Torrent share limit subcommand",
		Long:  `Do various torrent share limit operations`,
	}

	command.AddCommand(RunTorrentShareLimitSet())

	return command
}

// RunTorrentShareLimitSet cmd to set share limits for torrents
func RunTorrentShareLimitSet() *cobra.Command {
	var (
		dry                      bool
		shareAll                 bool
		hashes                   []string
		includeCategory          []string
		includeTags              []string
		excludeTags              []string
		ratioLimit               float64
		seedingTimeLimit         int64
		inactiveSeedingTimeLimit int64
	)

	var command = &cobra.Command{
		Use:   "set",
		Short: "Set torrent share limits",
		Long: `Set share limits (ratio, seeding time, inactive seeding time) for torrents.

Limit values use qBittorrent's special semantics:
  -2   use the global limit (default)
  -1   no limit (unlimited)
  >=0  a specific value (ratio, or minutes for the seeding time limits)

qBittorrent applies all three limits in a single request, so any limit you do not
set is reset to the global limit (-2). The applied values are printed when the
command runs.

Note: on qBittorrent 5.x (Web API 2.12+) this also resets each torrent's share
limit action and mode to their defaults.`,
		Example: `  qbt torrent share-limit set --hashes hash1,hash2 --ratio 2.0
  qbt torrent share-limit set --all --seeding-time 1440
  qbt torrent share-limit set --include-category movies --ratio 1.5 --seeding-time 10080
  qbt torrent share-limit set --hashes hash1 --ratio -1`,
	}

	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().BoolVar(&shareAll, "all", false, "Set share limits for all torrents")
	command.Flags().StringSliceVar(&hashes, "hashes", []string{}, "Torrent hashes, as comma separated list")
	command.Flags().StringSliceVarP(&includeCategory, "include-category", "c", []string{}, "Set share limits for torrents in these categories. Comma separated")
	command.Flags().StringSliceVar(&includeTags, "include-tags", []string{}, "Include torrents with any of these tags. Comma separated")
	command.Flags().StringSliceVar(&excludeTags, "exclude-tags", []string{}, "Exclude torrents with any of these tags. Comma separated")
	command.Flags().Float64Var(&ratioLimit, "ratio", -2, "Ratio limit. -2 = global, -1 = unlimited, >=0 = ratio")
	command.Flags().Int64Var(&seedingTimeLimit, "seeding-time", -2, "Seeding time limit in MINUTES. -2 = global, -1 = unlimited, >=0 = minutes")
	command.Flags().Int64Var(&inactiveSeedingTimeLimit, "inactive-seeding-time", -2, "Inactive seeding time limit in MINUTES. -2 = global, -1 = unlimited, >=0 = minutes")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		// require at least one limit to be set so we don't silently reset
		// every limit to the global value
		if !cmd.Flags().Changed("ratio") && !cmd.Flags().Changed("seeding-time") && !cmd.Flags().Changed("inactive-seeding-time") {
			return errors.New("you must set at least one limit: --ratio, --seeding-time or --inactive-seeding-time")
		}

		if err := validateShareLimits(ratioLimit, seedingTimeLimit, inactiveSeedingTimeLimit); err != nil {
			return err
		}

		// require at least one target
		if !shareAll && len(hashes) == 0 && len(includeCategory) == 0 && len(includeTags) == 0 {
			return errors.New("no torrents specified. Use --hashes, --all, --include-category or --include-tags")
		}

		if len(hashes) > 0 {
			if err := utils.ValidateHash(hashes); err != nil {
				return errors.Wrap(err, "invalid hashes supplied")
			}
		}

		config.InitConfig()

		qbtSettings := qbittorrent.Config{
			Host:      config.Qbit.Addr,
			APIKey:    config.Qbit.APIKey,
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

		// resolve target hashes
		if shareAll {
			hashes = []string{"all"}
		} else if len(includeCategory) > 0 || len(includeTags) > 0 {
			// when only tags are provided, fetch all torrents (empty category)
			// and filter them down by tag
			if len(includeCategory) == 0 {
				includeCategory = []string{""}
			}

			// append category/tag matches to any explicitly provided hashes
			// so the two selection methods union instead of overwrite
			for _, category := range includeCategory {
				torrents, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Category: category})
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
			log.Println("No torrents found to set share limits on")
			return nil
		}

		opts := qbittorrent.ShareLimitOptions{
			RatioLimit:               ratioLimit,
			SeedingTimeLimit:         seedingTimeLimit,
			InactiveSeedingTimeLimit: inactiveSeedingTimeLimit,
		}

		limitsDesc := formatShareLimits(opts)

		target := strconv.Itoa(len(hashes)) + " torrent(s)"
		if shareAll {
			target = "all torrents"
		}

		if dry {
			log.Printf("dry-run: would set share limits (%s) on %s\n", limitsDesc, target)
			return nil
		}

		err := batchRequests(hashes, func(start, end int) error {
			return qb.SetTorrentShareLimitCtx(ctx, hashes[start:end], opts)
		})
		if err != nil {
			return errors.Wrap(err, "could not set share limits")
		}

		log.Printf("successfully set share limits (%s) on %s\n", limitsDesc, target)

		return nil
	}

	return command
}

// validateShareLimits checks that each limit is within qBittorrent's accepted
// range: -2 (global), -1 (unlimited) or any value >= 0.
func validateShareLimits(ratioLimit float64, seedingTimeLimit, inactiveSeedingTimeLimit int64) error {
	// ratio is a float, so a simple `< -2` check would let fractional
	// sentinels like -1.5 or -0.5 through; only -2, -1 and >= 0 are valid
	if ratioLimit < 0 && ratioLimit != -1 && ratioLimit != -2 {
		return errors.Errorf("invalid ratio limit %s: must be -2 (global), -1 (unlimited) or >= 0", strconv.FormatFloat(ratioLimit, 'f', -1, 64))
	}

	if seedingTimeLimit < -2 {
		return errors.Errorf("invalid seeding-time limit %d: must be -2 (global), -1 (unlimited) or >= 0", seedingTimeLimit)
	}

	if inactiveSeedingTimeLimit < -2 {
		return errors.Errorf("invalid inactive-seeding-time limit %d: must be -2 (global), -1 (unlimited) or >= 0", inactiveSeedingTimeLimit)
	}

	return nil
}

// formatShareLimits builds a human readable description of the share limits.
func formatShareLimits(opts qbittorrent.ShareLimitOptions) string {
	return strings.Join([]string{
		"ratio=" + formatRatioLimit(opts.RatioLimit),
		"seeding-time=" + formatTimeLimit(opts.SeedingTimeLimit),
		"inactive-seeding-time=" + formatTimeLimit(opts.InactiveSeedingTimeLimit),
	}, ", ")
}

func formatRatioLimit(ratioLimit float64) string {
	switch ratioLimit {
	case -2:
		return "global"
	case -1:
		return "unlimited"
	default:
		return strconv.FormatFloat(ratioLimit, 'f', 2, 64)
	}
}

func formatTimeLimit(minutes int64) string {
	switch minutes {
	case -2:
		return "global"
	case -1:
		return "unlimited"
	default:
		return strconv.FormatInt(minutes, 10) + "m"
	}
}
