package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/autobrr/go-qbittorrent"
	"github.com/spf13/cobra"
)

// RunTorrentPause cmd to pause torrents
func RunTorrentPause() *cobra.Command {
	var (
		pauseAll bool
		names    bool
		hashes   []string
	)

	var command = &cobra.Command{
		Use:   "pause",
		Short: "Pause specified torrent(s)",
		Long:  `Pauses torrents indicated by hash, name or a prefix of either. Whitespace indicates next prefix unless argument is surrounded by quotes`,
	}

	command.Flags().BoolVar(&pauseAll, "all", false, "Pauses all torrents")
	command.Flags().StringSliceVar(&hashes, "hashes", []string{}, "Add hashes as comma separated list")
	command.Flags().BoolVar(&names, "names", false, "Provided arguments will be read as torrent names")

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

		if pauseAll {
			hashes = []string{"all"}
		}

		if len(hashes) == 0 {
			log.Printf("No torrents found to pause with provided search terms")
			return
		}

		err := batchRequests(hashes, func(start, end int) error {
			return qb.PauseCtx(ctx, hashes[start:end])
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not pause torrents: %v\n", err)
			os.Exit(1)
			return
		}

		log.Printf("torrent(s) successfully paused")
	}

	return command
}
