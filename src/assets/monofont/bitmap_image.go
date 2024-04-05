package monofont

import (
	"image"
	"image/color"
)

var (
	colorZero = color.Alpha{0}
	colorOne  = color.Alpha{0xff}
)

type bitmapImage struct {
	data   []byte
	width  uint
	height uint
	offset uint
	bounds image.Rectangle
}

func newBitmapImage(data []byte, w, h int) *bitmapImage {
	// data is expected to be uncompressed.
	return &bitmapImage{
		width:  uint(w),
		height: uint(h),
		data:   data,
		bounds: image.Rect(0, 0, w, h),
	}
}

func (img *bitmapImage) WithOffset(offset uint) *bitmapImage {
	return &bitmapImage{
		data:   img.data,
		width:  img.width,
		offset: offset,
		bounds: img.bounds,
	}
}

func (img *bitmapImage) ColorModel() color.Model {
	return color.AlphaModel
}

func (img *bitmapImage) Bounds() image.Rectangle {
	return img.bounds
}

func (img *bitmapImage) At(x, y int) color.Color {
	i := (uint(y) * img.width) + uint(x) + img.offset
	byteIndex := i / 8
	byteShift := i % 8
	data := img.data
	if byteIndex < uint(len(data)) {
		b := data[byteIndex]
		v := b >> byte(byteShift) & 0b1
		if v == 1 {
			return colorOne
		}
	}
	return colorZero
}
