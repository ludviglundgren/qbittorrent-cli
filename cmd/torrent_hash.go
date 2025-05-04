package cmd

import (
	"fmt"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RunTorrentHash cmd to add torrents
func RunTorrentHash() *cobra.Command {
	var command = &cobra.Command{
		Use:     "hash",
		Short:   "Print the hash of a torrent file or magnet",
		Example: `  qbt torrent hash file.torrent`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a torrent file or magnet URI as first argument")
			}

			return nil
		},
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		hash := ""

		if strings.HasPrefix(filePath, "magnet:") {
			metadata, err := metainfo.ParseMagnetUri(filePath)
			if err != nil {
				return errors.Wrapf(err, "could not parse magnet URI: %s", filePath)
			}

			hash = metadata.InfoHash.HexString()

		} else {
			metadata, err := metainfo.LoadFromFile(filePath)
			if err != nil {
				return errors.Wrapf(err, "could not parse torrent file: %s", filePath)
			}

			hash = metadata.HashInfoBytes().HexString()
		}

		fmt.Println(hash)

		return nil
	}

	return command
}
