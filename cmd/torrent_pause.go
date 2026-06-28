package cmd

import (
	"log"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/utils"

	"github.com/autobrr/go-qbittorrent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RunTorrentPause cmd to pause torrents
func RunTorrentPause() *cobra.Command {
	var (
		pauseAll bool
		hashes   []string
	)

	var command = &cobra.Command{
		Use:   "pause",
		Short: "Pause specified torrent(s)",
		Long:  `Pause the torrent(s) indicated by the supplied hash(es), or pause every torrent with --all.`,
		Example: `  qbt torrent pause --all
  qbt torrent pause HASH1 HASH2
  qbt torrent pause --hashes HASH1,HASH2`,
	}

	command.Flags().BoolVar(&pauseAll, "all", false, "Pauses all torrents")
	command.Flags().StringSliceVar(&hashes, "hashes", []string{}, "Add hashes as comma separated list")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		// Treat positional arguments as hashes too, so `qbt torrent pause HASH` works
		// alongside the --hashes flag.
		hashes = append(hashes, args...)

		if pauseAll {
			hashes = []string{"all"}
		} else {
			if len(hashes) == 0 {
				return errors.New("no torrents specified: provide hash(es) as arguments or with --hashes, or use --all")
			}

			if err := utils.ValidateHash(hashes); err != nil {
				return errors.Wrap(err, "invalid hashes supplied")
			}
		}

		config.InitConfig()

		qbtSettings := qbittorrent.Config{
			Host:      config.Qbit.Addr,
			APIKey:    config.Qbit.APIKey,
			Username:  config.Qbit.Login,
			Password:  config.Qbit.Password,
			BasicUser: config.Qbit.BasicUser,
			BasicPass: config.Qbit.BasicPass,
		}

		qb := qbittorrent.NewClient(qbtSettings)

		ctx := cmd.Context()

		if err := qb.LoginCtx(ctx); err != nil {
			return errors.Wrap(err, "could not login to qbit")
		}

		if err := batchRequests(hashes, func(start, end int) error {
			return qb.PauseCtx(ctx, hashes[start:end])
		}); err != nil {
			return errors.Wrap(err, "could not pause torrents")
		}

		log.Printf("torrent(s) successfully paused")

		return nil
	}

	return command
}
