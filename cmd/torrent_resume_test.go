package cmd

import (
	"io"
	"strings"
	"testing"
)

// TestRunTorrentResume_requiresTarget guards against the false-positive reported
// in issue #132: `qbt torrent resume` with no target must fail loudly instead of
// printing "torrent(s) successfully resumed". Both error paths return before any
// network call, so the command can be exercised without a live qBittorrent.
func TestRunTorrentResume_requiresTarget(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "no target returns error instead of false-positive success",
			args:    []string{},
			wantErr: "no torrents specified",
		},
		{
			name:    "invalid positional hash is rejected",
			args:    []string{"not-a-valid-hash"},
			wantErr: "invalid hashes supplied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := RunTorrentResume()
			command.SetArgs(tt.args)
			command.SetOut(io.Discard)
			command.SetErr(io.Discard)

			err := command.Execute()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}
