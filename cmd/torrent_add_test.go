package cmd

import (
	"io"
	"strings"
	"testing"
)

// TestRunTorrentAdd_rejectsNameWithGlob guards the invariant that --name can
// only be applied to a single torrent. A glob can match many files, so the
// command must reject the combination before contacting qBittorrent.
func TestRunTorrentAdd_rejectsNameWithGlob(t *testing.T) {
	command := RunTorrentAdd()
	command.SetArgs([]string{"./files/*.torrent", "--name", "Some Name"})
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)

	err := command.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cannot use --name") {
		t.Fatalf("expected error containing %q, got %q", "cannot use --name", err.Error())
	}
}
