package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/autobrr/go-qbittorrent"
	"github.com/spf13/cobra"
)

// RunTorrentAdd cmd to add torrents
func RunTorrentAdd() *cobra.Command {
	var (
		dry           bool
		paused        bool
		skipHashCheck bool
		removeStalled bool
		savePath      string
		category      string
		tags          []string
		ignoreRules   bool
		uploadLimit   uint64
		downloadLimit uint64
		sleep         time.Duration
	)

	var command = &cobra.Command{
		Use:   "add",
		Short: "Add torrent(s)",
		Long:  `Add new torrent(s) to qBittorrent from file or magnet. Supports glob pattern for files like: ./files/*.torrent`,
		Example: `  qbt torrent add my-file.torrent --category test --tags tag1
  qbt torrent add ./files/*.torrent --paused --skip-hash-check
  qbt torrent add magnet:?xt=urn:btih:5dee65101db281ac9c46344cd6b175cdcad53426&dn=download`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a torrent file, glob or magnet as first argument")
			}

			return nil
		},
	}

	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().BoolVar(&paused, "paused", false, "Add torrent in paused state")
	command.Flags().BoolVar(&skipHashCheck, "skip-hash-check", false, "Skip hash check")
	command.Flags().BoolVar(&ignoreRules, "ignore-rules", false, "Ignore rules from config")
	command.Flags().BoolVar(&removeStalled, "remove-stalled", false, "Remove stalled torrents from re-announce")
	command.Flags().StringVar(&savePath, "save-path", "", "Add torrent to the specified path")
	command.Flags().StringVar(&category, "category", "", "Add torrent to the specified category")
	command.Flags().Uint64Var(&uploadLimit, "limit-ul", 0, "Set torrent upload speed limit. Unit in bytes/second")
	command.Flags().Uint64Var(&downloadLimit, "limit-dl", 0, "Set torrent download speed limit. Unit in bytes/second")
	command.Flags().DurationVar(&sleep, "sleep", 200*time.Millisecond, "Set the amount of time to wait between adding torrents in seconds")
	command.Flags().StringArrayVar(&tags, "tags", []string{}, "Add tags to torrent")

	command.Run = func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		// args
		// first arg is path to torrent file
		filePath := args[0]

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
			fmt.Fprintf(os.Stderr, "could not login to qbit: %q\n", err)
			os.Exit(1)
		}

		if config.Rules.Enabled && !ignoreRules {
			activeDownloads, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterDownloading})
			if err != nil {
				log.Fatalf("could not fetch torrents: %q\n", err)
			}

			if len(activeDownloads) >= config.Rules.MaxActiveDownloads {
				log.Printf("max active downloads of (%d) reached, skip adding\n", config.Rules.MaxActiveDownloads)
				return
			}
		}

		options := map[string]string{}
		if paused {
			options["paused"] = "true"
		}
		if skipHashCheck {
			options["skip_checking"] = "true"
		}
		if savePath != "" {
			//options["savepath"] = savePath
			options["autoTMM"] = "false"
		}
		if category != "" {
			options["category"] = category
		}
		if tags != nil {
			options["tags"] = strings.Join(tags, ",")
		}
		if uploadLimit > 0 {
			options["upLimit"] = strconv.FormatUint(uploadLimit, 10)
		}
		if downloadLimit > 0 {
			options["dlLimit"] = strconv.FormatUint(uploadLimit, 10)
		}

		if strings.HasPrefix(filePath, "magnet:") {
			if dry {
				log.Printf("dry-run: successfully added torrent from magnet %s!\n", filePath)

				return
			}

			if err := qb.AddTorrentFromUrlCtx(ctx, filePath, options); err != nil {
				log.Fatalf("adding torrent failed: %q\n", err)
			}

			hash := ""

			// some trackers are bugged or slow, so we need to re-announce the torrent until it works
			if config.Reannounce.Enabled && !paused {
				magnet, err := metainfo.ParseMagnetUri(filePath)
				if err != nil {
					fmt.Printf("could not parse magnet URI: %s\n", filePath)
				}

				hash := magnet.InfoHash.String()

				go func() {
					if err := checkTrackerStatus(ctx, qb, removeStalled, hash); err != nil {
						log.Fatalf("could not get tracker status for torrent: %q\n", err)
					}
				}()
			}

			log.Printf("successfully added torrent from magnet: %s %s\n", filePath, hash)
			return
		} else {
			var files []string
			_, err := os.Lstat(filePath)
			if err == nil {
				files = []string{filePath}
			} else {
				files, err = filepath.Glob(filePath)
				if err != nil {
					log.Fatalf("could not find files matching: %s err: %q\n", filePath, err)
				}
			}

			if len(files) == 0 {
				log.Printf("found 0 torrents matching %s\n", filePath)
				return
			}

			log.Printf("found (%d) torrent(s) to add\n", len(files))

			success := 0
			for _, file := range files {
				if dry {
					log.Printf("dry-run: torrent %s successfully added!\n", file)

					continue
				}

				// set savePath again
				options["savepath"] = savePath

				if err := qb.AddTorrentFromFileCtx(ctx, file, options); err != nil {
					log.Fatalf("adding torrent failed: %q\n", err)
				}

				// Get meta info from file to find out the hash for later use
				t, err := metainfo.LoadFromFile(file)
				if err != nil {
					fmt.Printf("could not open file: %s", file)
					continue
				}

				hash := t.HashInfoBytes().String()

				// some trackers are bugged or slow, so we need to re-announce the torrent until it works
				if config.Reannounce.Enabled && !paused {
					go func() {
						if err := checkTrackerStatus(ctx, qb, removeStalled, hash); err != nil {
							log.Printf("could not get tracker status for torrent: %s err: %q\n", hash, err)
						}
					}()
				}

				success++

				log.Printf("successfully added torrent: %s\n", hash)

				if len(files) > 1 {
					log.Printf("sleeping %v before adding next torrent...\n", sleep)

					time.Sleep(sleep)

					continue
				}

			}

			log.Printf("successfully added %d torrent(s)\n", success)
		}
	}

	return command
}

func checkTrackerStatus(ctx context.Context, qb *qbittorrent.Client, removeStalled bool, hash string) error {
	announceOK := false
	attempts := 0

	time.Sleep(time.Duration(config.Reannounce.Interval) * time.Millisecond)

	for attempts < config.Reannounce.Attempts {
		trackers, err := qb.GetTorrentTrackersCtx(ctx, hash)
		if err != nil {
			log.Fatalf("could not get trackers of torrent: %v %v", hash, err)
		}

		// check if status not working or something else
		_, working := findTrackerStatus(trackers, 2)
		if working {
			announceOK = true
			break
		}

		if err := qb.ReAnnounceTorrentsCtx(ctx, []string{hash}); err != nil {
			return err
		}

		time.Sleep(time.Duration(config.Reannounce.Interval) * time.Millisecond)
		attempts++
		continue
	}

	if !announceOK {
		if removeStalled {
			log.Println("Announce not ok, deleting torrent")

			if err := qb.DeleteTorrentsCtx(ctx, []string{hash}, false); err != nil {
				return err
			}
		}
	}

	return nil
}

// Check if status not working or something else
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)#get-torrent-trackers
//
//	0 Tracker is disabled (used for DHT, PeX, and LSD)
//	1 Tracker has not been contacted yet
//	2 Tracker has been contacted and is working
//	3 Tracker is updating
//	4 Tracker has been contacted, but it is not working (or doesn't send proper replies)
func findTrackerStatus(slice []qbittorrent.TorrentTracker, val int) (int, bool) {
	for i, item := range slice {
		if int(item.Status) == val {
			return i, true
		}
	}
	return -1, false
}
