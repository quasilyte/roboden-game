package menus

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
)

func reverseStrings(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

func createSubImage(img resource.Image) *ebiten.Image {
	if int(img.DefaultFrameWidth) == img.Data.Bounds().Dx() && int(img.DefaultFrameHeight) == img.Data.Bounds().Dy() {
		return img.Data
	}
	if img.DefaultFrameWidth == 0 && img.DefaultFrameHeight == 0 {
		return img.Data
	}
	width := int(img.DefaultFrameWidth)
	height := int(img.DefaultFrameHeight)
	if width == 0 {
		width = img.Data.Bounds().Dx()
	}
	if height == 0 {
		height = img.Data.Bounds().Dy()
	}
	return img.Data.SubImage(image.Rectangle{
		Max: image.Point{
			X: width,
			Y: height,
		},
	}).(*ebiten.Image)
}
