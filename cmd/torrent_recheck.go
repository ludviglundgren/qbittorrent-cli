package cmd

import (
	"log"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/utils"

	"github.com/autobrr/go-qbittorrent"
	"github.com/pkg/errors"
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
		Example: `  qbt torrent recheck --hashes HASH
  qbt torrent recheck --hashes HASH1,HASH2
`,
	}

	command.Flags().StringSliceVar(&hashes, "hashes", []string{}, "Add hashes as comma separated list")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		if len(hashes) == 0 {
			return errors.Errorf("no hashes supplied to recheck")
		}

		err := utils.ValidateHash(hashes)
		if err != nil {
			return errors.Wrap(err, "invalid hashes supplied")
		}

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
			return errors.Wrap(err, "could not login to qbit")
		}

		err = batchRequests(hashes, func(start, end int) error {
			return qb.RecheckCtx(ctx, hashes[start:end])
		})
		if err != nil {
			return errors.Wrap(err, "could not reched torrents")
		}

		log.Printf("torrent(s) successfully recheckd")

		return nil
	}

	return command
}
