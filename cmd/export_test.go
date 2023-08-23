package cmd

import "testing"

func Test_processHashes(t *testing.T) {
	type args struct {
		sourceDir string
		exportDir string
		hashes    map[string]struct{}
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
				hashes: map[string]struct{}{
					"5ba4939a00a9b21629a0ad7d376898b768d997a3": {},
				},
				dry: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := processHashes(tt.args.sourceDir, tt.args.exportDir, tt.args.hashes, tt.args.dry, tt.args.verbose); (err != nil) != tt.wantErr {
				t.Errorf("processHashes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
