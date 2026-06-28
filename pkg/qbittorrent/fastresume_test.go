package qbittorrent

import (
	"reflect"
	"testing"

	"github.com/zeebo/bencode"
)

func TestTrackerTiers_UnmarshalBencode(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    [][]string
		wantErr bool
	}{
		{
			name:  "nested tiers (standard qBittorrent format)",
			input: [][]string{{"https://a/announce"}, {"udp://b:1337"}},
			want:  [][]string{{"https://a/announce"}, {"udp://b:1337"}},
		},
		{
			name:  "nested tier with multiple urls",
			input: [][]string{{"https://a/announce", "https://a-backup/announce"}},
			want:  [][]string{{"https://a/announce", "https://a-backup/announce"}},
		},
		{
			// this is the shape that produced "Cannot store string into []string"
			name:  "flat list of urls",
			input: []string{"https://a/announce", "udp://b:1337"},
			want:  [][]string{{"https://a/announce"}, {"udp://b:1337"}},
		},
		{
			name:  "single url string",
			input: "https://a/announce",
			want:  [][]string{{"https://a/announce"}},
		},
		{
			name:  "single whitespace separated string",
			input: "https://a/announce udp://b:1337",
			want:  [][]string{{"https://a/announce"}, {"udp://b:1337"}},
		},
		{
			name:  "empty list",
			input: []string{},
			want:  nil,
		},
		{
			name:    "unsupported type",
			input:   int64(5),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := bencode.EncodeBytes(tt.input)
			if err != nil {
				t.Fatalf("could not encode test input: %v", err)
			}

			var got TrackerTiers
			err = bencode.DecodeBytes(encoded, &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalBencode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual([][]string(got), tt.want) {
				t.Errorf("UnmarshalBencode() = %#v, want %#v", [][]string(got), tt.want)
			}
		})
	}
}

// TestFastresume_DecodeFlatTrackers ensures a full fastresume that stores the
// trackers field as a flat list of strings can still be decoded. This is the
// exact failure reported in issue #135.
func TestFastresume_DecodeFlatTrackers(t *testing.T) {
	raw := map[string]interface{}{
		"save_path": "/downloads",
		"trackers":  []string{"https://a/announce", "udp://b:1337"},
	}

	encoded, err := bencode.EncodeBytes(raw)
	if err != nil {
		t.Fatalf("could not encode fastresume: %v", err)
	}

	var fr Fastresume
	if err := bencode.DecodeBytes(encoded, &fr); err != nil {
		t.Fatalf("could not decode fastresume with flat trackers: %v", err)
	}

	want := [][]string{{"https://a/announce"}, {"udp://b:1337"}}
	if !reflect.DeepEqual([][]string(fr.Trackers), want) {
		t.Errorf("Trackers = %#v, want %#v", [][]string(fr.Trackers), want)
	}
}
