package cmd

import (
	"log"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/utils"
	"github.com/pkg/errors"

	"github.com/autobrr/go-qbittorrent"
	"github.com/spf13/cobra"
)

// RunTorrentResume cmd to resume torrents
func RunTorrentResume() *cobra.Command {
	var (
		resumeAll bool
		hashes    []string
	)

	var command = &cobra.Command{
		Use:   "resume",
		Short: "Resume specified torrent(s)",
		Long:  `Resumes torrents indicated by hash, name or a prefix of either. Whitespace indicates next prefix unless argument is surrounded by quotes`,
	}

	command.Flags().BoolVar(&resumeAll, "all", false, "resumes all torrents")
	command.Flags().StringSliceVar(&hashes, "hashes", []string{}, "Add hashes as comma separated list")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		if len(hashes) > 0 {
			if err := utils.ValidateHash(hashes); err != nil {
				return errors.Wrap(err, "invalid hashes supplied")
			}
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

		if resumeAll {
			hashes = []string{"all"}
		}

		err := batchRequests(hashes, func(start, end int) error {
			return qb.ResumeCtx(ctx, hashes[start:end])
		})
		if err != nil {
			return errors.Wrap(err, "could not resume torrents")
		}

		log.Printf("torrent(s) successfully resumed")

		return nil
	}

	return command
}

// batchRequests split into multiple requests because qbit uses application/x-www-form-urlencoded
// which might lead to too big requests
func batchRequests(hashes []string, fn func(start, end int) error) error {
	// Split the hashes into groups of 20 to avoid flooding qbittorrent
	batch := 25
	for i := 0; i < len(hashes); i += batch {
		j := i + batch
		if j > len(hashes) {
			j = len(hashes)
		}

		if err := fn(i, j); err != nil {
			return err
		}

		time.Sleep(time.Second * 1)
	}

	return nil
}
