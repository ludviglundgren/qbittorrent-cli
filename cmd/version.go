package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func RunVersion(version, commit, date string) *cobra.Command {
	var command = &cobra.Command{
		Use:          "version",
		Short:        "Print the version",
		Example:      `  qbt version`,
		SilenceUsage: false,
	}
	command.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("Version:", version)
		fmt.Println("Commit:", commit)
		fmt.Println("Date:", date)
	}
	return command
}
