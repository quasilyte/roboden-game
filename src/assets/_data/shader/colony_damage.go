//go:build ignore
// +build ignore

package main

var HP float

func Fragment(pos vec4, texCoord vec2, _ vec4) vec4 {
	c := imageSrc0At(texCoord)
	c2 := imageSrc1At(texCoord)
	if c[3] != 0.0 && c2[3] != 0.0 {
		alpha := c[3]
		c *= clamp(HP+(1.0-c2[3]), 0.0, 1.0)
		c[3] = alpha
	}
	return c
}
