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

// RunRemove cmd to remove torrents
func RunRemove() *cobra.Command {
	var (
		removeAll		bool
		deleteFiles		bool
		hashes			bool 
		names			bool 
	)

	var command = &cobra.Command{
		Use:   "remove",
		Short: "Removes specified torrents",
		Long:  `Removes torrents indicated by hash, name or a prefix of either; 
				whitespace indicates next prefix unless argument is surrounded by quotes`,
	}

	command.Flags().BoolVar(&removeAll, "all", false, "Removes all torrents")
	command.Flags().BoolVar(&deleteFiles, "delete-files", false, "Also delete downloaded files from torrent(s)")
	command.Flags().BoolVar(&hashes, "hashes", false, "Provided arguments will be read as torrent hashes")
	command.Flags().BoolVar(&names, "names", false, "Provided arguments will be read as torrent names")

	command.Run = func(cmd *cobra.Command, args []string) {
		if !removeAll && len(args) < 1 {
			log.Printf("Please provide atleast one torrent hash/name as an argument")
			return
		}

		if !removeAll && !hashes && !names {
			log.Printf("Please specifiy if arguments are to be read as hashes or names (--hashes / --names)")
			return
		}

		config.InitConfig()
		qbtSettings := qbittorrent.Settings{
			Hostname: config.Qbit.Host,
			Port:	  config.Qbit.Port,
			Username: config.Qbit.Login,
			Password: config.Qbit.Password,
		}
		qb := qbittorrent.NewClient(qbtSettings)

		err := qb.Login()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}
		
		if removeAll {
			qb.DeleteTorrents([]string{"all"}, deleteFiles)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not delete torrents: %v\n", err)
				os.Exit(1)
			}

			log.Printf("All torrents removed successfully")
			return
		}

		foundTorrents, err := qb.GetTorrentsByPrefixes(args, hashes, names)
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

			qb.DeleteTorrents(hashesToRemove[i:j], deleteFiles)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not delete torrents: %v\n", err)
				os.Exit(1)
			}

			time.Sleep(time.Second * 1)
		}

		log.Printf("torrent(s) successfully deleted")
	}

	return command
}
