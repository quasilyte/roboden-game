package staging

import (
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/viewport"
)

type tutorialHintNode struct {
	rect        *ge.Rect
	label       *ge.Label
	targetLine  *ge.Line
	targetLine2 *ge.Line
	camera      *viewport.Camera

	pos       gmath.Vec
	targetPos ge.Pos
	screenPos bool

	text   string
	width  float64
	height float64
}

func newScreenTutorialHintNode(camera *viewport.Camera, pos, targetPos gmath.Vec, text string) *tutorialHintNode {
	return &tutorialHintNode{
		pos:       pos,
		targetPos: ge.Pos{Offset: targetPos},
		text:      text,
		camera:    camera,
		screenPos: true,
	}
}

func newWorldTutorialHintNode(camera *viewport.Camera, pos gmath.Vec, targetPos ge.Pos, text string) *tutorialHintNode {
	return &tutorialHintNode{
		pos:       pos,
		targetPos: targetPos,
		text:      text,
		camera:    camera,
	}
}

func (hint *tutorialHintNode) Init(scene *ge.Scene) {
	ff := scene.Context().Loader.LoadFont(assets.FontTiny)
	bounds := text.BoundString(ff.Face, hint.text)
	hint.width = float64(bounds.Dx()) + 16
	hint.height = float64(bounds.Dy()) + 20

	hint.rect = ge.NewRect(scene.Context(), hint.width, hint.height)
	hint.rect.OutlineColorScale.SetColor(ge.RGB(0x5e5a5d))
	hint.rect.OutlineWidth = 1
	hint.rect.FillColorScale.SetRGBA(0x13, 0x1a, 0x22, 230)
	hint.rect.Centered = false
	hint.rect.Pos.Offset = hint.pos

	hint.label = scene.NewLabel(assets.FontTiny)
	hint.label.AlignHorizontal = ge.AlignHorizontalCenter
	hint.label.AlignVertical = ge.AlignVerticalCenter
	hint.label.Width = hint.width
	hint.label.Height = hint.height
	hint.label.Pos.Offset = hint.pos
	hint.label.Text = hint.text
	hint.label.ColorScale.SetColor(ge.RGB(0x9dd793))

	if !hint.targetPos.Resolve().IsZero() {
		hint.targetLine = ge.NewLine(ge.Pos{}, ge.Pos{})
		hint.targetLine.Width = 1
		var clr ge.ColorScale
		clr.SetColor(ge.RGB(0x9dd793))
		hint.targetLine.SetColorScale(clr)
		hint.camera.AddGraphicsAbove(hint.targetLine)

		hint.targetLine2 = ge.NewLine(ge.Pos{}, ge.Pos{})
		hint.targetLine2.Width = 1
		hint.targetLine2.SetColorScale(clr)
		hint.camera.AddGraphicsAbove(hint.targetLine2)
	}

	scene.AddGraphicsAbove(hint.rect, 1)
	scene.AddGraphicsAbove(hint.label, 1)
}

func (hint *tutorialHintNode) Update(delta float64) {
	if hint.targetLine != nil {
		beginPos := hint.camera.Offset.Add(hint.pos)
		beginPos.Y++
		var endPos gmath.Vec
		if hint.screenPos {
			endPos = hint.camera.Offset.Add(hint.targetPos.Offset)
		} else {
			endPos = hint.targetPos.Resolve()
		}
		if endPos.X > (hint.camera.Offset.X + hint.pos.X + hint.width/2) {
			beginPos.X += hint.width
		}

		hint.targetLine.BeginPos = ge.Pos{Offset: beginPos}
		hint.targetLine.EndPos = ge.Pos{Offset: endPos}

		beginPos.Y += hint.height - 2
		hint.targetLine2.BeginPos = ge.Pos{Offset: beginPos}
		hint.targetLine2.EndPos = ge.Pos{Offset: endPos}
	}
}

func (hint *tutorialHintNode) IsDisposed() bool {
	return hint.rect.IsDisposed()
}

func (hint *tutorialHintNode) Dispose() {
	hint.rect.Dispose()
	hint.label.Dispose()

	if hint.targetLine != nil {
		hint.targetLine.Dispose()
		hint.targetLine2.Dispose()
	}
}
