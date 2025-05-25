package staging

import (
	"image/color"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type healthBarComponent struct {
	world *worldState
	scene *ge.Scene

	pos    *gmath.Vec
	offset gmath.Vec

	maxHealth float64
	health    *float64

	width  float64
	height float64

	bgRect     *ge.Rect
	healthRect *ge.Rect

	disposed bool
}

func newHealthBarComponent(world *worldState, scene *ge.Scene, pos *gmath.Vec, offset gmath.Vec, health *float64, maxHealth float64, width float64) *healthBarComponent {
	return &healthBarComponent{
		world:     world,
		scene:     scene,
		pos:       pos,
		offset:    offset,
		health:    health,
		maxHealth: maxHealth,
		width:     width,
		height:    4,
	}
}

func (h *healthBarComponent) Init() {
	h.bgRect = ge.NewRect(h.scene.Context(), h.width, h.height)
	h.bgRect.FillColorScale.SetColor(color.RGBA{R: 40, G: 40, B: 40, A: 180})
	h.bgRect.OutlineWidth = 0
	h.bgRect.Centered = false
	h.bgRect.Visible = h.world.showHealthbars
	h.world.stage.AddGraphicsAbove(h.bgRect)

	h.healthRect = ge.NewRect(h.scene.Context(), h.width-2, h.height-2)
	h.healthRect.OutlineWidth = 0
	h.healthRect.Centered = false
	h.healthRect.Visible = h.world.showHealthbars
	h.world.stage.AddGraphicsAbove(h.healthRect)

	h.Update()
}

func (h *healthBarComponent) Update() {
	if h.disposed {
		return
	}

	visible := h.world.showHealthbars
	h.bgRect.Visible = visible
	h.healthRect.Visible = visible

	if !visible {
		return
	}

	healthPercent := *h.health / h.maxHealth
	if healthPercent < 0 {
		healthPercent = 0
	}

	barPos := h.pos.Add(h.offset).Sub(gmath.Vec{X: h.width / 2})
	h.bgRect.Pos.Offset = barPos
	h.healthRect.Pos.Offset = barPos.Add(gmath.Vec{X: 1, Y: 1})

	barWidth := (h.width - 2) * healthPercent
	if barWidth <= 0 {
		h.healthRect.Visible = false
		return
	}

	h.healthRect.Width = barWidth

	var barColor color.RGBA
	switch {
	case healthPercent > 0.6:
		barColor = color.RGBA{R: 80, G: 200, B: 80, A: 220}
	case healthPercent > 0.3:
		barColor = color.RGBA{R: 200, G: 200, B: 80, A: 220}
	default:
		barColor = color.RGBA{R: 200, G: 80, B: 80, A: 220}
	}
	h.healthRect.FillColorScale.SetColor(barColor)
}

func (h *healthBarComponent) Dispose() {
	if h.disposed {
		return
	}
	h.disposed = true
	if h.bgRect != nil {
		h.bgRect.Dispose()
	}
	if h.healthRect != nil {
		h.healthRect.Dispose()
	}
}
