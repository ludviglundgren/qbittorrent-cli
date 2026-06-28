package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/autobrr/go-qbittorrent"
	"github.com/zeebo/bencode"
)

func Test_export_processHashes(t *testing.T) {
	type args struct {
		sourceDir string
		exportDir string
		hashes    map[string]qbittorrent.Torrent
		dry       bool
		verbose   bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_1",
			args: args{
				sourceDir: "../test/config/qBittorrent/BT_backup",
				exportDir: "../test/export",
				hashes: map[string]qbittorrent.Torrent{
					"5ba4939a00a9b21629a0ad7d376898b768d997a3": {},
				},
				dry: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := processExport(tt.args.sourceDir, tt.args.exportDir, tt.args.hashes, tt.args.dry, tt.args.verbose); (err != nil) != tt.wantErr {
				t.Errorf("processExport() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// minimal single-file torrent with no announce, which forces processExport down
// the path that reads the matching .fastresume to recover the trackers.
const torrentNoAnnounce = "d4:infod6:lengthi1e4:name1:a12:piece lengthi16384e6:pieces0:ee"

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("could not write %s: %v", path, err)
	}
}

// Test_processExport_continuesOnBadFastresume verifies that a single problematic
// fastresume file does not abort the whole export: the bad file is skipped and the
// remaining torrents are still exported (issue #135).
func Test_processExport_continuesOnBadFastresume(t *testing.T) {
	sourceDir := t.TempDir()
	exportDir := t.TempDir()

	const badHash = "1111111111111111111111111111111111111111"  // sorts first
	const goodHash = "2222222222222222222222222222222222222222" // sorts after the bad one

	// bad pair: valid torrent but an unparseable fastresume
	writeFile(t, filepath.Join(sourceDir, badHash+".torrent"), []byte(torrentNoAnnounce))
	writeFile(t, filepath.Join(sourceDir, badHash+".fastresume"), []byte("this is not bencode"))

	// good pair: valid torrent + fastresume that stores trackers as a flat list
	goodFastresume, err := bencode.EncodeBytes(map[string]interface{}{
		"save_path": "/downloads",
		"trackers":  []string{"https://tracker/announce"},
	})
	if err != nil {
		t.Fatalf("could not encode good fastresume: %v", err)
	}
	writeFile(t, filepath.Join(sourceDir, goodHash+".torrent"), []byte(torrentNoAnnounce))
	writeFile(t, filepath.Join(sourceDir, goodHash+".fastresume"), goodFastresume)

	hashes := map[string]qbittorrent.Torrent{
		badHash:  {},
		goodHash: {},
	}

	if err := processExport(sourceDir, exportDir, hashes, false, false); err != nil {
		t.Fatalf("processExport() returned error, want nil: %v", err)
	}

	// the good torrent (and its fastresume) must have been exported
	mustExist(t, filepath.Join(exportDir, goodHash+".torrent"))
	mustExist(t, filepath.Join(exportDir, goodHash+".fastresume"))

	// the bad torrent must have been skipped, not exported
	mustNotExist(t, filepath.Join(exportDir, badHash+".torrent"))
	mustNotExist(t, filepath.Join(exportDir, badHash+".fastresume"))

	// the flat-list trackers from the fastresume must have been recovered into the
	// exported .torrent (announce + announce-list), proving the end-to-end fix
	mi, err := metainfo.LoadFromFile(filepath.Join(exportDir, goodHash+".torrent"))
	if err != nil {
		t.Fatalf("could not load exported torrent: %v", err)
	}
	if mi.Announce != "https://tracker/announce" {
		t.Errorf("exported torrent Announce = %q, want %q", mi.Announce, "https://tracker/announce")
	}
	wantAnnounceList := metainfo.AnnounceList{{"https://tracker/announce"}}
	if !reflect.DeepEqual(mi.AnnounceList, wantAnnounceList) {
		t.Errorf("exported torrent AnnounceList = %#v, want %#v", mi.AnnounceList, wantAnnounceList)
	}
}

// minimal single-file torrent that already carries an announce URL. Torrents like
// this (pre v4.5.x qBittorrent) keep the trackers in the .torrent, so processExport
// takes the plain copy path instead of rebuilding from the .fastresume.
const torrentWithAnnounce = "d8:announce17:http://t/announce4:infod6:lengthi1e4:name1:a12:piece lengthi16384e6:pieces0:ee"

// Test_processExport_copyBranchExportsRealFastresume guards the copy path: the
// exported .fastresume must be a copy of the SOURCE .fastresume, not of the
// .torrent file (the latter was a real bug). The copy path only runs from the
// second announce-carrying torrent onwards, so we export two of them.
func Test_processExport_copyBranchExportsRealFastresume(t *testing.T) {
	sourceDir := t.TempDir()
	exportDir := t.TempDir()

	const firstHash = "1111111111111111111111111111111111111111"
	const secondHash = "2222222222222222222222222222222222222222"

	torrentBytes := []byte(torrentWithAnnounce)
	secondFastresume := []byte("FASTRESUME-CONTENT-FOR-SECOND-TORRENT")

	writeFile(t, filepath.Join(sourceDir, firstHash+".torrent"), torrentBytes)
	writeFile(t, filepath.Join(sourceDir, firstHash+".fastresume"), []byte("FASTRESUME-CONTENT-FOR-FIRST-TORRENT"))
	writeFile(t, filepath.Join(sourceDir, secondHash+".torrent"), torrentBytes)
	writeFile(t, filepath.Join(sourceDir, secondHash+".fastresume"), secondFastresume)

	hashes := map[string]qbittorrent.Torrent{
		firstHash:  {},
		secondHash: {},
	}

	if err := processExport(sourceDir, exportDir, hashes, false, false); err != nil {
		t.Fatalf("processExport() returned error, want nil: %v", err)
	}

	// the second torrent goes through the copy path; its exported .fastresume must
	// byte-equal the source .fastresume (and must NOT be the .torrent content)
	got, err := os.ReadFile(filepath.Join(exportDir, secondHash+".fastresume"))
	if err != nil {
		t.Fatalf("could not read exported fastresume: %v", err)
	}
	if !bytes.Equal(got, secondFastresume) {
		t.Errorf("exported fastresume = %q, want source fastresume %q", got, secondFastresume)
	}
	if bytes.Equal(got, torrentBytes) {
		t.Errorf("exported fastresume wrongly contains the .torrent bytes")
	}
}

// Test_processExport_allFailReturnsError verifies that when every matched torrent
// fails to export, processExport surfaces an error instead of reporting success.
func Test_processExport_allFailReturnsError(t *testing.T) {
	sourceDir := t.TempDir()
	exportDir := t.TempDir()

	const hash = "1111111111111111111111111111111111111111"
	writeFile(t, filepath.Join(sourceDir, hash+".torrent"), []byte(torrentNoAnnounce))
	writeFile(t, filepath.Join(sourceDir, hash+".fastresume"), []byte("this is not bencode"))

	hashes := map[string]qbittorrent.Torrent{hash: {}}

	if err := processExport(sourceDir, exportDir, hashes, false, false); err == nil {
		t.Fatal("processExport() returned nil, want error when all torrents fail")
	}

	// nothing should have been written to the export dir
	mustNotExist(t, filepath.Join(exportDir, hash+".torrent"))
	mustNotExist(t, filepath.Join(exportDir, hash+".fastresume"))
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %s (%v)", path, err)
	}
}

func mustNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file to not exist: %s (err=%v)", path, err)
	}
}
