package importer

import (
	"testing"
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
				SourceDir: "../../test/import/sessions/",
				QbitDir:   "../../test/output/rtorrent",
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
