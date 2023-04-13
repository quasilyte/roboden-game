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
	slidePixPos := vec2(mod(actualPixPos.x-(100*Time), width), actualPixPos.y)
	c := imageSrc0At(slidePixPos/pixSize + originTexPos)
	if actualPixPos.x <= 10.0 {
		// width=110 x=2 a=0.2
		// width=110 x=8 a=0.8
		c *= actualPixPos.x * 0.1
	} else if actualPixPos.x >= (width - 10.0) {
		// width=110 x=102 delta=8 a=0.8
		// width=110 x=108 delta=2 a=0.2
		c *= (width - actualPixPos.x) * 0.1
	}
	return c
}
