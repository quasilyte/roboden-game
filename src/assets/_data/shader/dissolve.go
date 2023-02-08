//go:build ignore
// +build ignore

package main

var Time float

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	limit := abs(2*fract(Time/3) - 1)
	level := imageSrc1UnsafeAt(texCoord).x

	if limit-0.1 < level && level < limit && imageSrc0UnsafeAt(texCoord).a != 0.0 {
		v := imageSrc0UnsafeAt(texCoord)
		return v * 0.5
	}

	return step(limit, level) * imageSrc0UnsafeAt(texCoord)
}
