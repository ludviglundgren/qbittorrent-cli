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

// RunTorrentPause cmd to pause torrents
func RunTorrentPause() *cobra.Command {
	var (
		pauseAll bool
		hashes   bool
		names    bool
	)

	var command = &cobra.Command{
		Use:   "pause",
		Short: "Pause specified torrent(s)",
		Long:  `Pauses torrents indicated by hash, name or a prefix of either. Whitespace indicates next prefix unless argument is surrounded by quotes`,
	}

	command.Flags().BoolVar(&pauseAll, "all", false, "Pauses all torrents")
	command.Flags().BoolVar(&hashes, "hashes", false, "Provided arguments will be read as torrent hashes")
	command.Flags().BoolVar(&names, "names", false, "Provided arguments will be read as torrent names")

	command.Run = func(cmd *cobra.Command, args []string) {
		if !pauseAll && len(args) < 1 {
			log.Printf("Please provide atleast one torrent hash/name as an argument")
			return
		}

		if !pauseAll && !hashes && !names {
			log.Printf("Please specifiy if arguments are to be read as hashes or names (--hashes / --names)")
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

		if pauseAll {
			if err := qb.Pause(ctx, []string{"all"}); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not pause torrents: %v\n", err)
				os.Exit(1)
			}

			log.Printf("All torrents paused successfully")
			return
		}

		foundTorrents, err := qb.GetTorrentsByPrefixes(ctx, args, hashes, names)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to retrieve torrents: %v\n", err)
			os.Exit(1)
		}

		hashesToPause := []string{}
		for _, torrent := range foundTorrents {
			hashesToPause = append(hashesToPause, torrent.Hash)
		}

		if len(hashesToPause) < 1 {
			log.Printf("No torrents found to pause with provided search terms")
			return
		}

		// Split the hashes to pause into groups of 20 to avoid flooding qbittorrent
		batch := 20
		for i := 0; i < len(hashesToPause); i += batch {
			j := i + batch
			if j > len(hashesToPause) {
				j = len(hashesToPause)
			}

			if err := qb.Pause(ctx, hashesToPause[i:j]); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not pause torrents: %v\n", err)
				os.Exit(1)
			}

			time.Sleep(time.Second * 1)
		}

		log.Printf("torrent(s) successfully paused")
	}

	return command
}
