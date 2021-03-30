package importer

import (
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/ludviglundgren/qbittorrent-cli/pkg/torrent"
)

func TestRTorrentImport_Import(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name: "import_from_rtorrent",
			opts: Options{
				QbitDir:   "../../test/output/rtorrent",
				SourceDir: "../../test/import/sessions/",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &RTorrentImport{}
			if err := i.Import(tt.opts); (err != nil) != tt.wantErr {
				t.Errorf("Import() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestRTorrentImport_ImportFastResume(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name: "import_from_rtorrent",
			opts: Options{
				QbitDir:   "../../test/output/rtorrent",
				SourceDir: "../../test/import/sessions/",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//i := &RTorrentImport{}
			//if err := i.Import(tt.opts); (err != nil) != tt.wantErr {
			//	t.Errorf("Import() error = %v, wantErr %v", err, tt.wantErr)
			//}
			matches, _ := filepath.Glob(tt.opts.QbitDir + "/*.fastresume")

			for _, match := range matches {
				torrentFile, err := torrent.OpenDecodeRaw(match)
				if err != nil {
					log.Printf("Can't decode torrent file %v. Continue", match)
					continue
				}
				//hash := torrent.CalculateInfoHash(torrentFile)
				fmt.Printf("file: %+v\n", torrentFile)

				//torrentFile2, err := torrent.OpenDecodeRaw("../../sync/qbit/post-start/" + hash + ".fastresume")
				//if err != nil {
				//	log.Printf("Can't decode torrent file %v. Continue", match)
				//	continue
				//}
				//fmt.Printf("file: %+v\n", torrentFile2)
			}
		})
	}
}
