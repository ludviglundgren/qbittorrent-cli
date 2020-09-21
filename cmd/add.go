package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/l3uddz/go-qbittorrent/qbt"
	"github.com/spf13/cobra"
)

// RunAdd cmd to add torrents
func RunAdd() *cobra.Command {
	var dry bool

	var command = &cobra.Command{
		Use:   "add",
		Short: "Add torrent",
		Long:  `Add new torrent to Deluge from file`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a torrent file as first argument")
			}

			return nil
		},
	}
	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")

	command.Run = func(cmd *cobra.Command, args []string) {
		// args
		// first arg is path to torrent file
		filePath := args[0]

		if !dry {
			connStr := fmt.Sprintf("http://%v:%v/", config.Qbit.Host, config.Qbit.Port)
			qb := qbt.NewClient(connStr)

			err := qb.Login(qbt.LoginOptions{
				Username: config.Qbit.Login,
				Password: config.Qbit.Password,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
				os.Exit(1)

			}

			options := map[string]string{}
			success, err := qb.DownloadFromFile(filePath, options)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: adding torrent failed: %v\n", err)
				os.Exit(1)
			}

			if !success {
				fmt.Fprintf(os.Stderr, "ERROR: adding torrent failed: %v\n", err)
			}

			fmt.Print("Torrent successfully added!\n")
		} else {
			fmt.Println("dry-run: Torrent successfully added!")
		}
	}

	return command
}
