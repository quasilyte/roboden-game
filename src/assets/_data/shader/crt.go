//go:build ignore

// This is a port of a public domain crt-lottes shader written by Timothy Lottes.
// Rewritten to Kage by Elias Daler. The license is still public domain.
// The original source code can be found here: https://github.com/libretro/glsl-shaders/blob/master/crt/shaders/crt-lottes.glsl
//
// Changes:
//
// 1. DO_BLOOM is assumed. If you don't want bloom, set BloomAmount to 0
// 2. Accureate linear gamma is used because it looks better
// 3. Clamp fix is removed - there's no need for it, I think

package main

var TextureSize vec2 // input screen size (e.g. 256x244 for SNES)
var ScreenSize vec2  // output screen size (e.g. when rendering at 4x resolution it is 1024x976)

// Default settings:
// const HardScan = -8.0
// const HardPix = -3.0
// const WarpX = 0.031
// const WarpY = 0.041
// const MaskDark = 0.5
// const MaskLight = 1.5
// const ShadowMask = 3.0
// const BrightBoost = 1.0
// const HardBloomPix = -1.5
// const HardBloomScan = -2.0
// const BloomAmount = 0.05
// const Shape = 2.0

const HardScan = -2.0
const HardPix = -3.0
const WarpX = 0.031
const WarpY = 0.041
const MaskDark = 0.8
const MaskLight = 1.2
const ShadowMask = 3.0
const BrightBoost = 1.05
const HardBloomPix = -1.8
const HardBloomScan = -2.0
const BloomAmount = 0.05
const Shape = 2.5

func ToLinear1(c float) float {
	if c <= 0.04045 {
		return c / 12.92
	}
	return pow((c+0.055)/1.055, 2.4)
}

func ToLinear(c vec3) vec3 {
	return vec3(ToLinear1(c.r), ToLinear1(c.g), ToLinear1(c.b))
}

// Linear to sRGB.
// Assuming using sRGB typed textures this should not be needed.
func ToSrgb1(c float) float {
	if c < 0.0031308 {
		return c * 12.92
	}
	return 1.055*pow(c, 0.41666) - 0.055
}

func ToSrgb(c vec3) vec3 {
	return vec3(ToSrgb1(c.r), ToSrgb1(c.g), ToSrgb1(c.b))
}

// Nearest emulated sample given floating point position and texel offset.
// Also zero's off screen.
func Fetch(pos vec2, off vec2) vec3 {
	pos = (floor(pos*TextureSize.xy+off) + vec2(0.5, 0.5)) / TextureSize.xy
	origin, size := imageSrcRegionOnTexture()
	pos = pos*size + origin // IMPORTANT: go back to atlas coordinates from texture coordinates which OpenGL uses
	return ToLinear(BrightBoost * imageSrc0At(pos.xy).rgb)
}

// Distance in emulated pixels to nearest texel.
func Dist(pos vec2) vec2 {
	pos = pos * TextureSize.xy
	return -((pos - floor(pos)) - vec2(0.5))
}

// 1D Gaussian.
func Gaus(pos float, scale float) float {
	return exp2(scale * pow(abs(pos), Shape))
}

// 3-tap Gaussian filter along horz line.
func Horz3(pos vec2, off float) vec3 {
	b := Fetch(pos, vec2(-1.0, off))
	c := Fetch(pos, vec2(0.0, off))
	d := Fetch(pos, vec2(1.0, off))
	dst := Dist(pos).x

	// Convert distance to weight.
	scale := HardPix
	wb := Gaus(dst-1.0, scale)
	wc := Gaus(dst+0.0, scale)
	wd := Gaus(dst+1.0, scale)

	// Return filtered sample.
	return (b*wb + c*wc + d*wd) / (wb + wc + wd)
}

// 5-tap Gaussian filter along horz line.
func Horz5(pos vec2, off float) vec3 {
	a := Fetch(pos, vec2(-2.0, off))
	b := Fetch(pos, vec2(-1.0, off))
	c := Fetch(pos, vec2(0.0, off))
	d := Fetch(pos, vec2(1.0, off))
	e := Fetch(pos, vec2(2.0, off))

	dst := Dist(pos).x
	// Convert distance to weight.
	scale := HardPix
	wa := Gaus(dst-2.0, scale)
	wb := Gaus(dst-1.0, scale)
	wc := Gaus(dst+0.0, scale)
	wd := Gaus(dst+1.0, scale)
	we := Gaus(dst+2.0, scale)

	// Return filtered sample.
	return (a*wa + b*wb + c*wc + d*wd + e*we) / (wa + wb + wc + wd + we)
}

// 7-tap Gaussian filter along horz line.
func Horz7(pos vec2, off float) vec3 {
	a := Fetch(pos, vec2(-3.0, off))
	b := Fetch(pos, vec2(-2.0, off))
	c := Fetch(pos, vec2(-1.0, off))
	d := Fetch(pos, vec2(0.0, off))
	e := Fetch(pos, vec2(1.0, off))
	f := Fetch(pos, vec2(2.0, off))
	g := Fetch(pos, vec2(3.0, off))

	dst := Dist(pos).x
	// Convert distance to weight.
	scale := HardBloomPix
	wa := Gaus(dst-3.0, scale)
	wb := Gaus(dst-2.0, scale)
	wc := Gaus(dst-1.0, scale)
	wd := Gaus(dst+0.0, scale)
	we := Gaus(dst+1.0, scale)
	wf := Gaus(dst+2.0, scale)
	wg := Gaus(dst+3.0, scale)

	// Return filtered sample.
	return (a*wa + b*wb + c*wc + d*wd + e*we + f*wf + g*wg) / (wa + wb + wc + wd + we + wf + wg)
}

// Return scanline weight.
func Scan(pos vec2, off float) float {
	dst := Dist(pos).y
	return Gaus(dst+off, HardScan)
}

// Return scanline weight for Bloom.
func BloomScan(pos vec2, off float) float {
	dst := Dist(pos).y

	return Gaus(dst+off, HardBloomScan)
}

// Allow nearest three lines to effect pixel.
func Tri(pos vec2) vec3 {
	a := Horz3(pos, -1.0)
	b := Horz5(pos, 0.0)
	c := Horz3(pos, 1.0)

	wa := Scan(pos, -1.0)
	wb := Scan(pos, 0.0)
	wc := Scan(pos, 1.0)

	return a*wa + b*wb + c*wc
}

// Small Bloom.
func Bloom(pos vec2) vec3 {
	a := Horz5(pos, -2.0)
	b := Horz7(pos, -1.0)
	c := Horz7(pos, 0.0)
	d := Horz7(pos, 1.0)
	e := Horz5(pos, 2.0)

	wa := BloomScan(pos, -2.0)
	wb := BloomScan(pos, -1.0)
	wc := BloomScan(pos, 0.0)
	wd := BloomScan(pos, 1.0)
	we := BloomScan(pos, 2.0)

	return a*wa + b*wb + c*wc + d*wd + e*we
}

// Distortion of scanlines, and end of screen alpha.
func Warp(pos vec2) vec2 {
	pos = pos*2.0 - 1.0
	pos *= vec2(1.0+(pos.y*pos.y)*WarpX, 1.0+(pos.x*pos.x)*WarpY)

	return pos*0.5 + 0.5
}

// Shadow mask.
func Mask(pos vec2) vec3 {
	mask := vec3(MaskDark, MaskDark, MaskDark)

	if ShadowMask == 1.0 {
		// Very compressed TV style shadow mask.
		line := MaskLight
		odd := 0.0

		if fract(pos.x*0.166666666) < 0.5 {
			odd = 1.0
		}
		if fract((pos.y+odd)*0.5) < 0.5 {
			line = MaskDark
		}

		pos.x = fract(pos.x * 0.333333333)

		if pos.x < 0.333 {
			mask.r = MaskLight
		} else if pos.x < 0.666 {
			mask.g = MaskLight
		} else {
			mask.b = MaskLight
		}

		mask *= line
	} else if ShadowMask == 2.0 {
		// Aperture-grille.
		pos.x = fract(pos.x * 0.333333333)

		if pos.x < 0.333 {
			mask.r = MaskLight
		} else if pos.x < 0.666 {
			mask.g = MaskLight
		} else {
			mask.b = MaskLight
		}
	} else if ShadowMask == 3.0 {
		// Stretched VGA style shadow mask (same as prior shaders).
		pos.x += pos.y * 3.0
		pos.x = fract(pos.x * 0.166666666)

		if pos.x < 0.333 {
			mask.r = MaskLight
		} else if pos.x < 0.666 {
			mask.g = MaskLight
		} else {
			mask.b = MaskLight
		}
	} else if ShadowMask == 4.0 {
		// VGA style shadow mask.
		pos.xy = floor(pos.xy * vec2(1.0, 0.5))
		pos.x += pos.y * 3.0
		pos.x = fract(pos.x * 0.166666666)

		if pos.x < 0.333 {
			mask.r = MaskLight
		} else if pos.x < 0.666 {
			mask.g = MaskLight
		} else {
			mask.b = MaskLight
		}
	}
	return mask
}

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	// Adjust the texture position to [0, 1].
	pos := texCoord
	origin, size := imageSrcRegionOnTexture()
	pos -= origin
	pos /= size

	pos = Warp(pos)
	outColor := Tri(pos)

	//Add Bloom
	outColor.rgb += Bloom(pos) * BloomAmount

	fragCoord := position.xy
	fragCoord.y = ScreenSize.y - fragCoord.y // in OpenGL, Y is pointing up

	if ShadowMask > 0.0 {
		outColor.rgb *= Mask(fragCoord * 1.000001)
	}

	return vec4(ToSrgb(outColor.rgb), 1.0)
	// return vec4(imageSrc0At(texCoord).rgb, 1.0)
}
