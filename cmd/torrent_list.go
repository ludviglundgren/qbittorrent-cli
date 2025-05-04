package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/utils"

	"github.com/autobrr/go-qbittorrent"
	"github.com/autobrr/go-qbittorrent/errors"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// RunTorrentList cmd to list torrents
func RunTorrentList() *cobra.Command {
	var (
		filter   = "all"
		category string
		tag      string
		hashes   []string
		output   string
	)

	var command = &cobra.Command{
		Use:     "list",
		Short:   "List torrents",
		Long:    `List all torrents, or torrents with a specific filters. Get by filter, category, tag and hashes. Can be combined`,
		Example: `qbt torrent list --filter=downloading --category=linux-iso`,
	}
	command.Flags().StringVar(&output, "output", "", "Print as [formatted text (default), json]")
	command.Flags().StringVarP(&filter, "filter", "f", "all", "Filter by state. Available filters: all, downloading, seeding, completed, paused, active, inactive, resumed, \nstalled, stalled_uploading, stalled_downloading, errored")
	command.Flags().StringVarP(&category, "category", "c", "", "Filter by category. All categories by default.")
	command.Flags().StringVarP(&tag, "tag", "t", "", "Filter by tag. Single tag: tag1")
	command.Flags().StringSliceVar(&hashes, "hashes", []string{}, "Filter by hashes. Separated by comma: \"hash1,hash2\".")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		if len(hashes) > 0 {
			if err := utils.ValidateHash(hashes); err != nil {
				return errors.Wrap(err, "invalid hashes supplied")
			}
		}

		config.InitConfig()

		qbtSettings := qbittorrent.Config{
			Host:      config.Qbit.Addr,
			Username:  config.Qbit.Login,
			Password:  config.Qbit.Password,
			BasicUser: config.Qbit.BasicUser,
			BasicPass: config.Qbit.BasicPass,
		}

		qb := qbittorrent.NewClient(qbtSettings)

		ctx := cmd.Context()

		if err := qb.LoginCtx(ctx); err != nil {
			return errors.Wrap(err, "could not login to qbit")
		}

		req := qbittorrent.TorrentFilterOptions{
			Filter:   qbittorrent.TorrentFilter(strings.ToLower(filter)),
			Category: category,
			Tag:      tag,
			Hashes:   hashes,
		}

		// get torrent list with default filter of all
		torrents, err := qb.GetTorrentsCtx(ctx, req)
		if err != nil {
			return errors.Wrap(err, "could not get torrents")
		}

		if len(torrents) == 0 {
			log.Printf("No torrents found with filter: %s\n", filter)
			return nil
		}

		switch output {
		case "json":
			res, err := json.Marshal(torrents)
			if err != nil {
				return errors.Wrap(err, "could not marshal torrents list to json")
			}
			fmt.Println(string(res))

		default:
			if err := printList(torrents); err != nil {
				return errors.Wrap(err, "could not print torrent list")
			}
		}

		return nil
	}

	return command
}

var torrentItemTemplate = `{{ range .}}
[*] {{.Name}}
    Hash: {{.Hash}} State: {{.State}}
    Category: {{.Category}} Tags: [{{.Tags}}]
    Size: {{.Size}} Downloaded: {{.Completed}} / {{.TotalSize}} Uploaded: {{.Uploaded}}
    Added: {{.Added}} Time Active: {{.TimeActive}}
    Save path: {{.SavePath}}
{{end}}
`

type ItemData struct {
	Hash       string
	Name       string
	State      string
	Category   string
	Tags       string
	SavePath   string
	Added      string
	TimeActive string
	Size       string
	TotalSize  string
	Completed  string
	Uploaded   string
}

func printList(torrents []qbittorrent.Torrent) error {
	tmpl, err := template.New("item").Parse(torrentItemTemplate)
	if err != nil {
		return err
	}

	var data []ItemData

	for _, torrent := range torrents {

		//if torrent.DlSpeed > 0 {
		//	fmt.Printf(
		//		"DL Speed: %s/s%s",
		//		humanize.Bytes(uint64(torrent.DlSpeed)),
		//		space,
		//	)
		//} else if torrent.UpSpeed > 0 {
		//	fmt.Printf(
		//		"UP Speed: %s/s%s",
		//		humanize.Bytes(uint64(torrent.UpSpeed)),
		//		space,
		//	)
		//}

		d := ItemData{
			Hash:       torrent.Hash,
			Name:       torrent.Name,
			State:      string(torrent.State),
			Size:       humanize.Bytes(uint64(torrent.Size)),
			TotalSize:  humanize.Bytes(uint64(torrent.TotalSize)),
			Category:   torrent.Category,
			Tags:       torrent.Tags,
			SavePath:   torrent.SavePath,
			Added:      humanize.RelTime(time.Unix(0, 0), time.Unix(int64(torrent.AddedOn), 0), "ago", ""),
			TimeActive: humanize.RelTime(time.Unix(0, 0), time.Unix(int64(torrent.TimeActive), 0), "", ""),
			Completed:  humanize.Bytes(uint64(torrent.Completed)),
			Uploaded:   humanize.Bytes(uint64(torrent.Uploaded)),
		}

		data = append(data, d)
	}

	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		return err
	}

	return nil
}
