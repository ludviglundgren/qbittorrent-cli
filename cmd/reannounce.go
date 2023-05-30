package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/spf13/cobra"
)

// RunReannounce cmd to reannounce torrents
func RunReannounce() *cobra.Command {
	var (
		dry      bool
		hash     string
		category string
		tag      string
		attempts int
		interval int
	)

	var command = &cobra.Command{
		Use:   "reannounce",
		Short: "Reannounce torrent(s)",
		Long:  `Reannounce torrents if needed.`,
	}

	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().StringVar(&hash, "hash", "", "Reannounce torrent with hash")
	command.Flags().StringVar(&category, "category", "", "Reannounce torrents with category")
	command.Flags().StringVar(&tag, "tag", "", "Reannounce torrents with tag")
	command.Flags().IntVar(&attempts, "attempts", 50, "Reannounce torrents X times")
	command.Flags().IntVar(&interval, "interval", 7000, "Reannounce torrents X times with interval Y. In MS")

	command.Run = func(cmd *cobra.Command, args []string) {
		config.InitConfig()

		if !dry {
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

			req := &qbittorrent.GetTorrentsRequest{
				Filter:   string(qbittorrent.TorrentFilterDownloading),
				Category: category,
				Tag:      tag,
				Hashes:   hash,
			}

			activeDownloads, err := qb.GetTorrentsWithFilters(ctx, req)
			if err != nil {
				log.Fatalf("could not fetch torrents: err: %q", err)
			}

			for _, torrent := range activeDownloads {
				if torrent.Progress == 0 && torrent.TimeActive < 120 {
					go func(torrent qbittorrent.Torrent) {
						log.Printf("torrent %s %s not working, active for %ds, re-announcing...\n", torrent.Hash, torrent.Name, torrent.TimeActive)

						// some trackers are bugged or slow, so we need to re-announce the torrent until it works
						if err = reannounceTorrent(ctx, qb, interval, attempts, torrent.Hash); err != nil {
							log.Printf("could not re-announce torrent: %s %s err: %q\n", torrent.Hash, torrent.Name, err)
						}

						log.Printf("successfully re-announced torrent: %s %s err: %q\n", torrent.Hash, torrent.Name, err)

					}(torrent)
				}
			}

			log.Println("torrents successfully re-announced")
		} else {
			log.Println("dry-run: torrents successfully re-announced!")
		}
	}

	return command
}

func reannounceTorrent(ctx context.Context, qb *qbittorrent.Client, interval, attempts int, hash string) error {
	announceOK := false
	attempt := 0

	time.Sleep(time.Duration(interval) * time.Millisecond)

	for attempt < attempts {
		trackers, err := qb.GetTorrentTrackers(ctx, hash)
		if err != nil {
			log.Fatalf("could not get trackers of torrent: %s err: %q", hash, err)
		}

		// check if status not working or something else
		_, working := findTrackerStatus(trackers, 2)
		if working {
			announceOK = true
			break
		}

		if err = qb.ReAnnounceTorrents(ctx, []string{hash}); err != nil {
			return err
		}

		time.Sleep(time.Duration(interval) * time.Millisecond)
		attempt++
		continue
	}

	if !announceOK {
		log.Println("announce still not ok")
		return errors.New("announce still not ok")
	}

	return nil
}
