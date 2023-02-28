package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
)

type tutorialManager struct {
	input *input.Handler

	scene *ge.Scene

	dialogueWindow *ge.Sprite
	windowRect     gmath.Rect

	choiceWindow *choiceWindowNode

	label   *ge.Label
	labelBg *ge.Rect

	tutorialStep int
	steps        []tutorialStep

	enabled bool
}

type tutorialStep struct {
	text string

	refreshChoices bool
}

func newTutorialManager(h *input.Handler, choices *choiceWindowNode) *tutorialManager {
	return &tutorialManager{
		input:        h,
		choiceWindow: choices,
	}
}

func (m *tutorialManager) Init(scene *ge.Scene) {
	m.scene = scene

	m.dialogueWindow = scene.NewSprite(assets.ImageTutorialDialogue)
	m.dialogueWindow.Centered = false
	m.dialogueWindow.Pos.Offset = gmath.Vec{X: 16, Y: 320 - 20}
	scene.AddGraphics(m.dialogueWindow)

	m.windowRect = gmath.Rect{
		Min: m.dialogueWindow.AnchorPos().Resolve(),
	}
	m.windowRect.Max = m.windowRect.Min.Add(gmath.Vec{
		X: m.dialogueWindow.ImageWidth(),
		Y: m.dialogueWindow.ImageHeight(),
	})

	l := scene.NewLabel(assets.FontTiny)
	l.ColorScale.SetColor(ge.RGB(0x9dd793))
	l.Pos.Offset = m.dialogueWindow.Pos.Offset.Add(gmath.Vec{X: 25, Y: 20})
	l.Width = 228 + 72
	l.Height = 186
	l.AlignHorizontal = ge.AlignHorizontalCenter
	l.AlignVertical = ge.AlignVerticalCenter
	bg := ge.NewRect(scene.Context(), l.Width+4, l.Height+12)
	bg.Centered = false
	bg.Pos = l.Pos.WithOffset(-2, -6)
	bg.FillColorScale.SetColor(ge.RGB(0x080c10))
	scene.AddGraphics(bg)
	scene.AddGraphics(l)
	m.label = l
	m.labelBg = bg

	m.enabled = true

	d := m.scene.Dict()

	m.steps = []tutorialStep{
		{
			text: d.Get("tutorial.message1"),
		},
		{
			text: d.Get("tutorial.message2"),
		},
		{
			text: d.Get("tutorial.message3"),
		},

		{
			refreshChoices: true,
			text:           d.Get("tutorial.message4"),
		},

		{
			text: d.Get("tutorial.message5"),
		},

		{
			text: d.Get("tutorial.message6"),
		},

		{
			text: d.Get("tutorial.message7"),
		},

		{
			refreshChoices: true,
			text:           d.Get("tutorial.message8"),
		},

		{
			refreshChoices: true,
			text:           d.Get("tutorial.message9"),
		},

		{
			text: d.Get("tutorial.message10"),
		},

		{
			text: d.Get("tutorial.message11"),
		},

		{
			refreshChoices: true,
			text:           d.Get("tutorial.message12"),
		},

		{
			text: d.Get("tutorial.message13"),
		},

		{
			text: d.Get("tutorial.message14"),
		},

		{
			text: d.Get("tutorial.message15"),
		},

		{
			text: d.Get("tutorial.message_last"),
		},
	}

	l.Text = m.steps[0].text
}

func (m *tutorialManager) IsDisposed() bool {
	return false
}

func (m *tutorialManager) Update(delta float64) {
	if !m.enabled {
		return
	}

	cursor := m.choiceWindow.cursor
	if pos, ok := cursor.ClickPos(controls.ActionClick); ok {
		if m.windowRect.Contains(pos) {
			m.enabled = false
			m.setWindowVisibility(false)
			m.scene.Audio().PlaySound(assets.AudioClick)
			m.scene.DelayedCall(0.25, func() {
				m.enabled = true
				m.openNextMessage()
			})
		}
	}
}

func (m *tutorialManager) setWindowVisibility(visible bool) {
	m.dialogueWindow.Visible = visible
	m.label.Visible = visible
	m.labelBg.Visible = visible
}

func (m *tutorialManager) openNextMessage() {
	m.tutorialStep++
	if m.tutorialStep >= len(m.steps) {
		return
	}
	step := m.steps[m.tutorialStep]
	m.label.Text = step.text
	if step.refreshChoices {
		m.choiceWindow.ForceRefresh()
	}
	m.setWindowVisibility(true)
}
