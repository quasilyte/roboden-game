//go:build ignore
// +build ignore

package main

//kage:unit pixels

var Amount float

func Fragment(position vec4, pixelCoord vec2, color vec4) vec4 {
	neighbor := Amount * -1
	center := Amount*4 + 1

	clr := imageSrc0UnsafeAt(pixelCoord)

	x := pixelCoord.x
	y := pixelCoord.y
	rgb := (imageSrc0At(vec2(x+0, y+1)).rgb * neighbor) +
		(imageSrc0At(vec2(x-1, y+0)).rgb * neighbor) +
		(imageSrc0At(vec2(x+0, y+1)).rgb * neighbor) +
		(imageSrc0At(vec2(x+1, y+0)).rgb * neighbor) +
		(imageSrc0At(vec2(x+0, y-1)).rgb * neighbor) +
		(clr.rgb * center)

	return vec4(rgb, clr.a)
}
