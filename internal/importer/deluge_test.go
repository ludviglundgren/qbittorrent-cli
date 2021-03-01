package importer

import "testing"

func TestDelugeImport_Import(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name: "good_deluge_dir",
			opts: Options{
				DelugeDir: "../../test/import/deluge",
				QbitDir:   "../../test/output/",
			},
			wantErr: false,
		},
		{
			name: "bad_deluge_dir",
			opts: Options{
				DelugeDir: "../../test/import/",
				QbitDir:   "../../test/output/",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			di := &DelugeImport{}

			if err := di.Import(tt.opts); (err != nil) != tt.wantErr {
				t.Errorf("Import() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
