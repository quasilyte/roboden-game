package staging

import (
	"image/color"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/viewport"
)

type beamNode struct {
	from   ge.Pos
	to     ge.Pos
	color  color.RGBA
	width  float64
	camera *viewport.Camera

	line *ge.Line
}

var (
	repairBeamColor          = ge.RGB(0x6ac037)
	rechargerBeamColor       = ge.RGB(0x66ced6)
	railgunBeamColor         = ge.RGB(0xbd1844)
	dominatorBeamColorCenter = ge.RGB(0x7a51f2)
	dominatorBeamColorRear   = ge.RGB(0x5433c3)
	builderBeamColor         = color.RGBA{R: 0xae, G: 0x4c, B: 0x78, A: 150}
	stunnerBeamColor         = ge.RGB(0x7d21cd)
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

func newBeamNode(camera *viewport.Camera, from, to ge.Pos, c color.RGBA) *beamNode {
	return &beamNode{
		camera: camera,
		from:   from,
		to:     to,
		color:  c,
		width:  1,
	}
}

func (b *beamNode) Init(scene *ge.Scene) {
	b.line = ge.NewLine(b.from, b.to)
	var c ge.ColorScale
	c.SetColor(b.color)
	b.line.SetColorScale(c)
	b.line.Width = b.width
	b.camera.AddGraphicsAbove(b.line)
}

func (b *beamNode) IsDisposed() bool { return b.line.IsDisposed() }

func (b *beamNode) Update(delta float64) {
	if b.line.GetAlpha() < 0.1 {
		b.line.Dispose()
		return
	}
	b.line.SetAlpha(b.line.GetAlpha() - float32(delta*4))
}
