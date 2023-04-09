//go:build ignore
// +build ignore

package main

var Time float

func Fragment(_ vec4, texCoord vec2, _ vec4) vec4 {
	pixSize := imageSrcTextureSize()
	originTexPos, srcRegion := imageSrcRegionOnTexture()
	width := pixSize.x * srcRegion.x
	actualTexPos := texCoord - originTexPos
	actualPixPos := actualTexPos * pixSize
	actualPixPos.x = mod(actualPixPos.x-(100*Time), width)
	return imageSrc0At(actualPixPos/pixSize + originTexPos)
}
