package cmd

import (
	"fmt"
	"os"
	"math"
	"strconv"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

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
		qbtSettings := qbittorrent.Settings{
			Hostname: config.Qbit.Host,
			Port:     config.Qbit.Port,
			Username: config.Qbit.Login,
			Password: config.Qbit.Password,
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
	for _, torrent := range torrents {
		fmt.Printf("%-60s\t[Status: %s]\n", torrent.Name, torrent.State)

		fmt.Printf(
			"Downloaded: %s / %s%10s",
			bytesToHumanReadable(float64(torrent.Completed)),
			bytesToHumanReadable(float64(torrent.TotalSize)),
			"",
		)

		if torrent.DlSpeed > 0 {
			fmt.Printf(
				"DL Speed: %s/s%10s\t",
				bytesToHumanReadable(float64(torrent.DlSpeed)),
				"",
			)
		} else if torrent.UpSpeed > 0 {
			fmt.Printf(
				"UP Speed: %s/s%10s\t",
				bytesToHumanReadable(float64(torrent.UpSpeed)),
				"",
			)
		} else {
			fmt.Printf("%27s\t", "")
		}

		hours := (torrent.TimeActive / 60) / 60
		minutes := (torrent.TimeActive / 60) - (hours * 60)
		fmt.Printf("Time Active: %dh%dm\n", hours, minutes)

		fmt.Printf("Save Path: %s\n", torrent.SavePath)
		fmt.Println()
	}
}

func bytesToHumanReadable(size float64) string {
	if size <= 0 {
		return "0.0KB"
	}

	var suffixes [5]string
	suffixes[0] = "B"
	suffixes[1] = "KB"
	suffixes[2] = "MB"
	suffixes[3] = "GB"
	suffixes[4] = "TB"

	base := math.Log(size) / math.Log(1024)
	newSize := math.Pow(1024, base - math.Floor(base))
	suffix := suffixes[int(math.Floor(base))]

	return strconv.FormatFloat(newSize, 'f', 1, 64) + suffix
}
