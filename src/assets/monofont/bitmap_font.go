package monofont

import (
	"fmt"
	"image"
	"unicode"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type bitmapFont struct {
	img         *bitmapImage
	glyphWidth  int
	glyphHeight int

	MinRune      rune
	MaxRune      rune
	RuneToIndex  []uint16
	GlyphBitSize uint
	DotX         fixed.Int26_6
	DotY         fixed.Int26_6
}

func newBitmapFont(img *bitmapImage, dotX, dotY int) *bitmapFont {
	return &bitmapFont{
		img:         img,
		glyphWidth:  int(img.width),
		glyphHeight: int(img.height),
		DotX:        fixed.I(dotX),
		DotY:        fixed.I(dotY),
	}
}

func (f *bitmapFont) Close() error {
	return nil
}

func (f *bitmapFont) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	// maskp remains a zero value as we don't need it.
	dr, mask, advance, ok = f.glyph(dot, r)
	return dr, mask, maskp, advance, ok
}

func (f *bitmapFont) glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask *bitmapImage, advance fixed.Int26_6, ok bool) {
	// First do a quick range check.
	if r > f.MaxRune || r < f.MinRune {
		return dr, mask, advance, false
	}

	// Map rune to its index inside the associated data.
	index, ok := f.getRuneIndex(r)
	if !ok {
		if panicOnUndefined {
			panic(fmt.Sprintf("requesting an undefined rune %v (%q)", r, r))
		}
		return dr, mask, advance, false
	}

	rw := f.glyphWidth
	rh := f.glyphHeight
	dx := (dot.X - f.DotX).Floor()
	dy := (dot.Y - f.DotY).Floor()
	dr = image.Rect(dx, dy, dx+rw, dy+rh)

	offset := index * f.GlyphBitSize
	mask = f.img.WithOffset(offset)
	advance = fixed.I(rw)
	return dr, mask, advance, true
}

func (f *bitmapFont) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	if r > f.MaxRune || r < f.MinRune {
		return 0, false
	}
	return fixed.I(f.glyphWidth), true
}

func (f *bitmapFont) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	if r > f.MaxRune || r < f.MinRune {
		return bounds, advance, false
	}
	bounds = fixed.Rectangle26_6{
		Min: fixed.Point26_6{X: -f.DotX, Y: -f.DotY},
		Max: fixed.Point26_6{
			X: -f.DotX + fixed.I(f.glyphWidth),
			Y: -f.DotY + fixed.I(f.glyphHeight),
		},
	}
	advance = fixed.I(f.glyphWidth)
	return bounds, advance, true
}

func (f *bitmapFont) Kern(r0, r1 rune) fixed.Int26_6 {
	if unicode.Is(unicode.Mn, r1) {
		return -fixed.I(f.glyphWidth)
	}
	return 0

}

func (f *bitmapFont) Metrics() font.Metrics {
	return font.Metrics{
		Height:  fixed.I(f.glyphHeight),
		Ascent:  f.DotY,
		Descent: fixed.I(f.glyphHeight) - f.DotY,
	}
}

func (f *bitmapFont) getRuneIndex(r rune) (uint, bool) {
	u := uint(r)
	slice := f.RuneToIndex
	if u < uint(len(slice)) {
		i := slice[u]
		if i > 0 {
			return uint(i - 1), true
		}
	}
	return 0, false
}

// func (f *Face) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
// 	if r >= 0x10000 {
// 		return
// 	}

// 	rw := f.runeWidth(r)
// 	dx := (dot.X - f.dotX).Floor()
// 	dy := (dot.Y - f.dotY).Floor()
// 	dr = image.Rect(dx, dy, dx+rw, dy+f.charHeight())

// 	mx := (int(r) % charXNum) * f.charFullWidth()
// 	my := (int(r) / charXNum) * f.charHeight()
// 	mask = f.image.SubImage(image.Rect(mx, my, mx+rw, my+f.charHeight()))
// 	maskp = image.Pt(mx, my)
// 	advance = fixed.I(f.runeWidth(r))
// 	ok = true
// 	return
// }
