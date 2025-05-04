package cmd

import (
	"log"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/autobrr/go-qbittorrent"
	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RunTorrentCompare cmd to compare torrents between clients
func RunTorrentCompare() *cobra.Command {
	var (
		dry           bool
		tagDuplicates bool
		tag           string

		sourceAddr      string
		sourceUser      string
		sourcePass      string
		sourceBasicUser string
		sourceBasicPass string

		compareAddr      string
		compareUser      string
		comparePass      string
		compareBasicUser string
		compareBasicPass string
	)

	var command = &cobra.Command{
		Use:     "compare",
		Short:   "Compare torrents",
		Long:    `Compare torrents between clients`,
		Example: `  qbt torrent compare --addr http://localhost:10000 --user u --pass p --compare-addr http://url.com:10000 --compare-user u --compare-pass p`,
		//Args: func(cmd *cobra.Command, args []string) error {
		//	if len(args) < 1 {
		//		return errors.New("requires a torrent file as first argument")
		//	}
		//
		//	return nil
		//},
	}
	command.Flags().BoolVar(&dry, "dry-run", false, "dry run")
	command.Flags().BoolVar(&tagDuplicates, "tag-duplicates", false, "tag duplicates on compare")
	command.Flags().StringVar(&tag, "tag", "compare-dupe", "set a custom tag for duplicates on compare. default: compare-dupe")

	command.Flags().StringVar(&sourceAddr, "host", "", "Source host")
	command.Flags().StringVar(&sourceUser, "user", "", "Source user")
	command.Flags().StringVar(&sourcePass, "pass", "", "Source pass")
	command.Flags().StringVar(&sourceBasicUser, "basic-user", "", "Source basic auth user")
	command.Flags().StringVar(&sourceBasicPass, "basic-pass", "", "Source basic auth pass")

	command.Flags().StringVar(&compareAddr, "compare-host", "", "Secondary host")
	command.Flags().StringVar(&compareUser, "compare-user", "", "Secondary user")
	command.Flags().StringVar(&comparePass, "compare-pass", "", "Secondary pass")
	command.Flags().StringVar(&compareBasicUser, "compare-basic-user", "", "Secondary basic auth user")
	command.Flags().StringVar(&compareBasicPass, "compare-basic-pass", "", "Secondary basic auth pass")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		config.InitConfig()

		if sourceAddr == "" {
			sourceAddr = config.Qbit.Host
		}
		if sourceUser == "" {
			sourceUser = config.Qbit.Login
		}
		if sourcePass == "" {
			sourcePass = config.Qbit.Password
		}
		if sourceBasicUser == "" {
			sourceBasicUser = config.Qbit.BasicUser
		}
		if sourceBasicPass == "" {
			sourceBasicPass = config.Qbit.BasicPass
		}

		qbtSettings := qbittorrent.Config{
			Host:      sourceAddr,
			Username:  sourceUser,
			Password:  sourcePass,
			BasicUser: sourceBasicUser,
			BasicPass: sourceBasicPass,
		}
		qb := qbittorrent.NewClient(qbtSettings)

		ctx := cmd.Context()

		if err := qb.LoginCtx(ctx); err != nil {
			return errors.Wrap(err, "could not login to qbit")
		}

		sourceData, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{})
		if err != nil {
			return errors.Wrap(err, "could not get torrents")
		}

		log.Printf("Found: %d torrents on source\n", len(sourceData))

		// Start comparison
		for _, compareConfig := range config.Compare {
			compareAddr := compareConfig.Addr
			compareUser := compareConfig.Login
			comparePass := compareConfig.Password
			compareBasicUser := compareConfig.BasicUser
			compareBasicPass := compareConfig.BasicPass

			qbtSettingsCompare := qbittorrent.Config{
				Host:      compareAddr,
				Username:  compareUser,
				Password:  comparePass,
				BasicUser: compareBasicUser,
				BasicPass: compareBasicPass,
			}
			qbCompare := qbittorrent.NewClient(qbtSettingsCompare)

			if err = qbCompare.LoginCtx(ctx); err != nil {
				return errors.Wrapf(err, "could not login to qbit: %s", compareAddr)
			}

			compareData, err := qbCompare.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{})
			if err != nil {
				return errors.Wrapf(err, "could not get torrents from compare client")
			}

			log.Printf("Found: %d torrents on compare\n", len(compareData))

			duplicateTorrents, err := compare(sourceData, compareData)
			if err != nil {
				return err
			}

			// Process duplicate torrents
			if tagDuplicates {
				if !dry {
					log.Printf("found: %d duplicate torrents from compare %s\n", len(duplicateTorrents), compareAddr)

					batch := 20
					for i := 0; i < len(duplicateTorrents); i += batch {
						j := i + batch
						if j > len(duplicateTorrents) {
							j = len(duplicateTorrents)
						}

						if err := qbCompare.AddTagsCtx(ctx, duplicateTorrents[i:j], tag); err != nil {
							return errors.Wrapf(err, "could not set tag: %s", tag)
						}

						// sleep before next request
						time.Sleep(time.Second * 1)
					}
				} else {
					log.Printf("dry-run: found: %d duplicate torrents from compare %s\n", len(duplicateTorrents), compareAddr)
				}
			}

			// --rm-duplicates

			// --save save to file
		}

		return nil
	}

	return command
}

func compare(source, compare []qbittorrent.Torrent) ([]string, error) {
	sourceTorrents := make(map[string]qbittorrent.Torrent, 0)

	for _, s := range source {
		sourceTorrents[s.Hash] = qbittorrent.Torrent{
			Category:   s.Category,
			Downloaded: s.Downloaded,
			Hash:       s.Hash,
			Name:       s.Name,
			Progress:   s.Progress,
			Ratio:      s.Ratio,
			Size:       s.Size,
			State:      s.State,
			Tags:       s.Tags,
			Tracker:    s.Tracker,
			Uploaded:   s.Uploaded,
		}
	}

	duplicateTorrentIDs := make([]string, 0)
	duplicateTorrentsSlice := make([]qbittorrent.Torrent, 0)

	var totalSize uint64

	for _, c := range compare {
		if _, ok := sourceTorrents[c.Hash]; ok {
			duplicateTorrentIDs = append(duplicateTorrentIDs, c.Hash)

			totalSize += uint64(c.Size)

			duplicateTorrentsSlice = append(duplicateTorrentsSlice, qbittorrent.Torrent{
				Category:   c.Category,
				Downloaded: c.Downloaded,
				Hash:       c.Hash,
				Name:       c.Name,
				Progress:   c.Progress,
				Ratio:      c.Ratio,
				Size:       c.Size,
				State:      c.State,
				Tags:       c.Tags,
				Tracker:    c.Tracker,
				Uploaded:   c.Uploaded,
			})
		}
	}

	// print duplicates
	log.Printf("Found: %d duplicate torrents\n", len(duplicateTorrentsSlice))

	log.Printf("Total reclaimable space: %v\n", humanize.Bytes(totalSize))

	return duplicateTorrentIDs, nil
}
