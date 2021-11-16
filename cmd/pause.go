package cmd

import (
	"fmt"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/spf13/cobra"
)

// RunPause cmd to pause torrents
func RunPause() *cobra.Command {
	var command = &cobra.Command{
		Use:   "pause",
		Short: "Pause all torrents",
		Long:  `Pause all torrents`,
	}
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

		err = qb.Pause(nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not pause torrents %v\n", err)
			os.Exit(1)
		}
	}

	return command
}
