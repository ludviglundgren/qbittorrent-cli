package cmd

import (
	"fmt"
	"os"
	"strings"
	"log"

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

		torrents, err := qb.GetTorrents()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not retrieve torrents: %v\n", err)
			os.Exit(1)
		}

		foundHashes := map[string]bool{}
		for _, torrent := range torrents {
			if removeAll {
				foundHashes[torrent.Hash] = true
				continue
			}

			if hashes {
				for _, targetHash := range args {
					if strings.HasPrefix(torrent.Hash, targetHash) {
						foundHashes[torrent.Hash] = true
						break
					}
				}
			}

			if names {
				for _, targetName := range args {
					if strings.HasPrefix(torrent.Name, targetName) {
						foundHashes[torrent.Hash] = true
						break
					}
				}
			}
		}

		hashesToRemove := []string{}
		for hash := range foundHashes {
			hashesToRemove = append(hashesToRemove, hash)
		}

		err = qb.DeleteTorrents(hashesToRemove, deleteFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not delete torrents: %v\n", err)
			os.Exit(1)
		}

		log.Printf("torrent(s) successfully deleted")
	}

	return command
}
