//go:build ignore
// +build ignore

package main

var Time float
var Seed float

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	clr := imageSrc0UnsafeAt(texCoord)
	if clr.r <= 0.3 {
		return clr
	}

	actualPixCoord := tex2pixCoord(texCoord)
	h := shaderRand(actualPixCoord)
	p := applyPixPick(actualPixCoord, 1.0, 12, h)
	return imageSrc0At(pix2texCoord(p))
}

func shaderRand(pixCoord vec2) int {
	pixSize := imageSrcTextureSize()
	pixelOffset := int(pixCoord.x) + int(pixCoord.y*pixSize.x)
	seedMod := pixelOffset % int(Time)
	pixelOffset += seedMod
	result := pixelOffset + int(Time) + int(Seed)
	result += int(sin(pixCoord.x / 10))
	result += int(sin(pixCoord.y / 20))
	return result
}

func tex2pixCoord(texCoord vec2) vec2 {
	pixSize := imageSrcTextureSize()
	originTexCoord, _ := imageSrcRegionOnTexture()
	actualTexCoord := texCoord - originTexCoord
	actualPixCoord := actualTexCoord * pixSize
	return actualPixCoord
}

func pix2texCoord(actualPixCoord vec2) vec2 {
	pixSize := imageSrcTextureSize()
	actualTexCoord := actualPixCoord / pixSize
	originTexCoord, _ := imageSrcRegionOnTexture()
	texCoord := actualTexCoord + originTexCoord
	return texCoord
}

func applyPixPick(pixCoord vec2, dist float, m, hash int) vec2 {
	dir := hash % m
	if dir == int(0) {
		pixCoord.x += dist
	} else if dir == int(1) {
		pixCoord.x -= dist
	} else if dir == int(2) {
		pixCoord.y += dist
	} else if dir == int(3) {
		pixCoord.y -= dist
	}
	return pixCoord
}
