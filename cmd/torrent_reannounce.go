package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/autobrr/go-qbittorrent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	words = []string{"unregistered", "not registered", "not found", "not exist"}
)

// RunTorrentReannounce cmd to reannounce torrents
func RunTorrentReannounce() *cobra.Command {
	var (
		dry      bool
		hash     string
		category string
		tag      string
		attempts int
		interval int
		maxAge   int64
	)

	var command = &cobra.Command{
		Use:   "reannounce",
		Short: "Reannounce torrent(s)",
		Long:  `Reannounce torrents with non-OK tracker status.`,
	}

	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().StringVar(&hash, "hash", "", "Reannounce torrent with hash")
	command.Flags().StringVar(&category, "category", "", "Reannounce torrents with category")
	command.Flags().StringVar(&tag, "tag", "", "Reannounce torrents with tag")
	command.Flags().IntVar(&attempts, "attempts", 50, "Reannounce torrents X times")
	command.Flags().IntVar(&interval, "interval", 7000, "Reannounce torrents X times with interval Y. In MS")
	command.Flags().Int64Var(&maxAge, "max-age", 120, "Reannounce torrents up to X seconds old")

	command.RunE = func(cmd *cobra.Command, args []string) error {
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

		req := qbittorrent.TorrentFilterOptions{
			//Filter:   qbittorrent.TorrentFilterDownloading,
			Category: category,
			Tag:      tag,
			Hashes:   []string{hash},
		}

		activeDownloads, err := qb.GetTorrentsCtx(ctx, req)
		if err != nil {
			return errors.Wrap(err, "could not fetch torrents")
		}

		if hash != "" && len(activeDownloads) != 1 {
			return errors.Errorf("torrent not found: %s", hash)
		}

		if dry {
			log.Println("dry-run: torrents successfully re-announced!")
		} else {
			var wg sync.WaitGroup

			reannounceOptions := qbittorrent.ReannounceOptions{
				Interval:        interval,
				MaxAttempts:     attempts,
				DeleteOnFailure: false,
			}

			for _, torrent := range activeDownloads {
				if shouldReannounce(torrent, maxAge) {
					wg.Add(1)
					go func(t qbittorrent.Torrent) {
						defer wg.Done()
						if err := reannounceWithRetry(ctx, qb, t, &reannounceOptions); err != nil {
							log.Printf("Failed to reannounce torrent: %s %s err: %q\n", t.Hash, t.Name, err)
						} else {
							log.Printf("successfully re-announced torrent: %s %s\n", t.Hash, t.Name)
						}
					}(torrent)
				}
			}

			wg.Wait()

			log.Println("torrents successfully re-announced")
		}

		return nil
	}

	return command
}

func shouldReannounce(torrent qbittorrent.Torrent, maxAge int64) bool {
	if torrent.TimeActive > maxAge {
		return false
	}

	if torrent.NumSeeds > 0 || torrent.NumLeechs > 0 {
		return false
	}

	return !isTrackerStatusOK(torrent.Trackers)
}

func isTrackerStatusOK(trackers []qbittorrent.TorrentTracker) bool {
	for _, tracker := range trackers {
		if tracker.Status == qbittorrent.TrackerStatusDisabled {
			continue
		}

		// check for certain messages before the tracker status to catch ok status with unreg msg
		if isUnregistered(tracker.Message) {
			return false
		}

		if tracker.Status == qbittorrent.TrackerStatusOK {
			return true
		}
	}

	return false
}

func isUnregistered(msg string) bool {
	msg = strings.ToLower(msg)

	for _, v := range words {
		if strings.Contains(msg, v) {
			return true
		}
	}

	return false
}

func reannounceWithRetry(ctx context.Context, client *qbittorrent.Client, torrent qbittorrent.Torrent, opts *qbittorrent.ReannounceOptions) error {
	if err := client.ReannounceTorrentWithRetry(ctx, torrent.Hash, opts); err != nil {
		if errors.Is(err, qbittorrent.ErrReannounceTookTooLong) {
			return fmt.Errorf("reannouncement timeout for torrent %s", torrent.Hash)
		}
		return fmt.Errorf("reannouncement failed for torrent %s: %w", torrent.Hash, err)
	}

	return nil
}
