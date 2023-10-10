//go:build ignore
// +build ignore

package main

var HueAngle float

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	clr := imageSrc0UnsafeAt(texCoord)
	clr = vec4(hueRotate(clr.rgb, HueAngle), clr.a)
	return clr
}

func hueRotate(color vec3, hueAdjust float) vec3 {
	hueAdjust = -hueAdjust

	kRGBToYPrime := vec3(0.299, 0.587, 0.114)
	kRGBToI := vec3(0.596, -0.275, -0.321)
	kRGBToQ := vec3(0.212, -0.523, 0.311)

	kYIQToR := vec3(1.0, 0.956, 0.621)
	kYIQToG := vec3(1.0, -0.272, -0.647)
	kYIQToB := vec3(1.0, -1.107, 1.704)

	YPrime := dot(color, kRGBToYPrime)
	I := dot(color, kRGBToI)
	Q := dot(color, kRGBToQ)
	hue := atan2(Q, I)
	chroma := sqrt(I*I + Q*Q)

	hue += hueAdjust

	Q = chroma * sin(hue)
	I = chroma * cos(hue)

	yIQ := vec3(YPrime, I, Q)

	return vec3(dot(yIQ, kYIQToR), dot(yIQ, kYIQToG), dot(yIQ, kYIQToB))
}
