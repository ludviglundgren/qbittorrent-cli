package cmd

import (
	"log"

	"github.com/ludviglundgren/qbittorrent-cli/v2/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/v2/pkg/utils"

	"github.com/autobrr/go-qbittorrent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RunTorrentRename cmd to rename torrents
func RunTorrentRename() *cobra.Command {
	var command = &cobra.Command{
		Use:     "rename <hash> <name>",
		Short:   "Rename specified torrent",
		Long:    `Rename the torrent indicated by the supplied hash.`,
		Example: `  qbt torrent rename HASH "New Torrent Name"`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a torrent hash as first argument")
			}

			if len(args) < 2 {
				return errors.New("requires a new name as second argument")
			}

			if len(args) > 2 {
				return errors.New("accepts exactly 2 arguments: <hash> <name>")
			}

			return nil
		},
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		hash := args[0]
		name := args[1]

		if err := utils.ValidateHash([]string{hash}); err != nil {
			return errors.Wrap(err, "invalid hashes supplied")
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

		if err := qb.SetTorrentNameCtx(ctx, hash, name); err != nil {
			return errors.Wrapf(err, "could not rename torrent: %s", hash)
		}

		log.Printf("torrent %s successfully renamed to %q", hash, name)

		return nil
	}

	return command
}
