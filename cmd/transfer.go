package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/autobrr/go-qbittorrent"
	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RunTransfer cmd for transfer info
func RunTransfer() *cobra.Command {
	var command = &cobra.Command{
		Use:   "transfer",
		Short: "Transfer info subcommand",
		Long:  "Get status and speed info",
	}

	command.AddCommand(RunAppVersion())

	return command
}

// RunTransferInfo cmd to view transfer info
func RunTransferInfo() *cobra.Command {
	var command = &cobra.Command{
		Use:   "info",
		Short: "Get qBittorrent transfer info",
	}

	var (
		output string
	)

	command.Flags().StringVar(&output, "output", "", "Print as [formatted text (default), json]")

	command.RunE = func(cmd *cobra.Command, args []string) error {
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

		info, err := qb.GetTransferInfo()
		if err != nil {
			return errors.Wrap(err, "could not get app version")
		}

		switch output {
		case "json":
			res, err := json.Marshal(info)
			if err != nil {
				return errors.Wrap(err, "could not marshal transfer info")
			}

			fmt.Println(string(res))

		default:
			fmt.Printf("qBittorrent transfer info:\nConnection status: %s\nSession upload: %s\n\nSession download: %s\n", info.ConnectionStatus, humanize.Bytes(uint64(info.UpInfoData)), humanize.Bytes(uint64(info.DlInfoData)))
		}

		return nil
	}

	return command
}
