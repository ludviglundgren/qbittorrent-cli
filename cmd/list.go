package cmd

import (
	"fmt"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/spf13/cobra"
)

// RunList cmd to list torrents
func RunList() *cobra.Command {
	var json bool

	var command = &cobra.Command{
		Use:   "list",
		Short: "List torrents",
		Long:  `List all torrents`,
	}
	command.Flags().BoolVar(&json, "json", false, "print to json")

	command.Run = func(cmd *cobra.Command, args []string) {
		qbtSettings := qbittorrent.Settings{
			Hostname: config.Qbit.Host,
			Port:     config.Qbit.Port,
			Username: config.Qbit.Login,
			Password: config.Qbit.Password,
		}
		qb := qbittorrent.NewClient(qbtSettings)

		err := qb.Login()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		if json {
			torrents, err := qb.GetTorrentsRaw()
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not get torrents %v\n", err)
				os.Exit(1)
			}

			fmt.Println(torrents)
		} else {
			torrents, err := qb.GetTorrents()
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not get torrents %v\n", err)
				os.Exit(1)
			}

			fmt.Println(torrents)
		}
	}

	return command
}
