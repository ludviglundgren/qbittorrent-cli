package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/autobrr/go-qbittorrent"
	"github.com/spf13/cobra"
)

// RunTorrentTracker cmd for torrent tracker operations
func RunTorrentTracker() *cobra.Command {
	var command = &cobra.Command{
		Use:   "tracker",
		Short: "Torrent tracker subcommand",
		Long:  `Do various torrent category operations`,
	}

	command.AddCommand(RunTorrentTrackerEdit())

	return command
}

// RunTorrentTrackerEdit cmd for torrent tracker operations
func RunTorrentTrackerEdit() *cobra.Command {
	var command = &cobra.Command{
		Use:     "edit",
		Short:   "Edit torrent tracker",
		Long:    `Edit tracker for torrents via hashes`,
		Example: `  qbt torrent tracker edit --old url.old/test --new url.com/test`,
	}

	var (
		dry    bool
		oldURL string
		newURL string
	)

	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().StringVar(&oldURL, "old", "", "Old tracker URL to replace")
	command.Flags().StringVar(&newURL, "new", "", "New tracker URL")

	command.RunE = func(cmd *cobra.Command, args []string) error {
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
			fmt.Fprintf(os.Stderr, "could not login to qbit: %q\n", err)
			os.Exit(1)
		}

		if dry {
			log.Printf("dry-run: successfully updated tracker on torrents\n")

			return nil
		} else {
			torrents, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{})
			if err != nil {
				log.Fatalf("could not get torrents err: %q\n", err)
			}

			matches := 0

			for _, torrent := range torrents {
				if strings.Contains(torrent.Tracker, oldURL) {
					if err := qb.EditTrackerCtx(ctx, torrent.Hash, torrent.Tracker, newURL); err != nil {
						log.Fatalf("could not edit tracker for torrent: %s\n", torrent.Hash)
					}

					matches++
				}
			}

			log.Printf("successfully updated tracker for (%d) torrents\n", matches)
		}

		return nil
	}

	return command
}
