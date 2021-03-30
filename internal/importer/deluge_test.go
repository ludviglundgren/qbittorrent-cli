package importer

import "testing"

func TestDelugeImport_Import(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name: "import_from_deluge",
			opts: Options{
				SourceDir: "../../test/import/deluge",
				QbitDir:   "../../test/output/deluge",
			},
			wantErr: false,
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
