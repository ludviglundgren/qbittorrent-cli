package cmd

import (
	"io"
	"strings"
	"testing"
)

// TestRunTorrentPause_requiresTarget mirrors the resume guard (issue #132):
// `qbt torrent pause` with no target must fail loudly instead of silently
// no-opping. Both error paths return before any network call, so the command
// can be exercised without a live qBittorrent.
func TestRunTorrentPause_requiresTarget(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "no target returns error instead of no-op success",
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
			command := RunTorrentPause()
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
