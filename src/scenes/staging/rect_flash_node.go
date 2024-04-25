package staging

import (
	"image/color"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/viewport"
)

type rectFlashNode struct {
	clr    ge.ColorScale
	camera *viewport.Camera
	geom   gmath.Rect
	rect   *ge.Rect
}

func newRectFlashNode(cam *viewport.Camera, clr color.RGBA, geom gmath.Rect) *rectFlashNode {
	var colorScale ge.ColorScale
	colorScale.SetColor(clr)
	return &rectFlashNode{
		clr:    colorScale,
		camera: cam,
		geom:   geom,
	}
}

func (n *rectFlashNode) Init(scene *ge.Scene) {
	n.rect = ge.NewRect(scene.Context(), n.geom.Width(), n.geom.Height())
	n.rect.OutlineColorScale = n.clr
	n.rect.OutlineWidth = 2
	n.rect.FillColorScale = ge.ColorScale{}
	n.rect.Pos.Offset = rectCenter(n.geom)
	n.camera.UI.AddGraphicsAbove(n.rect)
}

func (n *rectFlashNode) IsDisposed() bool {
	return n.rect.IsDisposed()
}

func (n *rectFlashNode) Dispose() {
	n.rect.Dispose()
}

func (n *rectFlashNode) Update(delta float64) {
	n.rect.OutlineColorScale.A -= float32(delta * 0.5)
	if n.rect.OutlineColorScale.A < 0.1 {
		n.Dispose()
		return
	}

	n.rect.Width += 10 * delta
	n.rect.Height += 16 * delta
}
