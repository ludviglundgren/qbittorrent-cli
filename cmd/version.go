package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func RunVersion(version, commit, date string) *cobra.Command {
	var output string
	var command = &cobra.Command{
		Use:   "version",
		Short: "Print qbt version info",
		Example: `  qbt version
  qbt version --output json`,
		SilenceUsage: false,
	}

	command.Flags().StringVar(&output, "output", "text", "Print as [text, json]")

	command.Run = func(cmd *cobra.Command, args []string) {
		switch output {
		case "text":
			fmt.Printf(`qbt - qbitttorrent cli
Version: %s
Commit: %s
Date: %s
`, version, commit, date)
			return

		case "json":
			res, err := json.Marshal(versionInfo{Version: version, Commit: commit, Date: date})
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not marshal version info to json %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(res))

		}
	}

	return command
}

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}
