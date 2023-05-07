package staging

import (
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/viewport"
)

type messageNode struct {
	rect        *ge.Rect
	label       *ge.Label
	targetLine  *ge.Line
	targetLine2 *ge.Line
	camera      *viewport.Camera

	trackedObject ge.SceneObject
	uiLayer       *uiLayer

	pos       gmath.Vec
	targetPos ge.Pos
	screenPos bool

	text     string
	width    float64
	height   float64
	xpadding float64
}

func newScreenTutorialHintNode(camera *viewport.Camera, uiLayer *uiLayer, pos, targetPos gmath.Vec, text string) *messageNode {
	return &messageNode{
		uiLayer:   uiLayer,
		pos:       pos,
		targetPos: ge.Pos{Offset: targetPos},
		text:      text,
		camera:    camera,
		screenPos: true,
	}
}

func newWorldTutorialHintNode(camera *viewport.Camera, uiLayer *uiLayer, pos gmath.Vec, targetPos ge.Pos, text string) *messageNode {
	return &messageNode{
		uiLayer:   uiLayer,
		pos:       pos,
		targetPos: targetPos,
		text:      text,
		camera:    camera,
	}
}

func (m *messageNode) SetPos(pos gmath.Vec) {
	m.pos = pos
	m.rect.Pos.Offset = m.pos
	m.label.Pos.Offset = m.pos
}

func (m *messageNode) Init(scene *ge.Scene) {
	ff := scene.Context().Loader.LoadFont(assets.FontTiny)
	bounds := text.BoundString(ff.Face, m.text)
	m.width = (float64(bounds.Dx()) + 16) + m.xpadding
	m.height = (float64(bounds.Dy()) + 20)

	m.rect = ge.NewRect(scene.Context(), m.width, m.height)
	m.rect.OutlineColorScale.SetColor(ge.RGB(0x5e5a5d))
	m.rect.OutlineWidth = 1
	m.rect.FillColorScale.SetRGBA(0x13, 0x1a, 0x22, 230)
	m.rect.Centered = false
	m.rect.Pos.Offset = m.pos

	m.label = scene.NewLabel(assets.FontTiny)
	m.label.AlignHorizontal = ge.AlignHorizontalCenter
	m.label.AlignVertical = ge.AlignVerticalCenter
	m.label.Width = m.width
	m.label.Height = m.height
	m.label.Pos.Offset = m.pos
	m.label.Text = m.text
	m.label.ColorScale.SetColor(ge.RGB(0x9dd793))

	if !m.targetPos.Resolve().IsZero() {
		m.targetLine = ge.NewLine(ge.Pos{}, ge.Pos{})
		m.targetLine.Width = 1
		var clr ge.ColorScale
		clr.SetColor(ge.RGB(0x9dd793))
		m.targetLine.SetColorScale(clr)
		m.camera.AddGraphicsAbove(m.targetLine)

		m.targetLine2 = ge.NewLine(ge.Pos{}, ge.Pos{})
		m.targetLine2.Width = 1
		m.targetLine2.SetColorScale(clr)
		m.camera.AddGraphicsAbove(m.targetLine2)
	}

	m.uiLayer.AddGraphics(m.rect)
	m.uiLayer.AddGraphics(m.label)
}

func (m *messageNode) UpdateText(s string) {
	m.text = s
	m.label.Text = s
}

func (m *messageNode) Update(delta float64) {
	if m.targetLine != nil {
		m.targetLine.Visible = m.uiLayer.Visible
		m.targetLine2.Visible = m.uiLayer.Visible
	}
	if m.targetLine != nil && m.trackedObject != nil && m.trackedObject.IsDisposed() {
		m.targetLine.Dispose()
		m.targetLine2.Dispose()
		m.targetLine = nil
		m.targetLine2 = nil
		m.trackedObject = nil
	}

	if m.targetLine != nil {
		beginPos := m.camera.Offset.Add(m.pos)
		beginPos.Y++
		var endPos gmath.Vec
		if m.screenPos {
			endPos = m.camera.Offset.Add(m.targetPos.Offset)
		} else {
			endPos = m.targetPos.Resolve()
		}
		beginPos.X += m.width

		m.targetLine.BeginPos = ge.Pos{Offset: beginPos}
		m.targetLine.EndPos = ge.Pos{Offset: endPos}

		beginPos.Y += m.height - 2
		m.targetLine2.BeginPos = ge.Pos{Offset: beginPos}
		m.targetLine2.EndPos = ge.Pos{Offset: endPos}
	}
}

func (m *messageNode) HideLines() {
	if m.targetLine != nil {
		m.targetLine.Visible = false
		m.targetLine2.Visible = false
	}
}

func (m *messageNode) IsDisposed() bool {
	return m.rect.IsDisposed()
}

func (m *messageNode) Dispose() {
	m.rect.Dispose()
	m.label.Dispose()

	if m.targetLine != nil {
		m.targetLine.Dispose()
		m.targetLine2.Dispose()
	}
}
