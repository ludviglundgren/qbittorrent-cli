package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/spf13/cobra"
)

// RunRemove cmd to remove torrents
func RunRemove() *cobra.Command {
	var (
		removeAll    bool
		removePaused bool
		deleteFiles  bool
		hashes       bool
		names        bool
		dryRun       bool
	)

	var command = &cobra.Command{
		Use:   "remove",
		Short: "Removes specified torrents",
		Long: `Removes torrents indicated by hash, name or a prefix of either; 
				whitespace indicates next prefix unless argument is surrounded by quotes`,
	}

	command.Flags().BoolVar(&removeAll, "all", false, "Removes all torrents")
	command.Flags().BoolVar(&removePaused, "paused", false, "Removes all paused torrents")
	command.Flags().BoolVar(&deleteFiles, "delete-files", false, "Also delete downloaded files from torrent(s)")
	command.Flags().BoolVar(&hashes, "hashes", false, "Provided arguments will be read as torrent hashes")
	command.Flags().BoolVar(&names, "names", false, "Provided arguments will be read as torrent names")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Display what would be done without actually doing it")

	command.Run = func(cmd *cobra.Command, args []string) {
		if !removeAll && !removePaused && len(args) < 1 {
			log.Printf("Please provide at least one torrent hash/name as an argument")
			return
		}

		if !removeAll && !removePaused && !hashes && !names {
			log.Printf("Please specify if arguments are to be read as hashes or names (--hashes / --names)")
			return
		}

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

		if removeAll {
			if dryRun {
				log.Printf("Would remove all torrents")
			} else {
				if err := qb.DeleteTorrents(ctx, []string{"all"}, deleteFiles); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: could not delete torrents: %v\n", err)
					os.Exit(1)
				}

				log.Printf("All torrents removed successfully")
			}
			return
		}

		if removePaused {
			pausedTorrents, err := qb.GetTorrentsWithFilters(ctx, &qbittorrent.GetTorrentsRequest{Filter: "paused"})
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: failed to retrieve paused torrents: %v\n", err)
				os.Exit(1)
			}

			hashesToRemove := []string{}
			for _, torrent := range pausedTorrents {
				hashesToRemove = append(hashesToRemove, torrent.Hash)
			}

			if len(hashesToRemove) < 1 {
				log.Printf("No paused torrents found to remove")
				return
			}

			if dryRun {
				log.Printf("Paused torrents to be removed: %v", hashesToRemove)
			} else {
				if err := qb.DeleteTorrents(ctx, hashesToRemove, deleteFiles); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: could not delete paused torrents: %v\n", err)
					os.Exit(1)
				}

				log.Print("Paused torrents removed successfully")
			}
			return
		}

		foundTorrents, err := qb.GetTorrentsByPrefixes(ctx, args, hashes, names)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to retrieve torrents: %v\n", err)
			os.Exit(1)
		}

		hashesToRemove := []string{}
		for _, torrent := range foundTorrents {
			hashesToRemove = append(hashesToRemove, torrent.Hash)
		}

		if len(hashesToRemove) < 1 {
			log.Printf("No torrents found to remove with provided search terms")
			return
		}

		// Split the hashes to remove into groups of 20 to avoid flooding qbittorrent
		batch := 20
		for i := 0; i < len(hashesToRemove); i += batch {
			j := i + batch
			if j > len(hashesToRemove) {
				j = len(hashesToRemove)
			}

			if err := qb.DeleteTorrents(ctx, hashesToRemove[i:j], deleteFiles); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not delete torrents: %v\n", err)
				os.Exit(1)
			}

			time.Sleep(time.Second * 1)
		}

		log.Printf("torrent(s) successfully deleted")
	}

	return command
}
