package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"
)

// RunTag tag torrents
func RunTag() *cobra.Command {
	var (
		tagUnregistered bool
		dryRun          bool
	)

	defaultTags := []string{
		"Not Working",
		"added:",
		"Unregistered",
		//"Tracker Down",
		"t:",
		"Duplicates",
		"activity:",
		"Not Linked",
	}

	unregisteredMatches := []string{
		"unregistered",
		"not registered",
		"not found",
		"not exist",
		"unknown",
		"uploaded",
		"upgraded",
		"season pack",
		"packs are available",
		"pack is available",
		"internal available",
		"season pack out",
		"dead",
		"dupe",
		"complete season uploaded",
		"problem with",
		"specifically banned",
		"trumped",
		"i'm sorry dave, i can't do that", // weird stuff from racingforme
	}

	var command = &cobra.Command{
		Use:   "tag",
		Short: "tag torrents",
		Long:  `tag torrents`,
	}

	command.Flags().BoolVar(&tagUnregistered, "unregistered", false, "tag unregistered")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run, do not tag torrents")

	command.Run = func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		qbtSettings := qbittorrent.Settings{
			Addr:      config.Qbit.Addr,
			Hostname:  config.Qbit.Host,
			Port:      config.Qbit.Port,
			Username:  config.Qbit.Login,
			Password:  config.Qbit.Password,
			BasicUser: config.Qbit.BasicUser,
			BasicPass: config.Qbit.BasicPass,
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

		unregisteredTorrentIDs := make([]string, 0)

		var totalSize uint64
		var unregisteredSize uint64
		for _, t := range sourceData {

			// Skip if tracker has not been contacted yet
			if (t.TrackerStatus) == 0 { // 0 = Tracker has not been contacted yet
				continue // uo for debate if needed or not
			}

			// Check for unregistered
			if tagUnregistered {
				if t.Tracker == "" {
					unregisteredTorrentIDs = append(unregisteredTorrentIDs, t.Hash)
					unregisteredSize += uint64(t.Size)
				}
			}

			// Check for each tag in defaultTags
			for _, tag := range defaultTags {
				if strings.Contains(t.TrackerMessage, tag) {
					unregisteredTorrentIDs = append(unregisteredTorrentIDs, t.Hash)
					unregisteredSize += uint64(t.Size)
				}
			}

			// Check for each match in unregisteredMatches
			for _, match := range unregisteredMatches {
				if strings.Contains(t.TrackerMessage, match) {
					unregisteredTorrentIDs = append(unregisteredTorrentIDs, t.Hash)
					unregisteredSize += uint64(t.Size)
				}
			}

			totalSize += uint64(t.Size)
		}

		fmt.Printf("total torrents (%d) with a total size of: %s\n", len(sourceData), humanize.Bytes(totalSize))

		// --unregistered add tag unregistered
		if tagUnregistered {
			fmt.Printf("reclaimable space (%s) from (%d) unregistered torrents\n", humanize.Bytes(unregisteredSize), len(unregisteredTorrentIDs))

			if dryRun {
				fmt.Printf("dry-run: tagging (%d) unregistered torrents\n", len(unregisteredTorrentIDs))
				fmt.Printf("dry-run: successfully tagged (%d) unregistered torrents\n", len(unregisteredTorrentIDs))
			} else {
				fmt.Printf("tagging (%d) unregistered torrents\n", len(unregisteredTorrentIDs))

				// Split the slice into batches of 20 items.
				batch := 20
				for i := 0; i < len(unregisteredTorrentIDs); i += batch {
					j := i + batch
					if j > len(unregisteredTorrentIDs) {
						j = len(unregisteredTorrentIDs)
					}

					fmt.Printf("batch tagging (%d) unregistered torrents...\n", j)

					if err := qb.SetTag(ctx, unregisteredTorrentIDs[i:j], "unregistered"); err != nil {
						fmt.Printf("could not set tag, err: %q", err)
					}

					fmt.Printf("sleep 1 second before next...\n")

					// sleep before next request
					time.Sleep(time.Second * 1)
				}

				fmt.Printf("successfully tagged (%d) unregistered torrents\n", len(unregisteredTorrentIDs))
			}

		}
	}

	return command
}
