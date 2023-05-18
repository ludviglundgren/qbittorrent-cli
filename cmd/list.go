package cmd

import (
	"fmt"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// RunList cmd to list torrents
func RunList() *cobra.Command {
	var json bool

	var command = &cobra.Command{
		Use:   "list",
		Short: "List torrents",
		Long:  `List all torrents`,
	}
	command.Flags().BoolVar(&json, "json", false, "print to json")

	command.Run = func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		qbtSettings := qbittorrent.Settings{
			Hostname: config.Qbit.Host,
			Port:     config.Qbit.Port,
			Username: config.Qbit.Login,
			Password: config.Qbit.Password,
			SSL:      config.Qbit.SSL,
		}
		qb := qbittorrent.NewClient(qbtSettings)

		err := qb.Login()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		if json {
			torrents, err := qb.GetTorrentsRaw()
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not get torrents %v\n", err)
				os.Exit(1)
			}

			fmt.Println(torrents)
		} else {
			torrents, err := qb.GetTorrents()
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not get torrents %v\n", err)
				os.Exit(1)
			}

			printList(torrents)
		}
	}

	return command
}

func printList(torrents []qbittorrent.Torrent) {
	space := fmt.Sprintf("%*c", 4, ' ')
	for _, torrent := range torrents {
		fmt.Printf("[*] ")
		fmt.Printf("%-80.80s%s[%s]\n", torrent.Name, space, torrent.State)

		fmt.Printf("%sDownloaded: ", space)
		if torrent.AmountLeft <= 0 {
			fmt.Printf("%s%s", humanize.Bytes(uint64(torrent.TotalSize)), space)
		} else {
			fmt.Printf(
				"%s / %s%s",
				humanize.Bytes(uint64(torrent.Completed)),
				humanize.Bytes(uint64(torrent.TotalSize)),
				space,
			)
		}

		if torrent.Uploaded > 0 {
			fmt.Printf("Uploaded: %s%s", humanize.Bytes(uint64(torrent.Uploaded)), space)
		}

		if torrent.DlSpeed > 0 {
			fmt.Printf(
				"DL Speed: %s/s%s",
				humanize.Bytes(uint64(torrent.DlSpeed)),
				space,
			)
		} else if torrent.UpSpeed > 0 {
			fmt.Printf(
				"UP Speed: %s/s%s",
				humanize.Bytes(uint64(torrent.UpSpeed)),
				space,
			)
		}

		days := (torrent.TimeActive / (60 * 60 * 24))
		hours := (torrent.TimeActive / (60 * 60)) - (days * 24)
		minutes := (torrent.TimeActive / 60) - ((days * 1440) + (hours * 60))

		if days > 0 {
			fmt.Printf("Time Active: %dd %dh %dm\n", days, hours, minutes)
		} else if hours > 0 {
			fmt.Printf("Time Active: %dh %dm\n", hours, minutes)
		} else {
			fmt.Printf("Time Active: %dm\n", minutes)
		}

		fmt.Printf("%sSave Path: %s\n", space, torrent.SavePath)
		fmt.Printf("%sHash: %s\n", space, torrent.Hash)

		fmt.Println()
	}
}
