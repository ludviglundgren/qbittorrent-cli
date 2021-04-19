package cmd

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/spf13/cobra"
)

// RunAdd cmd to add torrents
func RunAdd() *cobra.Command {
	var (
		magnet        bool
		dry           bool
		paused        bool
		skipHashCheck bool
		savePath      string
		category      string
		tags          []string
		ignoreRules   bool
		uploadLimit   uint64
		downloadLimit uint64
	)

	var command = &cobra.Command{
		Use:   "add",
		Short: "Add torrent",
		Long:  `Add new torrent to qBittorrent from file`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a torrent file as first argument")
			}

			return nil
		},
	}
	command.Flags().BoolVar(&magnet, "magnet", false, "Add magnet link instead of torrent file")
	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().BoolVar(&paused, "paused", false, "Add torrent in paused state")
	command.Flags().BoolVar(&skipHashCheck, "skip-hash-check", false, "Skip hash check")
	command.Flags().BoolVar(&ignoreRules, "ignore-rules", false, "Ignore rules from config")
	command.Flags().StringVar(&savePath, "save-path", "", "Add torrent to the specified path")
	command.Flags().StringVar(&category, "category", "", "Add torrent to the specified category")
	command.Flags().Uint64Var(&uploadLimit, "limit-ul", 0, "Set torrent upload speed limit. Unit in bytes/second")
	command.Flags().Uint64Var(&downloadLimit, "limit-dl", 0, "Set torrent download speed limit. Unit in bytes/second")
	command.Flags().StringArrayVar(&tags, "tags", []string{}, "Add tags to torrent")

	command.Run = func(cmd *cobra.Command, args []string) {
		// args
		// first arg is path to torrent file
		filePath := args[0]

		if !dry {
			qbtSettings := qbittorrent.Settings{
				Hostname: config.Qbit.Host,
				Port:     config.Qbit.Port,
				Username: config.Qbit.Login,
				Password: config.Qbit.Password,
			}
			qb := qbittorrent.NewClient(qbtSettings)
			err := qb.Login()
			if err != nil {
				log.Fatalf("connection failed %v", err)
			}

			if config.Rules.Enabled && !ignoreRules {
				activeDownloads, err := qb.GetTorrentsFilter(qbittorrent.TorrentFilterDownloading)
				if err != nil {
					log.Fatalf("could not fetch torrents: %v", err)
				}

				if len(activeDownloads) >= config.Rules.MaxActiveDownloads {
					log.Fatalf("max active downloads reached, skip adding: %v", err)
				}
			}

			options := map[string]string{}
			if paused != false {
				options["paused"] = "true"
			}
			if skipHashCheck != false {
				options["skip_checking"] = "true"
			}
			if savePath != "" {
				options["savepath"] = savePath
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

			var res string
			if magnet {
				res, err = qb.AddTorrentFromMagnet(filePath, options)
			} else {
				res, err = qb.AddTorrentFromFile(filePath, options)
			}
			if err != nil {
				log.Fatalf("adding torrent failed: %v", err)
			}

			// some trackers are bugged or slow so we need to re-announce the torrent until it works
			if config.Reannounce.Enabled && !paused {
				err = checkTrackerStatus(*qb, res)
				if err != nil {
					log.Fatalf("could not get tracker status for torrent: %v", err)
				}
			}

			log.Printf("torrent successfully added: %v", res)
		} else {
			log.Println("dry-run: torrent successfully added!")
		}
	}

	return command
}

func checkTrackerStatus(qb qbittorrent.Client, hash string) error {
	announceOK := false
	attempts := 0

	time.Sleep(time.Duration(config.Reannounce.Interval) * time.Millisecond)

	for attempts < config.Reannounce.Attempts {
		trackers, err := qb.GetTorrentTrackers(hash)
		if err != nil {
			log.Fatalf("could not get trackers of torrent: %v %v", hash, err)
		}

		// check if status not working or something else
		_, working := findTrackerStatus(trackers, 2)

		if !working {
			err = qb.ReAnnounceTorrents([]string{hash})
			if err != nil {
				return err
			}

			time.Sleep(time.Duration(config.Reannounce.Interval) * time.Millisecond)
			attempts++
			continue
		} else {
			announceOK = true
			break
		}
	}

	if !announceOK {
		log.Println("Announce not ok, deleting torrent")
		err := qb.DeleteTorrents([]string{hash}, false)
		if err != nil {
			return err
		}
	}

	return nil
}

// Check if status not working or something else
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)#get-torrent-trackers
//  0 Tracker is disabled (used for DHT, PeX, and LSD)
//  1 Tracker has not been contacted yet
//  2 Tracker has been contacted and is working
//  3 Tracker is updating
//  4 Tracker has been contacted, but it is not working (or doesn't send proper replies)
func findTrackerStatus(slice []qbittorrent.TorrentTracker, val int) (int, bool) {
	for i, item := range slice {
		if item.Status == val {
			return i, true
		}
	}
	return -1, false
}
