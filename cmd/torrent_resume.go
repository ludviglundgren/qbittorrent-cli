package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

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

		if resumeAll {
			hashes = []string{"all"}
		}

		// Split the hashes into groups of 20 to avoid flooding qbittorrent
		batch := 20
		for i := 0; i < len(hashes); i += batch {
			j := i + batch
			if j > len(hashes) {
				j = len(hashes)
			}

			if err := qb.ResumeCtx(ctx, hashes[i:j]); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not resume torrents: %v\n", err)
				os.Exit(1)
			}

			time.Sleep(time.Second * 1)
		}

		log.Printf("torrent(s) successfully resumed")
	}

	return command
}
