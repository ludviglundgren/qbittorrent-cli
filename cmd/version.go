package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
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

	command.RunE = func(cmd *cobra.Command, args []string) error {
		switch output {
		case "text":
			fmt.Printf(`qbt - qbitttorrent cli
Version: %s
Commit: %s
Date: %s
`, version, commit, date)
			return nil

		case "json":
			res, err := json.Marshal(versionInfo{Version: version, Commit: commit, Date: date})
			if err != nil {

				return errors.Wrap(err, "could not marshal version info to json")
			}
			fmt.Println(string(res))

		}

		return nil
	}

	return command
}

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}
