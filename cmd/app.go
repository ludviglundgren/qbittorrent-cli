package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/autobrr/go-qbittorrent"
	"github.com/spf13/cobra"
)

// RunApp cmd for application info
func RunApp() *cobra.Command {
	var command = &cobra.Command{
		Use:   "app",
		Short: "App subcommand",
		Long:  "Do various app actions",
	}

	command.AddCommand(RunAppVersion())

	return command
}

// RunAppVersion cmd to view application info
func RunAppVersion() *cobra.Command {
	var command = &cobra.Command{
		Use:   "version",
		Short: "Get qBittorrent version info",
	}

	var (
		output string
	)

	command.Flags().StringVar(&output, "output", "", "Print as [formatted text (default), json]")

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

		appVersion, err := qb.GetAppVersionCtx(ctx)
		if err != nil {
			log.Fatal("could not get app version")
		}

		webapiVersion, err := qb.GetWebAPIVersionCtx(ctx)
		if err != nil {
			log.Fatal("could not get web api version")
		}

		switch output {
		case "json":
			fmt.Printf(`{"app_version":"%s","api_version":"%s"}`, appVersion, webapiVersion)

		default:
			fmt.Printf("qBittorrent version info:\nApp version: %s\nAPI version: %s\n", appVersion, webapiVersion)
		}

		return nil
	}

	return command
}
