package ics

import (
	"fmt"
	"image/color"
	"math"
	"strings"
)

// ColorName represents a CSS/X11 standard named color string.
type ColorName string

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

// Standard mapping of X11/CSS3 color names to hex codes.
// Note: where W3C/CSS and X11 clash (e.g. green, gray), we default to the W3C spec as it's standard for web,
// but include the alternate names.
var colorNameToHex = map[ColorName]string{
	"aliceblue":            "#f0f8ff",
	"antiquewhite":         "#faebd7",
	"aqua":                 "#00ffff",
	"aquamarine":           "#7fffd4",
	"azure":                "#f0ffff",
	"beige":                "#f5f5dc",
	"bisque":               "#ffe4c4",
	"black":                "#000000",
	"blanchedalmond":       "#ffebcd",
	"blue":                 "#0000ff",
	"blueviolet":           "#8a2be2",
	"brown":                "#a52a2a",
	"burlywood":            "#deb887",
	"cadetblue":            "#5f9ea0",
	"chartreuse":           "#7fff00",
	"chocolate":            "#d2691e",
	"coral":                "#ff7f50",
	"cornflowerblue":       "#6495ed",
	"cornsilk":             "#fff8dc",
	"crimson":              "#dc143c",
	"cyan":                 "#00ffff",
	"darkblue":             "#00008b",
	"darkcyan":             "#008b8b",
	"darkgoldenrod":        "#b8860b",
	"darkgray":             "#a9a9a9",
	"darkgreen":            "#006400",
	"darkgrey":             "#a9a9a9",
	"darkkhaki":            "#bdb76b",
	"darkmagenta":          "#8b008b",
	"darkolivegreen":       "#556b2f",
	"darkorange":           "#ff8c00",
	"darkorchid":           "#9932cc",
	"darkred":              "#8b0000",
	"darksalmon":           "#e9967a",
	"darkseagreen":         "#8fbc8f",
	"darkslateblue":        "#483d8b",
	"darkslategray":        "#2f4f4f",
	"darkslategrey":        "#2f4f4f",
	"darkturquoise":        "#00ced1",
	"darkviolet":           "#9400d3",
	"deeppink":             "#ff1493",
	"deepskyblue":          "#00bfff",
	"dimgray":              "#696969",
	"dimgrey":              "#696969",
	"dodgerblue":           "#1e90ff",
	"firebrick":            "#b22222",
	"floralwhite":          "#fffaf0",
	"forestgreen":          "#228b22",
	"fuchsia":              "#ff00ff",
	"gainsboro":            "#dcdcdc",
	"ghostwhite":           "#f8f8ff",
	"gold":                 "#ffd700",
	"goldenrod":            "#daa520",
	"gray":                 "#808080",
	"green":                "#008000",
	"greenyellow":          "#adff2f",
	"grey":                 "#808080",
	"honeydew":             "#f0fff0",
	"hotpink":              "#ff69b4",
	"indianred":            "#cd5c5c",
	"indigo":               "#4b0082",
	"ivory":                "#fffff0",
	"khaki":                "#f0e68c",
	"lavender":             "#e6e6fa",
	"lavenderblush":        "#fff0f5",
	"lawngreen":            "#7cfc00",
	"lemonchiffon":         "#fffacd",
	"lightblue":            "#add8e6",
	"lightcoral":           "#f08080",
	"lightcyan":            "#e0ffff",
	"lightgoldenrodyellow": "#fafad2",
	"lightgray":            "#d3d3d3",
	"lightgreen":           "#90ee90",
	"lightgrey":            "#d3d3d3",
	"lightpink":            "#ffb6c1",
	"lightsalmon":          "#ffa07a",
	"lightseagreen":        "#20b2aa",
	"lightskyblue":         "#87cefa",
	"lightslategray":       "#778899",
	"lightslategrey":       "#778899",
	"lightsteelblue":       "#b0c4de",
	"lightyellow":          "#ffffe0",
	"lime":                 "#00ff00",
	"limegreen":            "#32cd32",
	"linen":                "#faf0e6",
	"magenta":              "#ff00ff",
	"maroon":               "#800000",
	"mediumaquamarine":     "#66cdaa",
	"mediumblue":           "#0000cd",
	"mediumorchid":         "#ba55d3",
	"mediumpurple":         "#9370db",
	"mediumseagreen":       "#3cb371",
	"mediumslateblue":      "#7b68ee",
	"mediumspringgreen":    "#00fa9a",
	"mediumturquoise":      "#48d1cc",
	"mediumvioletred":      "#c71585",
	"midnightblue":         "#191970",
	"mintcream":            "#f5fffa",
	"mistyrose":            "#ffe4e1",
	"moccasin":             "#ffe4b5",
	"navajowhite":          "#ffdead",
	"navy":                 "#000080",
	"oldlace":              "#fdf5e6",
	"olive":                "#808000",
	"olivedrab":            "#6b8e23",
	"orange":               "#ffa500",
	"orangered":            "#ff4500",
	"orchid":               "#da70d6",
	"palegoldenrod":        "#eee8aa",
	"palegreen":            "#98fb98",
	"paleturquoise":        "#afeeee",
	"palevioletred":        "#db7093",
	"papayawhip":           "#ffefd5",
	"peachpuff":            "#ffdab9",
	"peru":                 "#cd853f",
	"pink":                 "#ffc0cb",
	"plum":                 "#dda0dd",
	"powderblue":           "#b0e0e6",
	"purple":               "#800080",
	"rebeccapurple":        "#663399",
	"red":                  "#ff0000",
	"rosybrown":            "#bc8f8f",
	"royalblue":            "#4169e1",
	"saddlebrown":          "#8b4513",
	"salmon":               "#fa8072",
	"sandybrown":           "#f4a460",
	"seagreen":             "#2e8b57",
	"seashell":             "#fff5ee",
	"sienna":               "#a0522d",
	"silver":               "#c0c0c0",
	"skyblue":              "#87ceeb",
	"slateblue":            "#6a5acd",
	"slategray":            "#708090",
	"slategrey":            "#708090",
	"snow":                 "#fffafa",
	"springgreen":          "#00ff7f",
	"steelblue":            "#4682b4",
	"tan":                  "#d2b48c",
	"teal":                 "#008080",
	"thistle":              "#d8bfd8",
	"tomato":               "#ff6347",
	"turquoise":            "#40e0d0",
	"violet":               "#ee82ee",
	"wheat":                "#f5deb3",
	"white":                "#ffffff",
	"whitesmoke":           "#f5f5f5",
	"yellow":               "#ffff00",
	"yellowgreen":          "#9acd32",
	// X11 specific alternate names that clash with W3C
	"greenX11":   "#00ff00",
	"grayX11":    "#bebebe",
	"greyX11":    "#bebebe",
	"maroonX11":  "#b03060",
	"purpleX11":  "#a020f0",
}

// ToColor converts the ColorName to an image/color.Color exactly.
func (cn ColorName) ToColor() (color.Color, error) {
	hex, ok := colorNameToHex[ColorName(strings.ToLower(string(cn)))]
	if !ok {
		return nil, fmt.Errorf("unknown color name: %s", cn)
	}
	return hexToColor(hex)
}

// ToHexString converts the ColorName to an exact Hex string.
func (cn ColorName) ToHexString() (string, error) {
	hex, ok := colorNameToHex[ColorName(strings.ToLower(string(cn)))]
	if !ok {
		return "", fmt.Errorf("unknown color name: %s", cn)
	}
	return hex, nil
}

// ColorNameToColorName returns the exact ColorName for a given hex string or color.Color.
func ColorNameFromHexString(hex string) (ColorName, bool) {
	hex = strings.ToLower(hex)
	if len(hex) > 0 && hex[0] != '#' {
		hex = "#" + hex
	}
	for name, h := range colorNameToHex {
		if h == hex {
			return name, true
		}
	}
	return "", false
}

func ColorNameFromColor(c color.Color) (ColorName, bool) {
	return ColorNameFromHexString(colorToHex(c))
}

// ColorNameToNearestColorName returns the closest ColorName for a given hex string using Euclidean distance in RGB space.
func ClosestColorNameFromHexString(hex string) (ColorName, error) {
	c, err := hexToColor(hex)
	if err != nil {
		return "", err
	}
	return ClosestColorNameFromColor(c), nil
}

// ClosestColorNameFromColor returns the closest ColorName for a given color.Color using Euclidean distance in RGB space.
func ClosestColorNameFromColor(c color.Color) ColorName {
	exact, ok := ColorNameFromColor(c)
	if ok {
		return exact
	}

	cr, cg, cb, _ := c.RGBA()
	cr, cg, cb = cr>>8, cg>>8, cb>>8

	var closestName ColorName
	minDist := math.MaxFloat64

	for name, hex := range colorNameToHex {
		hc, _ := hexToColor(hex)
		hcr, hcg, hcb, _ := hc.RGBA()
		hcr, hcg, hcb = hcr>>8, hcg>>8, hcb>>8

		dr := float64(cr) - float64(hcr)
		dg := float64(cg) - float64(hcg)
		db := float64(cb) - float64(hcb)
		dist := math.Sqrt(dr*dr + dg*dg + db*db)

		if dist < minDist {
			minDist = dist
			closestName = name
		}
	}
	return closestName
}
