package main

import (
	"log"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log.SetFlags(0)

	rootCmd := cmd.NewRootCmd(version, commit, date)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
