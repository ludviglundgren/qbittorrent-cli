package cmd

import (
	"testing"

	"github.com/magiconair/properties/assert"

	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"
)

func Test_compare(t *testing.T) {
	type args struct {
		source  []qbittorrent.Torrent
		compare []qbittorrent.Torrent
	}
	tests := []struct {
		name    string
		args    args
		matches int
		wantErr bool
	}{
		{
			name: "",
			args: args{
				source: []qbittorrent.Torrent{
					{Name: "test_1", Hash: "1ED3AEA7B65B9EFF17F85AB8240AFB896BA6A6FD"},
					{Name: "test_2", Hash: "CB93E8D27A32F80FB28FFD9BD07BF0A25912F3A5"},
				},
				compare: []qbittorrent.Torrent{
					{Name: "test_1", Hash: "1ED3AEA7B65B9EFF17F85AB8240AFB896BA6A6FD"},
				},
			},
			matches: 1,
			wantErr: false,
		},
		{
			name: "",
			args: args{
				source: []qbittorrent.Torrent{
					{Name: "test_2", Hash: "CB93E8D27A32F80FB28FFD9BD07BF0A25912F3A5"},
				},
				compare: []qbittorrent.Torrent{
					{Name: "test_1", Hash: "1ED3AEA7B65B9EFF17F85AB8240AFB896BA6A6FD"},
				},
			},
			matches: 0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if matches, err := compare(tt.args.source, tt.args.compare); (err != nil) != tt.wantErr {
				assert.Equal(t, matches, tt.matches)
			}
		})
	}
}
