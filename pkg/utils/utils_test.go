package utils

import "testing"

func TestValidateHash(t *testing.T) {
	type args struct {
		hashes []string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{name: "ok", args: args{hashes: []string{"6957bf5272f5b994132458a557864e3ea747489f"}}, want: true, wantErr: false},
		{name: "invalid", args: args{hashes: []string{"6957bf5272f5b994132458a557864e3ea747489"}}, want: false, wantErr: true},
		{name: "invalid_2", args: args{hashes: []string{"11111"}}, want: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHash(tt.args.hashes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
