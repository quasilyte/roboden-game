package assets

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/bitmapfont/v3"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	BitmapFont1 = bitmapfont.Face
	BitmapFont2 = scaleFont(BitmapFont1, 2)
	BitmapFont3 = scaleFont(BitmapFont1, 3)
)

func euclidianDiv(x, y int) int {
	if x < 0 {
		x -= y - 1
	}
	return x / y
}

type scaledImage struct {
	image image.Image
	scale int
}

func (s *scaledImage) ColorModel() color.Model {
	return s.image.ColorModel()
}

func (s *scaledImage) Bounds() image.Rectangle {
	b := s.image.Bounds()
	b.Min = b.Min.Mul(s.scale)
	b.Max = b.Max.Mul(s.scale)
	return b
}

func (s *scaledImage) At(x, y int) color.Color {
	x = euclidianDiv(x, s.scale)
	y = euclidianDiv(y, s.scale)
	return s.image.At(x, y)
}

func scaleFont(f font.Face, scale int) font.Face {
	if scale == 1 {
		return f
	}
	return &scaledFont{f, scale}
}

type scaledFont struct {
	font  font.Face
	scale int
}

func (s *scaledFont) Close() error {
	return s.font.Close()
}

func (s *scaledFont) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	dr, mask, maskp, advance, ok = s.font.Glyph(dot, r)
	if !ok {
		return
	}
	d := image.Pt(dot.X.Floor(), dot.Y.Floor())
	dr.Min = dr.Min.Sub(d).Mul(s.scale).Add(d)
	dr.Max = dr.Max.Sub(d).Mul(s.scale).Add(d)
	maskp = maskp.Mul(s.scale)
	advance *= fixed.Int26_6(s.scale)
	return dr, &scaledImage{mask, s.scale}, maskp, advance, ok
}

func (s *scaledFont) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	bounds, advance, ok = s.font.GlyphBounds(r)
	if !ok {
		return
	}
	bounds.Min.X *= fixed.Int26_6(s.scale)
	bounds.Min.Y *= fixed.Int26_6(s.scale)
	bounds.Max.X *= fixed.Int26_6(s.scale)
	bounds.Max.Y *= fixed.Int26_6(s.scale)
	advance *= fixed.Int26_6(s.scale)
	return
}

func (s *scaledFont) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	advance, ok = s.font.GlyphAdvance(r)
	if !ok {
		return
	}
	advance *= fixed.Int26_6(s.scale)
	return
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
