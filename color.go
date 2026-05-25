package ics

import (
	"fmt"
	"image/color"
)

func colorToHex(c color.Color) string {
	switch c := c.(type) {
	case color.NRGBA:
		if c.A < 255 {
			return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
		}
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	case color.RGBA:
		if c.A < 255 {
			return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
		}
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	r, g, b, a := c.RGBA()
	if a < 65535 {
		return fmt.Sprintf("#%02x%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
	}
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// hexToColor takes a hex color string like "#FF0000" or "#FF0000FF" and returns a color.Color.
func hexToColor(s string) (color.Color, error) {
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	switch len(s) {
	case 6: // RRGGBB
		var r, g, b uint8
		_, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return nil, err
		}
		return color.RGBA{R: r, G: g, B: b, A: 255}, nil
	case 8: // RRGGBBAA
		var r, g, b, a uint8
		_, err := fmt.Sscanf(s, "%02x%02x%02x%02x", &r, &g, &b, &a)
		if err != nil {
			return nil, err
		}
		return color.RGBA{R: r, G: g, B: b, A: a}, nil
	default:
		return nil, fmt.Errorf("invalid hex color length: %s", s)
	}
}
