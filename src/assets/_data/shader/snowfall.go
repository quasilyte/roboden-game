package main

//kage:unit pixels

var Time float
var OffsetY float
var OffsetX float
var ResolutionWidth float
var ResolutionHeight float

func snow(uv vec2, scale, t float) float {
	w := smoothstep(1.0, 0.0, -uv.y*(scale/20.0))
	if w < 0.1 {
		return 0.0
	}

	uv += t / scale
	uv.y += (t * 2.0 / scale)
	uv.x += sin(uv.y+t*0.5) / scale

	uv *= scale
	s := floor(uv)
	f := fract(uv)
	k := 3.0
	p := 0.3 + 0.35*sin(11.0*fract(sin((s+scale)*mat2(vec2(7, 3), vec2(6, 5)))*5.0)) - f
	d := length(p)
	k = min(d, k)
	k = smoothstep(0.0, k, sin(f.x+f.y)*0.01)
	return k * w
}

func Fragment(dst vec4, src vec2, color vec4) vec4 {
	t := 0.25 * -Time

	resolution := vec2(ResolutionWidth, ResolutionHeight)

	uv := (dst.xy*2.0 - resolution.xy) / min(resolution.x, resolution.y)
	uv.x += OffsetX
	uv.y *= 0.8
	uv.y += OffsetY

	c := vec3(0)
	c += snow(uv, 10.0, t) * 0.9
	c += snow(uv, 8.0, t)

	return vec4(c, 0)
}
