package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"
)

// RunTag tag torrents
func RunTag() *cobra.Command {
	var (
		tagUnregistered bool

		sourceHost string
		sourcePort uint
		sourceUser string
		sourcePass string
	)

	var command = &cobra.Command{
		Use:   "tag",
		Short: "tag torrents",
		Long:  `tag torrents`,
	}
	command.Flags().BoolVar(&tagUnregistered, "unregistered", false, "tag unregistered")

	command.Flags().StringVar(&sourceHost, "host", "", "Source host")
	command.Flags().UintVar(&sourcePort, "port", 0, "Source host")
	command.Flags().StringVar(&sourceUser, "user", "", "Source user")
	command.Flags().StringVar(&sourcePass, "pass", "", "Source pass")

	command.Run = func(cmd *cobra.Command, args []string) {
		qbtSettings := qbittorrent.Settings{
			Hostname: sourceHost,
			Port:     sourcePort,
			Username: sourceUser,
			Password: sourcePass,
		}
		qb := qbittorrent.NewClient(qbtSettings)

		err := qb.Login()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		sourceData, err := qb.GetTorrents()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not get torrents %v\n", err)
			os.Exit(1)
		}

		unregisteredTorrentIDs := make([]string, 0)

		var totalSize uint64
		var unregisteredSize uint64
		for _, t := range sourceData {
			if tagUnregistered {
				if t.Tracker == "" {
					unregisteredTorrentIDs = append(unregisteredTorrentIDs, t.Hash)
					unregisteredSize += uint64(t.Size)
				}
			}
			totalSize += uint64(t.Size)
		}

		fmt.Printf("Total torrents (%d) with a total size of: %v\n", len(sourceData), humanize.Bytes(totalSize))

		// --unregistered add tag unregistered
		if tagUnregistered {
			fmt.Printf("Reclaimable space from (%d) unregistered torrents: %v\n", len(unregisteredTorrentIDs), humanize.Bytes(unregisteredSize))

			// Split the slice into batches of 20 items.
			batch := 20
			for i := 0; i < len(unregisteredTorrentIDs); i += batch {
				j := i + batch
				if j > len(unregisteredTorrentIDs) {
					j = len(unregisteredTorrentIDs)
				}

				qb.SetTag(unregisteredTorrentIDs[i:j], "unregistered")

				// sleep before next request
				time.Sleep(time.Second * 1)
			}
		}
	}

	return command
}
