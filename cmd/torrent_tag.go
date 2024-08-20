package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/autobrr/go-qbittorrent"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var trackerIssues = []string{
	"tracker is down",
	"maintenance",
	"down",
	"it may be down",
	"unreachable",
	"(unreachable)",
	"bad gateway",
	"tracker unavailable",
}

var trackerMessages = []string{
	"unregistered",
	"not registered",
	"torrent not found",
	"torrent is not found",
	"unknown torrent",
	"retitled",
	"truncated",
	"torrent is not authorized for use on this tracker",
	"infohash not found",
	"not found",
	"not exist",
	"unknown",
	"unknown torrent",
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
	"trump",
	"trumped",
	"nuked",
	"i'm sorry dave, i can't do that",
	"problem with description",
	"problem with file",
	"problem with pack",
	"specifically banned",
	"other",
	"torrent has been deleted",
}

var ourTags = []string{
	"Not Working",
	"Unregistered",
	"Tracker Down",
	//"Duplicate",
}

// RunTorrentTag cmd for torrent tag operations
func RunTorrentTag() *cobra.Command {
	var command = &cobra.Command{
		Use:   "tag",
		Short: "Torrent tag subcommand",
		Long:  `Do various torrent tag operations`,
	}

	command.AddCommand(RunTorrentTagNotWorking())

	return command
}

type tagData struct {
	Hashes    []string
	TotalSize uint64
}

// RunTorrentTagNotWorking tag torrents
func RunTorrentTagNotWorking() *cobra.Command {
	var (
		dryRun          bool
		tagUnregistered bool
		tagNotWorking   bool
		tagTrackerDown  bool
		size            bool
	)

	var command = &cobra.Command{
		Use:     "run",
		Short:   "tag torrents",
		Long:    `tag torrents`,
		Example: `  qbt torrent tag run --unregistered --not-working --tracker-down`,
	}

	command.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run, do not tag torrents")
	command.Flags().BoolVar(&tagUnregistered, "unregistered", false, "tag unregistered")
	command.Flags().BoolVar(&tagNotWorking, "not-working", false, "tag not working torrents")
	command.Flags().BoolVar(&tagTrackerDown, "tracker-down", false, "tag tracker issues")
	command.Flags().BoolVar(&size, "size", false, "collect size per tag")

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

		ctx := context.Background()

		if err := qb.LoginCtx(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		sourceData, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not get torrents %v\n", err)
			os.Exit(1)
		}

		var totalSize uint64

		unregisteredTorrents := tagData{
			Hashes:    []string{},
			TotalSize: 0,
		}

		notworkingTorrents := tagData{
			Hashes:    []string{},
			TotalSize: 0,
		}

		removeTaggedTorrents := tagData{
			Hashes:    []string{},
			TotalSize: 0,
		}

		for _, torrent := range sourceData {
			// check for our default tags to first clear
			if torrent.Tags != "" {
				torrentTags := strings.Split(torrent.Tags, ",")

				for _, tag := range torrentTags {
					for _, defaultTag := range ourTags {
						if strings.Contains(tag, defaultTag) {
							removeTaggedTorrents.Hashes = append(removeTaggedTorrents.Hashes, torrent.Hash)
						}
					}
				}
			}

			var isUnregistered bool
			var isNotWorking bool
			//var isTrackerDown bool

			if tagUnregistered && torrent.Tracker == "" {
				isUnregistered = true
			}

			trackers, err := qb.GetTorrentTrackersCtx(ctx, torrent.Hash)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not get trackers for torrent %v: %v\n", torrent.Hash, err)
				continue
			}

			for _, tracker := range trackers {
				lowerTrackerMessage := strings.ToLower(tracker.Message)

				if tagUnregistered && !isUnregistered {
					for _, msg := range trackerMessages {
						if strings.Contains(lowerTrackerMessage, msg) {
							isUnregistered = true
							break
						}
					}
				}

				if tagNotWorking && !isNotWorking {
					for _, msg := range trackerIssues {
						if strings.Contains(lowerTrackerMessage, msg) {
							isNotWorking = true
							break
						}
					}
				}

				//if isUnregistered {
				//	break
				//}
			}

			if isUnregistered {
				unregisteredTorrents.TotalSize += uint64(torrent.Size)
				unregisteredTorrents.Hashes = append(unregisteredTorrents.Hashes, torrent.Hash)
			}

			if isNotWorking {
				notworkingTorrents.TotalSize += uint64(torrent.Size)
				notworkingTorrents.Hashes = append(notworkingTorrents.Hashes, torrent.Hash)
			}

			totalSize += uint64(torrent.Size)
		}

		fmt.Printf("total torrents (%d) with a total size of: %s\n", len(sourceData), humanize.Bytes(totalSize))

		// --unregistered add tag unregistered
		if tagUnregistered {
			fmt.Printf("reclaimable space (%s) from (%d) unregistered torrents\n", humanize.Bytes(unregisteredTorrents.TotalSize), len(unregisteredTorrents.Hashes))

			if dryRun {
				fmt.Printf("dry-run: tagging (%d) unregistered torrents\n", len(unregisteredTorrents.Hashes))
				fmt.Printf("dry-run: successfully tagged (%d) unregistered torrents\n", len(unregisteredTorrents.Hashes))
			} else {
				fmt.Printf("tagging (%d) unregistered torrents\n", len(unregisteredTorrents.Hashes))

				err := batchRequests(unregisteredTorrents.Hashes, func(start, end int) error {
					return qb.AddTagsCtx(ctx, unregisteredTorrents.Hashes[start:end], "Unregistered")
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not pause torrents: %v\n", err)
					os.Exit(1)
					return
				}

				fmt.Printf("successfully tagged (%d) unregistered torrents\n", len(unregisteredTorrents.Hashes))
			}
		}

		if tagNotWorking {
			fmt.Printf("reclaimable space (%s) from (%d) not working torrents\n", humanize.Bytes(notworkingTorrents.TotalSize), len(notworkingTorrents.Hashes))

			if dryRun {
				fmt.Printf("dry-run: tagging (%d) not working torrents\n", len(notworkingTorrents.Hashes))
				fmt.Printf("dry-run: successfully tagged (%d) not working torrents\n", len(notworkingTorrents.Hashes))
			} else {
				fmt.Printf("tagging (%d) not working torrents\n", len(notworkingTorrents.Hashes))

				err := batchRequests(notworkingTorrents.Hashes, func(start, end int) error {
					return qb.AddTagsCtx(ctx, notworkingTorrents.Hashes[start:end], "Not Working")
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not tag torrents: %v\n", err)
					os.Exit(1)
					return
				}

				fmt.Printf("successfully tagged (%d) not working torrents\n", len(notworkingTorrents.Hashes))
			}
		}
	}

	return command
}
