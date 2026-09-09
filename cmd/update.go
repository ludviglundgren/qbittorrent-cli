package cmd

import (
	"log"

	"github.com/blang/semver"
	"github.com/pkg/errors"
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

	command.RunE = func(cmd *cobra.Command, args []string) error {
		if !isReleaseVersion(version) {
			return errors.Errorf("self-update is only available for release builds, current version is %q. Reinstall with your package manager or run: go install github.com/ludviglundgren/qbittorrent-cli/v2/cmd/qbt@latest", version)
		}

		v, err := semver.ParseTolerant(version)
		if err != nil {
			return errors.Wrapf(err, "could not parse version string: %s", version)
		}

		latest, err := selfupdate.UpdateSelf(v, "ludviglundgren/qbittorrent-cli")
		if err != nil {
			return errors.Wrap(err, "binary update failed")
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

		return nil
	}

	return command
}

// isReleaseVersion reports whether version looks like a tagged release
// (v2.4.0) rather than a dev build or a Go pseudo-version
// (v2.3.1-0.20260909120000-5251a66d1c2f).
func isReleaseVersion(version string) bool {
	if version == "dev" || version == "" {
		return false
	}

	v, err := semver.ParseTolerant(version)
	if err != nil {
		return false
	}

	// pseudo-versions carry a "0.<timestamp>-<hash>" prerelease
	if len(v.Pre) == 2 && v.Pre[0].IsNum && v.Pre[0].VersionNum == 0 && !v.Pre[1].IsNum {
		return false
	}

	return true
}
