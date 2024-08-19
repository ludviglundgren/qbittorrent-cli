package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/autobrr/go-qbittorrent"
	"github.com/spf13/cobra"
)

// RunTorrentRecheck cmd to recheck torrents
func RunTorrentRecheck() *cobra.Command {
	var (
		hashes []string
	)

	var command = &cobra.Command{
		Use:   "recheck",
		Short: "Recheck specified torrent(s)",
		Long:  `Rechecks torrents indicated by hash(es).`,
	}

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
			fmt.Fprintf(os.Stderr, "connection failed: %v\n", err)
			os.Exit(1)
		}

		if len(hashes) == 0 {
			log.Println("No torrents found to recheck")
			return
		}

		err := batchRequests(hashes, func(start, end int) error {
			return qb.RecheckCtx(ctx, hashes[start:end])
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not recheck torrents: %v\n", err)
			os.Exit(1)
			return
		}

		log.Printf("torrent(s) successfully recheckd")
	}

	return command
}
