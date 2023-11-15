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

	highlightRect  *ge.Rect
	highlightStep  float64
	highlightValue float64
	highlight      bool

	trackedObject ge.SceneObject

	pos       gmath.Vec
	targetPos ge.Pos
	screenPos bool

	text     string
	width    float64
	height   float64
	xpadding float64
}

func estimateMessageBounds(s string, xpadding float64) (width, height float64) {
	bounds := text.BoundString(assets.BitmapFont1, s)
	width = (float64(bounds.Dx()) + 16) + xpadding
	height = (float64(bounds.Dy()) + 16)
	return width, height
}

func newScreenTutorialHintNode(camera *viewport.Camera, pos, targetPos gmath.Vec, text string) *messageNode {
	return &messageNode{
		pos:       pos,
		targetPos: ge.Pos{Offset: targetPos},
		text:      text,
		camera:    camera,
		screenPos: true,
	}
}

func newWorldTutorialHintNode(camera *viewport.Camera, pos gmath.Vec, targetPos ge.Pos, text string) *messageNode {
	return &messageNode{
		pos:       pos,
		targetPos: targetPos,
		text:      text,
		camera:    camera,
	}
}

func (m *messageNode) SetPos(pos gmath.Vec) {
	m.pos = pos
	m.rect.Pos.Offset = m.pos
	m.label.Pos.Offset = m.pos.Add(gmath.Vec{Y: 4})
}

func (m *messageNode) ContainsPos(pos gmath.Vec) bool {
	bounds := gmath.Rect{
		Min: m.pos,
		Max: m.pos.Add(gmath.Vec{X: m.width, Y: m.height}),
	}
	return bounds.Contains(pos)
}

func (m *messageNode) Init(scene *ge.Scene) {
	m.width, m.height = estimateMessageBounds(m.text, m.xpadding)

	m.rect = ge.NewRect(scene.Context(), m.width, m.height)
	m.rect.OutlineColorScale.SetColor(ge.RGB(0x5e5a5d))
	m.rect.OutlineWidth = 1
	m.rect.FillColorScale.SetRGBA(0x13, 0x1a, 0x22, 160)
	m.rect.Centered = false
	m.rect.Pos.Offset = m.pos

	m.highlightRect = ge.NewRect(scene.Context(), m.width+2, m.height+2)
	m.highlightRect.OutlineColorScale.SetColor(ge.RGB(0xe7c34b))
	m.highlightRect.FillColorScale.SetRGBA(0, 0, 0, 0)
	m.highlightRect.OutlineWidth = 2
	m.highlightRect.Centered = false
	m.highlightRect.Pos.Offset = m.pos.Sub(gmath.Vec{X: 1, Y: 1})
	m.highlightRect.Visible = false

	m.label = ge.NewLabel(assets.BitmapFont1)
	m.label.AlignHorizontal = ge.AlignHorizontalCenter
	m.label.AlignVertical = ge.AlignVerticalCenter
	m.label.Width = m.width
	m.label.Height = m.height
	m.label.Pos.Offset = m.pos.Add(gmath.Vec{Y: 4})
	m.label.Text = m.text
	m.label.SetColorScaleRGBA(0x9d, 0xd7, 0x93, 0xff)

	if !m.targetPos.Resolve().IsZero() {
		m.targetLine = ge.NewLine(ge.Pos{}, ge.Pos{})
		m.targetLine.Width = 1
		var clr ge.ColorScale
		clr.SetRGBA(0x9d, 0xd7, 0x93, 100)
		m.targetLine.SetColorScale(clr)
		m.camera.Private.AddGraphicsAbove(m.targetLine)

		m.targetLine2 = ge.NewLine(ge.Pos{}, ge.Pos{})
		m.targetLine2.Width = 1
		m.targetLine2.SetColorScale(clr)
		m.camera.Private.AddGraphicsAbove(m.targetLine2)
	}

	m.camera.UI.AddGraphicsAbove(m.rect)
	m.camera.UI.AddGraphicsAbove(m.highlightRect)
	m.camera.UI.AddGraphicsAbove(m.label)
}

func (m *messageNode) UpdateText(s string) {
	m.text = s
	m.label.Text = s
}

func (m *messageNode) Update(delta float64) {
	if m.highlight {
		m.highlightValue = gmath.Clamp(m.highlightValue+(delta*m.highlightStep), 0, 1)
		if m.highlightValue == 0 {
			m.highlightStep = +1
		} else if m.highlightValue == 1 {
			m.highlightStep = -1
		}
		m.highlightRect.OutlineColorScale.A = float32(m.highlightValue)
	}

	if m.targetLine != nil {
		m.targetLine.Visible = m.camera.UI.Visible
		m.targetLine2.Visible = m.camera.UI.Visible
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

func (m *messageNode) Highlight() {
	m.highlight = true
	m.highlightRect.OutlineColorScale.A = 0
	m.highlightRect.Visible = true
	m.highlightStep = 1
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
	m.highlightRect.Dispose()
	m.rect.Dispose()
	m.label.Dispose()

	if m.targetLine != nil {
		m.targetLine.Dispose()
		m.targetLine2.Dispose()
	}
}
