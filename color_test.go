package ics

import (
	"image/color"
	"testing"
)

func TestColorToHex(t *testing.T) {
	tests := []struct {
		name string
		c    color.Color
		want string
	}{
		{"RGBA solid", color.RGBA{R: 255, G: 0, B: 0, A: 255}, "#ff0000"},
		{"RGBA transparent", color.RGBA{R: 255, G: 0, B: 0, A: 128}, "#ff000080"},
		{"NRGBA solid", color.NRGBA{R: 0, G: 255, B: 0, A: 255}, "#00ff00"},
		{"NRGBA transparent", color.NRGBA{R: 0, G: 255, B: 0, A: 128}, "#00ff0080"},
		{"Generic solid (CMYK)", color.CMYK{C: 0, M: 255, Y: 255, K: 0}, "#ff0000"},
		{"Alpha short-circuit generic", color.Gray{Y: 128}, "#808080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := colorToHex(tt.c); got != tt.want {
				t.Errorf("colorToHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHexToColor(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		want    color.Color
		wantErr bool
	}{
		{"6 char exact", "#ff0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}, false},
		{"8 char exact", "#00ff0080", color.RGBA{R: 0, G: 255, B: 0, A: 128}, false},
		{"6 char no hash", "0000ff", color.RGBA{R: 0, G: 0, B: 255, A: 255}, false},
		{"invalid char", "#zz0000", nil, true},
		{"invalid length", "#f00", nil, true},
		{"icsx5 color test: NameAndColor", "lightblue", nil, true}, // We do not currently resolve css names to hex in hexToColor
		{"icsx5 color test: NameAndLegacyColor", "#123456", color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 255}, false},
		{"X11 color: red", "red", nil, true}, // RFC 7986 specifies CSS3 color names, but parser doesn't resolve them yet
		{"X11 color: papayawhip", "papayawhip", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hexToColor(tt.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("hexToColor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("hexToColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
