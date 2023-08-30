package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/autobrr/go-qbittorrent"
	"github.com/spf13/cobra"
)

// RunTorrentRemove cmd to remove torrents
func RunTorrentRemove() *cobra.Command {
	var (
		dryRun       bool
		removeAll    bool
		removePaused bool
		deleteFiles  bool
		hashes       []string
	)

	var command = &cobra.Command{
		Use:   "remove",
		Short: "Removes specified torrent(s)",
		Long:  `Removes torrents indicated by hash, name or a prefix of either. Whitespace indicates next prefix unless argument is surrounded by quotes`,
	}

	command.Flags().BoolVar(&dryRun, "dry-run", false, "Display what would be done without actually doing it")
	command.Flags().BoolVar(&removeAll, "all", false, "Removes all torrents")
	command.Flags().BoolVar(&removePaused, "paused", false, "Removes all paused torrents")
	command.Flags().BoolVar(&deleteFiles, "delete-files", false, "Also delete downloaded files from torrent(s)")
	command.Flags().StringSliceVar(&hashes, "hashes", []string{}, "Add hashes as comma separated list")

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

		ctx := cmd.Context()

		if err := qb.LoginCtx(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		if removeAll {
			hashes = []string{"all"}
		}

		if removePaused {
			pausedTorrents, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterPaused})
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: failed to retrieve paused torrents: %v\n", err)
				os.Exit(1)
			}

			for _, torrent := range pausedTorrents {
				hashes = append(hashes, torrent.Hash)
			}

			if dryRun {
				log.Printf("dry-run: found (%d) paused torrents to be removed\n", len(hashes))
			} else {
				log.Printf("found (%d) paused torrents to be removed\n", len(hashes))
			}
		}

		if len(hashes) == 0 {
			log.Println("No torrents found to remove")
			return
		}

		if dryRun {
			if hashes[0] == "all" {
				log.Println("dry-run: all torrents to be removed")
			} else {
				log.Printf("dry-run: (%d) torrents to be removed\n", len(hashes))
			}
		} else {
			if hashes[0] == "all" {
				log.Println("dry-run: all torrents to be removed")
			} else {
				log.Printf("dry-run: (%d) torrents to be removed\n", len(hashes))
			}

			err := batchRequests(hashes, func(start, end int) error {
				return qb.DeleteTorrentsCtx(ctx, hashes[start:end], deleteFiles)
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not delete torrents: %v\n", err)
				os.Exit(1)
				return
			}

			if hashes[0] == "all" {
				log.Println("successfully removed all torrents")
			} else {
				log.Printf("successfully removed (%d) torrents\n", len(hashes))
			}
		}

		log.Printf("torrent(s) successfully deleted\n")
	}

	return command
}
