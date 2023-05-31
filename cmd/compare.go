package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// RunCompare cmd to compare torrents between clients
func RunCompare() *cobra.Command {
	var (
		tagDuplicates bool
		overrideTag   string

		sourceHost string
		sourcePort uint
		sourceUser string
		sourcePass string

		compareHost string
		comparePort uint
		compareUser string
		comparePass string
	)

	var command = &cobra.Command{
		Use:   "compare",
		Short: "Compare torrents",
		Long:  `Compare torrents between clients`,
		//Args: func(cmd *cobra.Command, args []string) error {
		//	if len(args) < 1 {
		//		return errors.New("requires a torrent file as first argument")
		//	}
		//
		//	return nil
		//},
	}
	command.Flags().BoolVar(&tagDuplicates, "tag-duplicates", false, "tag duplicates on compare")
	command.Flags().StringVar(&overrideTag, "tag", "", "set a custom tag for duplicates on compare")

	command.Flags().StringVar(&sourceHost, "host", "", "Source host")
	command.Flags().UintVar(&sourcePort, "port", 0, "Source host")
	command.Flags().StringVar(&sourceUser, "user", "", "Source user")
	command.Flags().StringVar(&sourcePass, "pass", "", "Source pass")

	command.Flags().StringVar(&compareHost, "compare-host", "", "Secondary host")
	command.Flags().UintVar(&comparePort, "compare-port", 0, "Secondary host")
	command.Flags().StringVar(&compareUser, "compare-user", "", "Secondary user")
	command.Flags().StringVar(&comparePass, "compare-pass", "", "Secondary pass")

	command.Run = func(cmd *cobra.Command, args []string) {
		config.InitConfig()

		if sourceHost == "" {
			sourceHost = config.Qbit.Host
		}
		if sourcePort == 0 {
			sourcePort = config.Qbit.Port
		}
		if sourceUser == "" {
			sourceUser = config.Qbit.Login
		}
		if sourcePass == "" {
			sourcePass = config.Qbit.Password
		}

		qbtSettings := qbittorrent.Settings{
			Hostname: sourceHost,
			Port:     sourcePort,
			Username: sourceUser,
			Password: sourcePass,
		}
		qb := qbittorrent.NewClient(qbtSettings)

		ctx := context.Background()

		if err := qb.Login(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		sourceData, err := qb.GetTorrents(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not get torrents %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Found: %d torrents on source\n", len(sourceData))

		// Choose the tag depending on the flags
		var tag string
		if overrideTag != "" {
			tag = overrideTag
		} else if tagDuplicates {
			tag = "duplicate"
		} else {
			tag = "" // No flag is set, so the tag is empty
		}

		// Start comparison
		for _, compareConfig := range config.Compare {
			compareHost := compareConfig.Host
			comparePort := compareConfig.Port
			compareUser := compareConfig.Login
			comparePass := compareConfig.Password

			qbtSettingsCompare := qbittorrent.Settings{
				Hostname: compareHost,
				Port:     comparePort,
				Username: compareUser,
				Password: comparePass,
			}
			qbCompare := qbittorrent.NewClient(qbtSettingsCompare)

			if err = qbCompare.Login(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: connection failed to compare: %v\n", err)
				os.Exit(1)
			}

			compareData, err := qbCompare.GetTorrents(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not get torrents from compare: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Found: %d torrents on compare\n", len(compareData))

			duplicateTorrents, err := compare(sourceData, compareData)
			if err != nil {
				os.Exit(1)
			}

			// Process duplicate torrents
			if tag != "" {
				batch := 20
				for i := 0; i < len(duplicateTorrents); i += batch {
					j := i + batch
					if j > len(duplicateTorrents) {
						j = len(duplicateTorrents)
					}

					qbCompare.SetTag(ctx, duplicateTorrents[i:j], tag)

					// sleep before next request
					time.Sleep(time.Second * 1)
				}
			}

			// --rm-duplicates

			// --save save to file
		}
	}

	return command
}

func compare(source, compare []qbittorrent.Torrent) ([]string, error) {
	sourceTorrents := make(map[string]qbittorrent.TorrentBasic, 0)

	for _, s := range source {
		sourceTorrents[s.Hash] = qbittorrent.TorrentBasic{
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
	duplicateTorrentsSlice := make([]qbittorrent.TorrentBasic, 0)

	var totalSize uint64

	for _, c := range compare {
		if _, ok := sourceTorrents[c.Hash]; ok {
			duplicateTorrentIDs = append(duplicateTorrentIDs, c.Hash)

			totalSize += uint64(c.Size)

			duplicateTorrentsSlice = append(duplicateTorrentsSlice, qbittorrent.TorrentBasic{
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
	fmt.Printf("Found: %d duplicate torrents\n", len(duplicateTorrentsSlice))

	fmt.Printf("Total reclaimable space: %v\n", humanize.Bytes(totalSize))

	return duplicateTorrentIDs, nil
}
