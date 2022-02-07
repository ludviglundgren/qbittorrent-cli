package cmd

import (
	"fmt"
	"os"
	"log"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/spf13/cobra"
)

// RunPause cmd to pause torrents
func RunPause() *cobra.Command {
	var (
		pauseAll	bool
		hashes		bool
		names		bool
	)

	var command = &cobra.Command{
		Use:   "pause",
		Short: "Pause specified torrents",
		Long:  `Pauses torrents indicated by hash, name or a prefix of either; 
				whitespace indicates next prefix unless argument is surrounded by quotes`,
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
			Hostname: config.Qbit.Host,
			Port:     config.Qbit.Port,
			Username: config.Qbit.Login,
			Password: config.Qbit.Password,
		}
		qb := qbittorrent.NewClient(qbtSettings)

		err := qb.Login()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		if pauseAll {
			qb.Pause([]string{"all"})
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not pause torrents: %v\n", err)
				os.Exit(1)
			}

			log.Printf("All torrents paused successfully")
			return
		}

		foundTorrents, err := qb.GetTorrentsByPrefixes(args, hashes, names)
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

			qb.Pause(hashesToPause[i:j])
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not pause torrents: %v\n", err)
				os.Exit(1)
			}

			time.Sleep(time.Second * 1)
		}

		log.Printf("torrent(s) successfully paused")
	}

	return command
}
