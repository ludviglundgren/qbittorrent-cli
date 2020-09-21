package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/l3uddz/go-qbittorrent/qbt"
	"github.com/spf13/cobra"
)

// RunList cmd to list torrents
func RunList() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "List torrents",
		Long:  `List all torrents`,
	}

	command.Run = func(cmd *cobra.Command, args []string) {
		connStr := fmt.Sprintf("http://%v:%v/", config.Qbit.Host, config.Qbit.Port)
		qb := qbt.NewClient(connStr)

		loginOpts := qbt.LoginOptions{
			Username: config.Qbit.Login,
			Password: config.Qbit.Password,
		}
		err := qb.Login(loginOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		torrents, err := qb.Torrents(nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not get torrents %v\n", err)
			os.Exit(1)
		}

		res, _ := json.Marshal(torrents)

		fmt.Println(string(res))
	}

	return command
}
