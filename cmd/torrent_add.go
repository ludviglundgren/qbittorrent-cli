package cmd

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/autobrr/go-qbittorrent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type QBitStopCondition string

const (
	QBitStopConditionNone             QBitStopCondition = "None"
	QbitStopConditionMetadataReceived QBitStopCondition = "MetadataReceived"
	QbitStopConditionFilesChecked     QBitStopCondition = "FilesChecked"
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
		stopCondition string
		sleep         time.Duration
		recheck       bool
	)

	command := &cobra.Command{
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
	command.Flags().StringVar(&stopCondition, "stop-condition", "", "Add torrent with the specified stop condition. Possible values: None, MetadataReceived, FilesChecked. Example: --stop-condition MetadataReceived")
	command.Flags().Uint64Var(&uploadLimit, "limit-ul", 0, "Set torrent upload speed limit. Unit in bytes/second")
	command.Flags().Uint64Var(&downloadLimit, "limit-dl", 0, "Set torrent download speed limit. Unit in bytes/second")
	command.Flags().DurationVar(&sleep, "sleep", 200*time.Millisecond, "Set the amount of time to wait between adding torrents in seconds")
	command.Flags().StringArrayVar(&tags, "tags", []string{}, "Add tags to torrent")
	command.Flags().BoolVar(&recheck, "recheck", false, "Force recheck after adding (useful when using --paused)")

	command.RunE = func(cmd *cobra.Command, args []string) error {
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
			return errors.Wrap(err, "could not login to qbit")
		}

		if config.Rules.Enabled && !ignoreRules {
			activeDownloads, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterDownloading})
			if err != nil {
				return errors.Wrap(err, "could not fetch torrents")
			}

			if len(activeDownloads) >= config.Rules.MaxActiveDownloads {
				log.Printf("max active downloads of (%d) reached, skip adding\n", config.Rules.MaxActiveDownloads)
				return nil
			}
		}

		options := map[string]string{}
		if paused {
			options["paused"] = "true"
			options["stopped"] = "true"
		}
		if skipHashCheck {
			options["skip_checking"] = "true"
		}
		if savePath != "" {
			// options["savepath"] = savePath
			options["autoTMM"] = "false"
		}
		if category != "" {
			options["category"] = category
		}
		if tags != nil {
			options["tags"] = strings.Join(tags, ",")
		}
		if stopCondition != "" && (stopCondition == string(QBitStopConditionNone) || stopCondition == string(QbitStopConditionMetadataReceived) || stopCondition == string(QbitStopConditionFilesChecked)) {
			options["stop_condition"] = stopCondition
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

				return nil
			}

			if err := qb.AddTorrentFromUrlCtx(ctx, filePath, options); err != nil {
				return errors.Wrapf(err, "adding torrent %s failed", filePath)
			}

			hash := ""

			// some trackers are bugged or slow, so we need to re-announce the torrent until it works
			if config.Reannounce.Enabled && !paused {
				magnet, err := metainfo.ParseMagnetUri(filePath)
				if err != nil {
					return errors.Wrapf(err, "could not parse magnet URI: %s", filePath)
				}

				hash := magnet.InfoHash.String()

				wg := sync.WaitGroup{}

				wg.Add(1)

				go func() {
					defer wg.Done()
					if err := checkTrackerStatus(ctx, qb, removeStalled, hash); err != nil {
						log.Fatalf("could not get tracker status for torrent: %q\n", err)
					}
				}()

				wg.Wait()
			}

			if paused && recheck {
				magnet, err := metainfo.ParseMagnetUri(filePath)
				if err == nil {
					hash = magnet.InfoHash.String()
					if err := qb.RecheckCtx(ctx, []string{hash}); err != nil {
						log.Printf("could not recheck torrent: %s err: %q\n", hash, err)
					} else {
						log.Printf("rechecked torrent: %s\n", hash)
					}
				}
			}

			log.Printf("successfully added torrent from magnet: %s %s\n", filePath, hash)
			return nil
		} else {
			var files []string
			var err error

			var tempFile *os.File

			defer func() {
				if tempFile != nil {
					os.Remove(tempFile.Name())
					tempFile.Close()
				}
			}()

			if strings.HasPrefix(filePath, "https://") || strings.HasPrefix(filePath, "http://") {
				tempFile, err = os.CreateTemp("", "qbt-torrent-dl")
				if err != nil {
					return errors.Wrap(err, "could not create tmp file")
				}

				response, err := http.Get(filePath)
				if err != nil {
					return errors.Wrapf(err, "could not download file: %s", filePath)
				}

				defer response.Body.Close()

				if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusNoContent {
					return errors.Errorf("unexpected status: %d", response.StatusCode)
				}

				_, err = io.Copy(tempFile, response.Body)
				if err != nil {
					return errors.Wrap(err, "could not write download locally")
				}

				files = []string{tempFile.Name()}
			} else if IsGlobPattern(filePath) {
				files, err = filepath.Glob(filePath)
				if err != nil {
					return errors.Wrapf(err, "could not find files matching: %s", filePath)
				}
			} else {
				_, err := os.Lstat(filePath)
				if err != nil {
					return errors.Wrapf(err, "could not find file: %s", filePath)
				}

				files = []string{filePath}
			}

			if len(files) == 0 {
				log.Printf("found 0 torrents matching %s\n", filePath)
				return nil
			}

			log.Printf("found (%d) torrent(s) to add\n", len(files))

			wg := sync.WaitGroup{}

			success := 0
			for _, file := range files {
				if dry {
					log.Printf("dry-run: torrent %s successfully added!\n", file)

					continue
				}

				// set savePath again
				options["savepath"] = savePath

				if err := qb.AddTorrentFromFileCtx(ctx, file, options); err != nil {
					return errors.Wrapf(err, "could not add torrent: %s", file)
				}

				// Get meta info from file to find out the hash for later use
				t, err := metainfo.LoadFromFile(file)
				if err != nil {
					log.Printf("could not open file: %s", file)
					continue
				}

				hash := t.HashInfoBytes().String()

				if paused && recheck {
					if err := qb.RecheckCtx(ctx, []string{hash}); err != nil {
						log.Printf("could not recheck torrent: %s err: %q\n", hash, err)
					} else {
						log.Printf("rechecked torrent: %s\n", hash)
					}
				}

				// some trackers are bugged or slow, so we need to re-announce the torrent until it works
				if config.Reannounce.Enabled && !paused {
					wg.Add(1)

					go func() {
						defer wg.Done()
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

			wg.Wait()

			log.Printf("successfully added %d torrent(s)\n", success)
		}

		return nil
	}

	return command
}

// IsGlobPattern reports whether path contains any of the magic characters
// recognized by Match.
func IsGlobPattern(path string) bool {
	magicChars := `*?[`
	if runtime.GOOS != "windows" {
		magicChars = `*?[\`
	}
	return strings.ContainsAny(path, magicChars)
}

func checkTrackerStatus(ctx context.Context, qb *qbittorrent.Client, removeStalled bool, hash string) error {
	announceOK := false
	attempts := 0

	time.Sleep(time.Duration(config.Reannounce.Interval) * time.Millisecond)

	log.Printf("starting reannounce for %s\n", hash)

	for attempts < config.Reannounce.Attempts {
		log.Printf("[%d/%d] reannounce attempt for %s\n", attempts, config.Reannounce.Attempts, hash)

		trackers, err := qb.GetTorrentTrackersCtx(ctx, hash)
		if err != nil {
			log.Fatalf("could not get trackers of torrent: %v %v", hash, err)
		}

		// check if status not working or something else
		_, working := findTrackerStatus(trackers, 2)
		if working {
			log.Printf("[%d/%d] reannounce found working tracker for %s\n", attempts, config.Reannounce.Attempts, hash)
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
