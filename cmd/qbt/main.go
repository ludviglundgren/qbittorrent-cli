package main

import (
	"log"
	"os"
	"runtime/debug"

	"github.com/ludviglundgren/qbittorrent-cli/v2/cmd"
)

// Set via -ldflags by goreleaser. When absent (go install, go build) the
// values are filled in from the embedded build info instead.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log.SetFlags(0)

	version, commit, date = buildInfo(version, commit, date)

	rootCmd := cmd.NewRootCmd(version, commit, date)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// buildInfo fills any value not injected via ldflags from the module and VCS
// metadata the Go toolchain embeds in the binary. `go install pkg@v2.4.0`
// stamps the module version, and a build inside a git checkout stamps the
// revision and commit time.
func buildInfo(version, commit, date string) (string, string, string) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return version, commit, date
	}

	if version == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		version = bi.Main.Version
	}

	var modified bool
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			if commit == "none" {
				commit = s.Value
			}
		case "vcs.time":
			if date == "unknown" {
				date = s.Value
			}
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}

	if modified && commit != "none" {
		commit += "-dirty"
	}

	return version, commit, date
}
