// A Kage port of https://www.shadertoy.com/view/Ms23DR by https://github.com/Zyko0
//
// The original license comment is:
//	Loosely based on postprocessing shader by inigo quilez,
//	License Creative Commons Attribution-NonCommercial-ShareAlike 3.0 Unported License.

package main

//kage:unit pixels

func curve(uv vec2) vec2 {
	uv = (uv - 0.5) * 2
	uv *= 1.1
	uv.x *= (1 + pow((abs(uv.y)/8), 2))
	uv.y *= (1 + pow((abs(uv.x)/6), 2))
	uv = uv*0.5 + 0.5
	uv = uv*0.92 + 0.04

	return uv
}

func Fragment(dst vec4, src vec2, color vec4) vec4 {
	origin, size := imageSrcRegionOnTexture()
	q := (src - origin) / size
	uv := q
	uv = curve(uv)

	var col vec3
	col.r = imageSrc0At(vec2(uv.x+0.0001, uv.y+0.0001)*size+origin).x + 0.0025
	col.g = imageSrc0At(vec2(uv.x+0.0000, uv.y-0.0002)*size+origin).y + 0.0025
	col.b = imageSrc0At(vec2(uv.x-0.0002, uv.y+0.0000)*size+origin).z + 0.0025
	col.r += 0.04 * imageSrc0At((0.75*vec2(0.025, -0.027)+vec2(uv.x+0.001, uv.y+0.001))*size+origin).x
	col.g += 0.025 * imageSrc0At((0.75*vec2(-0.022, -0.02)+vec2(uv.x+0.000, uv.y-0.002))*size+origin).y
	col.b += 0.04 * imageSrc0At((0.75*vec2(-0.02, -0.018)+vec2(uv.x-0.002, uv.y+0.000))*size+origin).z

	col = clamp(col*0.6+0.4*col*col, 0, 1)

	vig := (40.0 * uv.x * uv.y * (1 - uv.x) * (1 - uv.y))
	col *= vec3(pow(vig, 0.3))
	col *= vec3(0.95, 1.05, 0.95)
	col *= 2.4

	scans := clamp(0.35+0.35*sin(uv.y*size.y*1.5), 0, 1)
	s := pow(scans, 3.7)
	col *= vec3(0.45 + 0.1*s)

	if uv.x < 0.0 || uv.x > 1.0 || uv.y < 0 || uv.y > 1 {
		col *= 0
	}

	col *= (1.0 - 0.25*vec3(clamp((mod(src.x, 2)-1)*2, 0, 1)))

	return vec4(col, 1) * 1.05
}
