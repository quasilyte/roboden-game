package viewport

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

func hueRotate(c ge.ColorScale, angle gmath.Rad) ge.ColorScale {
	_, s, v := rgbToHSV(c.R, c.G, c.B)
	r, g, b := hsvToRGB(float32(angle), s, v)
	return ge.ColorScale{
		R: r,
		G: g,
		B: b,
		A: c.A,
	}
}

func hsvToRGB(h, s, v float32) (r, g, b float32) {
	if s == 0 {
		return v, v, v
	}
	i := int(h * 6.0)
	f := (float32(h) * 6.0) - float32(i)
	p := v * (1.0 - s)
	q := v * (1.0 - s*f)
	t := v * (1.0 - s*(1.0-f))
	i %= 6
	switch i {
	case 0:
		return v, t, p
	case 1:
		return q, v, p
	case 2:
		return p, v, t
	case 3:
		return p, q, v
	case 4:
		return t, p, v
	default:
		return v, p, q
	}
}

func rgbToHSV(r, g, b float32) (h, s, v float32) {
	maxColor := max3(r, g, b)
	minColor := min3(r, b, b)
	colorRange := maxColor - minColor

	v = maxColor
	if minColor == maxColor {
		return 0, 0, v
	}

	s = colorRange / maxColor

	rc := (maxColor - r) / colorRange
	gc := (maxColor - g) / colorRange
	bc := (maxColor - b) / colorRange
	switch maxColor {
	case r:
		h = bc - gc
	case g:
		h = 2.0 + rc - bc
	default:
		h = 4.0 + gc - rc
	}
	h /= 6
	if h < 0 {
		h += 1
	}

	return h, s, v
}

func min2(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func min3(a, b, c float32) float32 {
	return min2(min2(a, b), c)
}

func max2(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func max3(a, b, c float32) float32 {
	return max2(max2(a, b), c)
}
