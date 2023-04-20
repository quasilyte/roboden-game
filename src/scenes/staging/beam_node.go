package staging

import (
	"image/color"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
)

type beamNode struct {
	from  ge.Pos
	to    ge.Pos
	world *worldState

	color      color.RGBA
	width      float64
	line       *ge.Line
	opaqueTime float64

	texture        *ge.Texture
	beamSlideSpeed float64
	shaderTime     float64
	texLine        *ge.TextureLine
}

var (
	dominatorBeamColorCenter = ge.RGB(0x7a51f2)
	dominatorBeamColorRear   = ge.RGB(0x5433c3)
	builderBeamColor         = color.RGBA{R: 0xae, G: 0x4c, B: 0x78, A: 150}
	destroyerBeamColor       = ge.RGB(0xf58f54)
	courierResourceBeamColor = ge.RGB(0xd2e352)
	prismBeamColor1          = ge.RGB(0x529eb8)
	prismBeamColor2          = ge.RGB(0x61bad8)
	prismBeamColor3          = ge.RGB(0x7bdbfc)
	prismBeamColor4          = ge.RGB(0xccf2ff)
	evoBeamColor             = ge.RGB(0xa641c2)
)

var prismBeamColors = []color.RGBA{
	prismBeamColor1,
	prismBeamColor2,
	prismBeamColor3,
	prismBeamColor4,
}

func newBeamNode(world *worldState, from, to ge.Pos, clr color.RGBA) *beamNode {
	return &beamNode{
		world: world,
		from:  from,
		to:    to,
		width: 1,
		color: clr,
	}
}

func newTextureBeamNode(world *worldState, from, to ge.Pos, texture *ge.Texture, beamSlideSpeed float64, opaqueTime float64) *beamNode {
	return &beamNode{
		world:          world,
		from:           from,
		to:             to,
		texture:        texture,
		beamSlideSpeed: beamSlideSpeed,
		opaqueTime:     opaqueTime,
	}
}

func (b *beamNode) Init(scene *ge.Scene) {
	if b.texture == nil {
		b.line = ge.NewLine(b.from, b.to)
		var c ge.ColorScale
		c.SetColor(b.color)
		b.line.SetColorScale(c)
		b.line.Width = b.width
		b.world.camera.AddGraphicsAbove(b.line)
	} else {
		b.texLine = ge.NewTextureLine(scene.Context(), b.from, b.to)
		b.texLine.SetTexture(b.texture)
		if b.beamSlideSpeed != 0 && b.world.graphicsSettings.AllShadersEnabled {
			b.texLine.Shader = scene.NewShader(assets.ShaderSlideX)
			b.texLine.Shader.SetFloatValue("Time", 0)
		}
		b.world.camera.AddGraphicsAbove(b.texLine)
	}
}

func (b *beamNode) IsDisposed() bool {
	if b.texture == nil {
		return b.line.IsDisposed()
	}
	return b.texLine.IsDisposed()
}

func (b *beamNode) Update(delta float64) {
	if b.texLine != nil && !b.texLine.Shader.IsNil() {
		b.shaderTime += delta * b.beamSlideSpeed
		b.texLine.Shader.SetFloatValue("Time", b.shaderTime)
	}
	if b.opaqueTime > 0 {
		b.opaqueTime -= delta
		return
	}

	if b.texture == nil {
		if b.line.GetAlpha() < 0.1 {
			b.line.Dispose()
			return
		}
		b.line.SetAlpha(b.line.GetAlpha() - float32(delta*4))
		return
	}

	if b.texLine.GetAlpha() < 0.1 {
		b.texLine.Dispose()
		return
	}
	b.texLine.SetAlpha(b.texLine.GetAlpha() - float32(delta*4))
}
