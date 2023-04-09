//go:build ignore
// +build ignore

package main

var Time float

func Fragment(_ vec4, texCoord vec2, _ vec4) vec4 {
	origin, size := imageSrcRegionOnTexture()
	x := origin.x + mod(texCoord.x-origin.x-(size.x*Time), size.x)
	return imageSrc0At(vec2(x, texCoord.y))
}
