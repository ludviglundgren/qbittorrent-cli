package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RunApp cmd for application info
func RunApp() *cobra.Command {
	var command = &cobra.Command{
		Use:   "app",
		Short: "app subcommand",
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
		Long:  ``,
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("get app info")
		return nil
	}

	return command
}
