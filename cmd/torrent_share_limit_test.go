package cmd

import (
	"testing"

	"github.com/autobrr/go-qbittorrent"
)

func Test_validateShareLimits(t *testing.T) {
	tests := []struct {
		name                     string
		ratioLimit               float64
		seedingTimeLimit         int64
		inactiveSeedingTimeLimit int64
		wantErr                  bool
	}{
		{name: "defaults global", ratioLimit: -2, seedingTimeLimit: -2, inactiveSeedingTimeLimit: -2, wantErr: false},
		{name: "unlimited", ratioLimit: -1, seedingTimeLimit: -1, inactiveSeedingTimeLimit: -1, wantErr: false},
		{name: "explicit values", ratioLimit: 2.0, seedingTimeLimit: 1440, inactiveSeedingTimeLimit: 0, wantErr: false},
		{name: "zero ratio", ratioLimit: 0, seedingTimeLimit: -2, inactiveSeedingTimeLimit: -2, wantErr: false},
		{name: "invalid ratio below -2", ratioLimit: -3, seedingTimeLimit: -2, inactiveSeedingTimeLimit: -2, wantErr: true},
		{name: "invalid fractional ratio -1.5", ratioLimit: -1.5, seedingTimeLimit: -2, inactiveSeedingTimeLimit: -2, wantErr: true},
		{name: "invalid fractional ratio -0.5", ratioLimit: -0.5, seedingTimeLimit: -2, inactiveSeedingTimeLimit: -2, wantErr: true},
		{name: "invalid seeding-time", ratioLimit: -2, seedingTimeLimit: -3, inactiveSeedingTimeLimit: -2, wantErr: true},
		{name: "invalid inactive-seeding-time", ratioLimit: -2, seedingTimeLimit: -2, inactiveSeedingTimeLimit: -10, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateShareLimits(tt.ratioLimit, tt.seedingTimeLimit, tt.inactiveSeedingTimeLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateShareLimits() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_formatRatioLimit(t *testing.T) {
	tests := []struct {
		name  string
		ratio float64
		want  string
	}{
		{name: "global", ratio: -2, want: "global"},
		{name: "unlimited", ratio: -1, want: "unlimited"},
		{name: "zero", ratio: 0, want: "0.00"},
		{name: "value", ratio: 2.5, want: "2.50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatRatioLimit(tt.ratio); got != tt.want {
				t.Errorf("formatRatioLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formatTimeLimit(t *testing.T) {
	tests := []struct {
		name    string
		minutes int64
		want    string
	}{
		{name: "global", minutes: -2, want: "global"},
		{name: "unlimited", minutes: -1, want: "unlimited"},
		{name: "zero", minutes: 0, want: "0m"},
		{name: "value", minutes: 1440, want: "1440m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatTimeLimit(tt.minutes); got != tt.want {
				t.Errorf("formatTimeLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formatShareLimits(t *testing.T) {
	opts := qbittorrent.ShareLimitOptions{RatioLimit: 2.0, SeedingTimeLimit: 1440, InactiveSeedingTimeLimit: -2}
	want := "ratio=2.00, seeding-time=1440m, inactive-seeding-time=global"
	if got := formatShareLimits(opts); got != want {
		t.Errorf("formatShareLimits() = %v, want %v", got, want)
	}
}
