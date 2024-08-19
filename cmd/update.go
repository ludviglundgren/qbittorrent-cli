package cmd

import (
	"log"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

func RunUpdate(version string) *cobra.Command {
	var command = &cobra.Command{
		Use:          "update",
		Short:        "Update qbittorrent-cli to latest version",
		Example:      `  qbt update`,
		SilenceUsage: false,
	}

	var verbose bool

	command.Flags().BoolVar(&verbose, "verbose", false, "Verbose output: Print changelog")

	command.Run = func(cmd *cobra.Command, args []string) {
		v, err := semver.ParseTolerant(version)
		if err != nil {
			log.Println("could not parse version:", err)
			return
		}

		latest, err := selfupdate.UpdateSelf(v, "ludviglundgren/qbittorrent-cli")
		if err != nil {
			log.Println("Binary update failed:", err)
			return
		}

		if latest.Version.Equals(v) {
			// latest version is the same as current version. It means current binary is up-to-date.
			log.Println("Current binary is the latest version", version)
		} else {
			log.Println("Successfully updated to version: ", latest.Version)

			if verbose {
				log.Println("Release note:\n", latest.ReleaseNotes)
			}
		}

		return
	}

	return command
}
