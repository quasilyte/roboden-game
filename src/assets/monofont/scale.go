package monofont

import (
	"fmt"
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Scale takes the bitmap font and returns its scaled version.
// Scaling the font is efficient and doesn't extra memory.
//
// A scaling factor of 1 is a no-op.
// A scaling factor of 2 makes the pixels twice as big.
//
// This function will only work with fonts created by
// this package. Any other font will make it panic.
func Scale(f font.Face, scaling uint) font.Face {
	if scaling == 0 {
		panic("a zero scaling factor is not supported")
	}
	if scaling == 1 {
		return f
	}

	bf, ok := f.(*bitmapFont)
	if !ok {
		panic(fmt.Sprintf("expected a bitmap font, got %T", f))
	}

	return &scaledFont{
		font:  bf,
		scale: int(scaling),
	}
}

type scaledFont struct {
	font  *bitmapFont
	scale int // A positive value, 2 or higher
}

func (sf *scaledFont) Close() error {
	return sf.font.Close()
}

func (s *scaledFont) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, _ image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	var bmask *bitmapImage
	dr, bmask, advance, ok = s.font.glyph(dot, r)
	if !ok {
		return dr, bmask, maskp, advance, false
	}

	d := image.Pt(dot.X.Floor(), dot.Y.Floor())
	dr.Min = dr.Min.Sub(d).Mul(s.scale).Add(d)
	dr.Max = dr.Max.Sub(d).Mul(s.scale).Add(d)
	advance *= fixed.Int26_6(s.scale)
	scaledMask := &scaledImage{
		img:   bmask,
		scale: s.scale,
		bounds: image.Rectangle{
			Min: bmask.bounds.Min.Mul(s.scale * 2),
			Max: bmask.bounds.Max.Mul(s.scale * 2),
		},
	}
	return dr, scaledMask, maskp, advance, true
}

func (s *scaledFont) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	advance, ok = s.font.GlyphAdvance(r)
	if !ok {
		return 0, false
	}
	advance *= fixed.Int26_6(s.scale)
	return advance, true
}

func (s *scaledFont) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	bounds, advance, ok = s.font.GlyphBounds(r)
	if !ok {
		return bounds, advance, false
	}
	bounds.Min.X *= fixed.Int26_6(s.scale)
	bounds.Min.Y *= fixed.Int26_6(s.scale)
	bounds.Max.X *= fixed.Int26_6(s.scale)
	bounds.Max.Y *= fixed.Int26_6(s.scale)
	advance *= fixed.Int26_6(s.scale)
	return bounds, advance, true
}

func (s *scaledFont) Kern(r0, r1 rune) fixed.Int26_6 {
	return s.font.Kern(r0, r1) * fixed.Int26_6(s.scale)
}

func (s *scaledFont) Metrics() font.Metrics {
	m := s.font.Metrics()
	return font.Metrics{
		Height:  m.Height * fixed.Int26_6(s.scale),
		Ascent:  m.Ascent * fixed.Int26_6(s.scale),
		Descent: m.Descent * fixed.Int26_6(s.scale),
	}
}

func euclidianDiv(x, y int) int {
	if x < 0 {
		x -= y - 1
	}
	return x / y
}

type scaledImage struct {
	img    *bitmapImage
	scale  int
	bounds image.Rectangle
}

func (s *scaledImage) ColorModel() color.Model {
	return s.img.ColorModel()
}

func (s *scaledImage) Bounds() image.Rectangle {
	return s.bounds
}

func (s *scaledImage) At(x, y int) color.Color {
	x = euclidianDiv(x, s.scale)
	y = euclidianDiv(y, s.scale)
	return s.img.At(x, y)
}
