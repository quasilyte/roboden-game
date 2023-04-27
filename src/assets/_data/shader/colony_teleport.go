//go:build ignore
// +build ignore

package main

var Time float

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

func shaderRand(pixCoord vec2) (seedMod, randValue int) {
	pixSize := imageSrcTextureSize()
	pixelOffset := int(pixCoord.x) + int(pixCoord.y*pixSize.x)
	seedMod = pixelOffset % int(Time)
	pixelOffset += seedMod
	return seedMod, pixelOffset + int(Time)
}

func applyVideoDegradation(x float, c vec4) vec4 {
	if c.a != 0.0 {
		if int(x+Time)%4 != int(0) {
			return c * vec4(0.75, 0.65, 0.95, 0.7)
		}
	}
	return c
}

func Fragment(pos vec4, texCoord vec2, _ vec4) vec4 {
	c := imageSrc0At(texCoord)

	actualPixCoord := tex2pixCoord(texCoord)
	if c.a != 0.0 {
		seedMod, h := shaderRand(actualPixCoord)
		dist := 1.0
		if seedMod == int(0) {
			dist = 2.0
		}
		p := applyPixPick(actualPixCoord, dist, 5, h)
		return applyVideoDegradation(pos.x, imageSrc0At(pix2texCoord(p)))
	}

	return c
}
