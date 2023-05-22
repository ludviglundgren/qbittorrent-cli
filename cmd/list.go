package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// RunList cmd to list torrents
func RunList() *cobra.Command {
	var (
		filter     = "all"
		category   string
		tag        string
		hashes     string
		outputJson bool
	)

	var command = &cobra.Command{
		Use:     "list",
		Short:   "List torrents",
		Long:    `List all torrents, or torrents with a specific filters. Get by filter, category, tag and hashes. Can be combined`,
		Example: `qbt list --filter=downloading --category=linux-iso`,
	}
	command.Flags().BoolVar(&outputJson, "json", false, "Print to json")
	command.Flags().StringVarP(&filter, "filter", "f", "all", "Filter by state. Available filters: all, downloading, seeding, completed, paused, active, inactive, resumed, \nstalled, stalled_uploading, stalled_downloading, errored")
	command.Flags().StringVarP(&category, "category", "c", "", "Filter by category. All categories by default.")
	command.Flags().StringVarP(&tag, "tag", "t", "", "Filter by tag. Single tag: tag1")
	command.Flags().StringVarP(&hashes, "hashes", "h", "", "Filter by hashes. Separated by | pipe: \"hash1|hash2\".")

	command.Run = func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		qbtSettings := qbittorrent.Settings{
			Addr:      config.Qbit.Addr,
			Hostname:  config.Qbit.Host,
			Port:      config.Qbit.Port,
			Username:  config.Qbit.Login,
			Password:  config.Qbit.Password,
			BasicUser: config.Qbit.BasicUser,
			BasicPass: config.Qbit.BasicPass,
		}

		qb := qbittorrent.NewClient(qbtSettings)

		ctx := context.Background()

		if err := qb.Login(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
			os.Exit(1)
		}

		req := qbittorrent.GetTorrentsRequest{
			Filter:   strings.ToLower(filter),
			Category: category,
			Tag:      tag,
			Hashes:   hashes,
		}

		// get torrent list with default filter of all
		torrents, err := qb.GetTorrentsWithFilters(ctx, &req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not get torrents %v\n", err)
			os.Exit(1)
		}

		if len(torrents) < 1 {
			fmt.Printf("No torrents found with filter: %s\n", filter)
			return
		}

		if outputJson {
			res, err := json.Marshal(torrents)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not marshal torrents to json %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(res))
			return
		}

		printList(torrents)
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

		days := torrent.TimeActive / (60 * 60 * 24)
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
