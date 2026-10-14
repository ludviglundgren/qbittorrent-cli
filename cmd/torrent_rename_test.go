package cmd

import (
	"io"
	"strings"
	"testing"
)

// TestRunTorrentRename_validatesArguments covers the error paths of
// `qbt torrent rename <hash> <name>`. All of them return before any network
// call, so the command can be exercised without a live qBittorrent.
func TestRunTorrentRename_validatesArguments(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "no arguments returns error",
			args:    []string{},
			wantErr: "requires a torrent hash",
		},
		{
			name:    "missing new name returns error",
			args:    []string{"1234567890abcdef1234567890abcdef12345678"},
			wantErr: "requires a new name",
		},
		{
			name:    "too many arguments returns error",
			args:    []string{"1234567890abcdef1234567890abcdef12345678", "New Name", "extra"},
			wantErr: "accepts exactly 2 arguments",
		},
		{
			name:    "invalid hash is rejected",
			args:    []string{"not-a-valid-hash", "New Name"},
			wantErr: "invalid hashes supplied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := RunTorrentRename()
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
