package gameui

import (
	"image/color"

	"github.com/quasilyte/ge"
)

type ProgressBar struct {
	pos ge.Pos

	width  float64
	height float64

	Visible bool

	bgColor color.RGBA
	fgColor color.RGBA

	bg       *ge.Rect
	progress *ge.Rect
}

func NewProgressBar(pos ge.Pos, width, height float64, bg, fg color.RGBA) *ProgressBar {
	return &ProgressBar{
		pos:     pos,
		width:   width,
		height:  height,
		bgColor: bg,
		fgColor: fg,
		Visible: true,
	}
}

func (b *ProgressBar) Init(scene *ge.Scene) {
	b.bg = ge.NewRect(scene.Context(), b.width, b.height)
	b.bg.Centered = false
	b.bg.Pos = b.pos
	b.bg.FillColorScale.SetColor(b.bgColor)
	scene.AddGraphics(b.bg)

	b.progress = ge.NewRect(scene.Context(), b.width-8, b.height-8)
	b.progress.Centered = false
	b.progress.Pos = b.pos.WithOffset(4, 4)
	b.progress.FillColorScale.SetColor(b.fgColor)
	scene.AddGraphics(b.progress)
}

func (b *ProgressBar) IsDisposed() bool { return b.bg.IsDisposed() }

func (b *ProgressBar) Dispose() {
	b.bg.Dispose()
	b.progress.Dispose()
}

func (b *ProgressBar) Update(delta float64) {
	b.bg.Visible = b.Visible
	b.progress.Visible = b.Visible
}

func (b *ProgressBar) SetValue(value float64) {
	fullWidth := b.width + 8
	b.progress.Width = fullWidth * value
}
