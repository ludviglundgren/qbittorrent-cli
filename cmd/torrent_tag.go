package cmd

import (
	"context"
	"fmt"
	"log"
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

//var DefaultTags = []string{
//	"Not Working",
//	"Unregistered",
//	"Tracker Down",
//	//"Duplicate",
//}

type DefaultTag string

const (
	DefaultTagUnregistered DefaultTag = "Unregistered"
	DefaultTagNotWorking   DefaultTag = "Not Working"
)

func (t DefaultTag) String() string {
	switch t {
	case DefaultTagUnregistered:
		return "Unregistered"
	case DefaultTagNotWorking:
		return "Not Working"
	}

	return ""
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
	Hashes     []string
	HashTagMap map[string][]string
	TotalSize  uint64
}

// RunTorrentTagNotWorking tag torrents
func RunTorrentTagNotWorking() *cobra.Command {
	var (
		dryRun          bool
		tagUnregistered bool
		tagNotWorking   bool
		//size            bool
	)

	var command = &cobra.Command{
		Use:     "issues",
		Short:   "tag torrents with issues",
		Long:    `Tag torrents that may have broken trackers or be unregistered`,
		Example: `  qbt torrent tag issues --unregistered --not-working`,
	}

	command.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run, do not tag torrents")
	command.Flags().BoolVar(&tagUnregistered, "unregistered", false, "tag unregistered")
	command.Flags().BoolVar(&tagNotWorking, "not-working", false, "tag not working torrents")
	//command.Flags().BoolVar(&size, "size", false, "collect size per tag")

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

		torrents, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not get torrents %v\n", err)
			os.Exit(1)
		}

		var totalSize uint64

		unregisteredTorrents := &tagData{
			Hashes:    []string{},
			TotalSize: 0,
		}

		notWorkingTorrents := &tagData{
			Hashes:    []string{},
			TotalSize: 0,
		}

		removeTaggedTorrents := &tagData{
			Hashes: []string{},
			HashTagMap: map[string][]string{
				DefaultTagUnregistered.String(): {},
				DefaultTagNotWorking.String():   {},
			},
			TotalSize: 0,
		}

		// process each torrent and check if tags should be added or removed
		for _, torrent := range torrents {
			totalSize += uint64(torrent.Size)

			trackers, err := qb.GetTorrentTrackersCtx(ctx, torrent.Hash)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not get trackers for torrent %v: %v\n", torrent.Hash, err)
				continue
			}

			processTorrentTags(torrent, trackers, removeTaggedTorrents, unregisteredTorrents, notWorkingTorrents, tagUnregistered, tagUnregistered)
		}

		log.Printf("total torrents (%d) with a total size of: %s\n", len(torrents), humanize.Bytes(totalSize))

		// remove tags from torrents that should not have certain tags
		if dryRun {
			log.Printf("dry-run: clearing defualt tags from torrents\n")
		} else {
			log.Printf("clearing defualt tags from torrents\n")

			for tag, hashes := range removeTaggedTorrents.HashTagMap {
				err := batchRequests(hashes, func(start, end int) error {
					return qb.RemoveTagsCtx(ctx, hashes[start:end], tag)
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "could remove tag %s from torrents %+v: %v\n", tag, hashes, err)
					os.Exit(1)
					return
				}

				log.Printf("successfully cleared tags from (%d) torrents\n", len(hashes))
			}
		}

		// --unregistered add tag unregistered
		if tagUnregistered {
			countUnregisteredTorrents := len(unregisteredTorrents.Hashes)

			log.Printf("reclaimable space (%s) from (%d) unregistered torrents\n", humanize.Bytes(unregisteredTorrents.TotalSize), countUnregisteredTorrents)

			if dryRun {
				log.Printf("dry-run: tagging (%d) unregistered torrents\n", countUnregisteredTorrents)
				log.Printf("dry-run: successfully tagged (%d) unregistered torrents\n", countUnregisteredTorrents)
			} else {
				log.Printf("tagging (%d) unregistered torrents\n", countUnregisteredTorrents)

				err := batchRequests(unregisteredTorrents.Hashes, func(start, end int) error {
					return qb.AddTagsCtx(ctx, unregisteredTorrents.Hashes[start:end], DefaultTagUnregistered.String())
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not tag torrents: %v\n", err)
					os.Exit(1)
					return
				}

				log.Printf("successfully tagged (%d) unregistered torrents\n", countUnregisteredTorrents)
			}
		}

		if tagNotWorking {
			countNotWorkingTorrents := len(notWorkingTorrents.Hashes)

			log.Printf("reclaimable space (%s) from (%d) not working torrents\n", humanize.Bytes(notWorkingTorrents.TotalSize), countNotWorkingTorrents)

			if dryRun {
				log.Printf("dry-run: tagging (%d) not working torrents\n", countNotWorkingTorrents)
				log.Printf("dry-run: successfully tagged (%d) not working torrents\n", countNotWorkingTorrents)
			} else {
				log.Printf("tagging (%d) not working torrents\n", len(notWorkingTorrents.Hashes))

				err := batchRequests(notWorkingTorrents.Hashes, func(start, end int) error {
					return qb.AddTagsCtx(ctx, notWorkingTorrents.Hashes[start:end], DefaultTagNotWorking.String())
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not tag torrents: %v\n", err)
					os.Exit(1)
					return
				}

				log.Printf("successfully tagged (%d) not working torrents\n", countNotWorkingTorrents)
			}
		}
	}

	return command
}

func processTorrentTags(torrent qbittorrent.Torrent, trackers []qbittorrent.TorrentTracker, removeTaggedTorrents *tagData, unregisteredTorrents *tagData, notWorkingTorrents *tagData, tagUnregistered bool, tagNotWorking bool) bool {
	var isUnregistered bool
	var isNotWorking bool
	//var isTrackerDown bool

	// check for our default tags to first clear
	if torrent.Tags != "" {
		torrentTags := strings.Split(torrent.Tags, ", ")

		for _, tag := range torrentTags {
			if strings.Contains(tag, DefaultTagUnregistered.String()) {
				isUnregistered = true
			}

			if strings.Contains(tag, DefaultTagNotWorking.String()) {
				isNotWorking = true
			}
		}
	}

	foundTrackerUnregistered := false
	foundTrackerNotWorking := false

	for _, tracker := range trackers {
		if tracker.Status == qbittorrent.TrackerStatusDisabled {
			continue
		}

		//if tracker.Url == "** [DHT] **" || tracker.Url == "** [PeX] **" || tracker.Url == "** [LSD] **" {
		//	continue
		//}
		lowerTrackerMessage := strings.ToLower(tracker.Message)

		if tagUnregistered {
			for _, msg := range trackerMessages {
				if strings.Contains(lowerTrackerMessage, msg) {
					//isUnregistered = true
					foundTrackerUnregistered = true
					break
				}
			}
		}

		if tagNotWorking {
			for _, msg := range trackerIssues {
				if strings.Contains(lowerTrackerMessage, msg) {
					//isNotWorking = true
					foundTrackerNotWorking = true
					break
				}
			}
		}
	}

	// if initial status was isNotWorking and the tracker message changed we need to clear the tag for the hash
	if isNotWorking && !foundTrackerNotWorking {
		removeTaggedTorrents.HashTagMap[DefaultTagNotWorking.String()] = append(removeTaggedTorrents.HashTagMap[DefaultTagUnregistered.String()], torrent.Hash)
	}

	if isUnregistered && !foundTrackerUnregistered {
		removeTaggedTorrents.HashTagMap[DefaultTagUnregistered.String()] = append(unregisteredTorrents.HashTagMap[DefaultTagUnregistered.String()], torrent.Hash)
	}

	// found new torrent to tag
	if !isUnregistered && foundTrackerUnregistered {
		unregisteredTorrents.TotalSize += uint64(torrent.Size)
		unregisteredTorrents.Hashes = append(unregisteredTorrents.Hashes, torrent.Hash)
	}

	// found new torrent to tag
	if !isNotWorking && foundTrackerNotWorking {
		notWorkingTorrents.TotalSize += uint64(torrent.Size)
		notWorkingTorrents.Hashes = append(notWorkingTorrents.Hashes, torrent.Hash)
	}

	return true
}
