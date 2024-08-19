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

	command.MarkFlagRequired("old")
	command.MarkFlagRequired("new")

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

		torrents, err := qb.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{})
		if err != nil {
			log.Fatalf("could not get torrents err: %q\n", err)
		}

		var torrentsToUpdate []qbittorrent.Torrent

		for _, torrent := range torrents {
			if strings.Contains(torrent.Tracker, oldURL) {
				torrentsToUpdate = append(torrentsToUpdate, torrent)
			}
		}

		if len(torrentsToUpdate) == 0 {
			log.Printf("found no torrents with tracker %q\n", oldURL)
			return nil
		}

		for i, torrent := range torrentsToUpdate {
			if dry {
				log.Printf("dry-run: [%d/%d] updating tracker for torrent %s %q\n", i+1, len(torrentsToUpdate), torrent.Hash, torrent.Name)

			} else {
				log.Printf("[%d/%d] updating tracker for torrent %s %q\n", i+1, len(torrentsToUpdate), torrent.Hash, torrent.Name)

				if err := qb.EditTrackerCtx(ctx, torrent.Hash, torrent.Tracker, newURL); err != nil {
					log.Fatalf("could not edit tracker for torrent: %s\n", torrent.Hash)
				}
			}
		}

		log.Printf("successfully updated tracker for (%d) torrents\n", len(torrentsToUpdate))

		return nil
	}

	return command
}
