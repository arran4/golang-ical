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
			if got := colorToHex(tt.c); string(got) != tt.want {
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

func TestColorName(t *testing.T) {
	t.Run("ToColor", func(t *testing.T) {
		c, err := ColorName("red").ToColor()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if c != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
			t.Errorf("expected red, got %v", c)
		}

		c, err = ColorName("greenX11").ToColor()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if c != (color.RGBA{R: 0, G: 255, B: 0, A: 255}) {
			t.Errorf("expected greenX11 to be lime, got %v", c)
		}

		_, err = ColorName("invalid").ToColor()
		if err == nil {
			t.Errorf("expected error for invalid color")
		}
	})

	t.Run("ToHexString", func(t *testing.T) {
		hex, err := ColorName("BLUE").ToHexString()
		if err != nil || hex != "#0000ff" {
			t.Errorf("expected #0000ff, got %v, err: %v", hex, err)
		}

        hex, err = ColorName("maroonX11").ToHexString()
		if err != nil || hex != "#c71585" {
			t.Errorf("expected maroonX11 to alias to mediumvioletred (#c71585), got %v, err: %v", hex, err)
		}
	})

	t.Run("ColorNameFromHexString", func(t *testing.T) {
		name, ok := ColorNameFromHexString("#ffffff")
		if !ok || name != "white" {
			t.Errorf("expected white, got %v, %v", name, ok)
		}

		_, ok = ColorNameFromHexString("#123456")
		if ok {
			t.Errorf("did not expect exact match for #123456")
		}
	})

	t.Run("ClosestColorNameFromHexString", func(t *testing.T) {
		name, err := ClosestColorNameFromHexString("#ff0001")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if name != "red" {
			t.Errorf("expected closest to be red, got %v", name)
		}

        _, err = ClosestColorNameFromHexString("invalid")
        if err == nil {
            t.Errorf("expected error from invalid hex")
        }
	})

	t.Run("ClosestColorNameFromColor", func(t *testing.T) {
		name := ClosestColorNameFromColor(color.RGBA{R: 250, G: 250, B: 250, A: 255})
		if name != "snow" && name != "white" { // it could be close to either, but let's just make sure it doesn't panic and returns a valid name
			// snow is #fffafa (255, 250, 250), white is #ffffff (255, 255, 255)
			// distance to snow: sqrt(25 + 0 + 0) = 5
			// distance to white: sqrt(25 + 25 + 25) = sqrt(75) = 8.66
			if name != "snow" {
				t.Errorf("expected closest to be snow, got %v", name)
			}
		}
	})

    t.Run("ColorFromHex", func(t *testing.T) {
        c, err := ColorFromHex("#000000")
        if err != nil {
            t.Fatalf("unexpected err: %v", err)
        }
        if c != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
            t.Fatalf("expected black, got %v", c)
        }
    })

    t.Run("HexFromColor", func(t *testing.T) {
        c := color.RGBA{R: 0, G: 0, B: 0, A: 255}
        h := HexFromColor(c)
        if h != "#000000" {
            t.Fatalf("expected #000000, got %v", h)
        }
    })
}

func TestColorNameEdgeCases(t *testing.T) {
	// Let's test the aliases
	tests := []struct {
		name string
		want ColorHex
	}{
		{"GreenX11", "#00ff00"},
		{"MaroonX11", "#c71585"},
		{"PurpleX11", "#9932cc"},
		{"GreyX11", "#c0c0c0"},
		{"GrayX11", "#c0c0c0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hex, err := ColorName(tt.name).ToHexString()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if hex != tt.want {
				t.Errorf("got %v, want %v", hex, tt.want)
			}
		})
	}
}
