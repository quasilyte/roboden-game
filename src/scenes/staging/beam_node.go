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
	repairBeamColor    = ge.RGB(0x6ac037)
	rechargerBeamColor = ge.RGB(0x66ced6)
	railgunBeamColor   = ge.RGB(0xbd1844)
	stunnerBeamColor   = ge.RGB(0x7d21cd)
	destroyerBeamColor = ge.RGB(0xf58f54)
	evoBeamColor       = ge.RGB(0xa641c2)
)

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
