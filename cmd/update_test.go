package cmd

import "testing"

func TestIsReleaseVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version string
		want    bool
	}{
		{"v2.4.0", true},
		{"2.4.0", true},
		{"v2.0.0-beta1", true},
		{"v2.0.0-rc.0", true},
		{"dev", false},
		{"", false},
		{"(devel)", false},
		{"v2.3.1-0.20260909123709-334cd70e1416", false},
		{"v2.3.1-0.20260909123709-334cd70e1416+dirty", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()
			if got := isReleaseVersion(tt.version); got != tt.want {
				t.Errorf("isReleaseVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}
