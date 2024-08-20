package cmd

import (
	"testing"

	"github.com/autobrr/go-qbittorrent"
)

func Test_processTorrentTags(t *testing.T) {
	type args struct {
		torrent              qbittorrent.Torrent
		trackers             []qbittorrent.TorrentTracker
		removeTaggedTorrents *tagData
		unregisteredTorrents *tagData
		notWorkingTorrents   *tagData
		tagUnregistered      bool
		tagNotWorking        bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				torrent: qbittorrent.Torrent{Hash: "1111", Tracker: "", Tags: "Unregistered"},
				trackers: []qbittorrent.TorrentTracker{
					{Url: "http://test.local", Status: 2, NumPeers: 1, NumSeeds: 1, NumLeechers: 0, NumDownloaded: 1, Message: "OK"},
				},
				removeTaggedTorrents: &tagData{Hashes: []string{}, HashTagMap: map[string][]string{DefaultTagUnregistered.String(): {}, DefaultTagNotWorking.String(): {}}},
				unregisteredTorrents: &tagData{Hashes: []string{}},
				notWorkingTorrents:   &tagData{Hashes: []string{}},
				tagUnregistered:      true,
				tagNotWorking:        true,
			},
			want: true,
		},
		{
			name: "test",
			args: args{
				torrent: qbittorrent.Torrent{Hash: "1111", Tracker: "", Tags: ""},
				trackers: []qbittorrent.TorrentTracker{
					{Url: "http://test.local", Status: 1, NumPeers: 1, NumSeeds: 1, NumLeechers: 0, NumDownloaded: 1, Message: "Unregistered torrent"},
				},
				removeTaggedTorrents: &tagData{Hashes: []string{}, HashTagMap: map[string][]string{DefaultTagUnregistered.String(): {}, DefaultTagNotWorking.String(): {}}},
				unregisteredTorrents: &tagData{Hashes: []string{}},
				notWorkingTorrents:   &tagData{Hashes: []string{}},
				tagUnregistered:      true,
				tagNotWorking:        true,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processTorrentTags(tt.args.torrent, tt.args.trackers, tt.args.removeTaggedTorrents, tt.args.unregisteredTorrents, tt.args.notWorkingTorrents, tt.args.tagUnregistered, tt.args.tagNotWorking); got != tt.want {
				t.Errorf("processTorrent() = %v, want %v", got, tt.want)
			}
		})
	}
}
