package staging

import (
	"strings"

	"github.com/quasilyte/colony-game/assets"
	"github.com/quasilyte/colony-game/controls"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
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
	l.Pos.Offset = m.dialogueWindow.Pos.Offset.Add(gmath.Vec{X: 26, Y: 20})
	l.Width = 228 + 40
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

	m.steps = []tutorialStep{
		{
			text: strings.Join([]string{
				"Welcome to the Roboden tutorial!",
				"",
				"You continue by clicking",
				"on the tutorial window.",
				"",
				"Try it right now, get to the",
				"next tutorial message.",
			}, "\n"),
		},
		{
			text: strings.Join([]string{
				"Great!",
				"",
				"The map is bigger than the",
				"camera boundaries.",
				"",
				"Try panning the camery by",
				"placing your cursor close to",
				"the edge of the screen.",
				"",
				"You can also use arrow keys.",
			}, "\n"),
		},
		{
			text: strings.Join([]string{
				"Have you discovered your second",
				"robot colony on this level?",
				"",
				"It's located to the south-west",
				"from your first base position.",
				"",
				"You can press [tab] to switch",
				"between colonies easily.",
				"If you only have one base, it'll",
				"center the camera on that.",
			}, "\n"),
		},

		{
			refreshChoices: true,
			text: strings.Join([]string{
				"You can move a colony by",
				"clicking the right mouse button",
				"on the destination spot.",
				"",
				"Keep in mind that colonies",
				"have limited max flight range.",
				"",
				"Try moving either of your bases.",
			}, "\n"),
		},

		{
			text: strings.Join([]string{
				"You lose if all of your colonies",
				"are destroyed.",
				"",
				"You win if you defeat the boss.",
				"",
				"There is a boss-detection radar",
				"in the upper right corner.",
				"Try finding the boss on this map.",
			}, "\n"),
		},

		{
			text: strings.Join([]string{
				"In this tutorial level the",
				"boss is immovable. It's also",
				"decreased in power levels",
				"as they're under 9000.",
				"",
				"It's much harder to defeat",
				"the boss in the real game.",
			}, "\n"),
		},

		{
			text: strings.Join([]string{
				"As you might have noticed,",
				"your colonies were active",
				"all this time.",
				"",
				"Every colony should have its",
				"drones: workers and fighters.",
				"",
				"Workers are good at keeping",
				"the base functioning while",
				"fighters are defending it and,",
				"ultimately, defeat the boss.",
			}, "\n"),
		},

		{
			text: strings.Join([]string{
				"All colonies need resources.",
				"",
				"The colony resource level is",
				"indicated by the middle yellow",
				"bar on its body. It can be empty.",
				"",
				"The resources are collected by",
				"workers around the base.",
			}, "\n"),
		},

		{
			refreshChoices: true,
			text: strings.Join([]string{
				"Apart from the movement,",
				"there is also a choice selection.",
				"",
				"Both movement and choices use",
				"the same action gauge that",
				"needs to recharge afterwards.",
			}, "\n"),
		},

		{
			refreshChoices: true,
			text: strings.Join([]string{
				"Try using any of the actions",
				"from the bottom right corner.",
				"",
				"Click on the action label with",
				"your left mouse button.",
				"It's also possible to use the",
				"hotkeys: [1]-[5] and [q]-[t].",
				"",
				"Actions are shared between the",
				"colonies, but the effects are not.",
			}, "\n"),
		},

		{
			text: strings.Join([]string{
				"Every choice does two main",
				"things. It adjusts the colony",
				"priorities and shifts the",
				"robot faction distribution.",
				"",
				"Priorities affect the way colony",
				"is behaving: what does it do,",
				"and how often.",
				"",
				"Factions affect the robots.",
			}, "\n"),
		},

		{
			text: strings.Join([]string{
				"The fifth action choice is an",
				"exception from the rules.",
				"",
				"It doesn't affect the factions.",
				"It doesn't change the priorities.",
			}, "\n"),
		},

		{
			text: strings.Join([]string{
				"Try getting more colored drones.",
				"",
				"Mine resources, pick faction",
				"choices, evolve your drones.",
				"",
				"If resources run out, move",
				"colony to a new location.",
			}, "\n"),
		},

		{
			text: strings.Join([]string{
				"Now you know all the basics.",
				"",
				"It's time for you to discover",
				"the rest by yourself.",
				"",
				"You can win the tutorial level",
				"by defeating the boss. Or you",
				"can leave by pressing [esc].",
			}, "\n"),
		},

		{
			text: strings.Join([]string{
				"Good luck and good hunting.",
			}, "\n"),
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

	if info, ok := m.input.JustPressedActionInfo(controls.ActionClick); ok {
		if m.windowRect.Contains(info.Pos) {
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
