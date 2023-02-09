package staging

import (
	"github.com/quasilyte/colony-game/assets"
	"github.com/quasilyte/colony-game/controls"
	"github.com/quasilyte/colony-game/gameui"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
)

type specialChoiceKind int

const (
	specialChoiceNone specialChoiceKind = iota
	specialChoiceMoveColony
	specialDecreaseRadius
	specialIncreaseRadius
	specialBuildColony
)

type selectedChoice struct {
	Faction   factionTag
	Option    choiceOption
	UseCursor bool
}

type choiceOptionSlot struct {
	icon    *ge.Sprite
	label   *ge.Label
	labelBg *ge.Rect
	option  choiceOption
}

type choiceOption struct {
	text    string
	effects []choiceOptionEffect
	special specialChoiceKind
	cost    float64
}

type choiceOptionEffect struct {
	priority colonyPriority
	value    float64
}

var specialChoicesList = []choiceOption{
	{
		text:    "move colony",
		special: specialChoiceMoveColony,
		cost:    20,
	},
	{
		text:    "build new colony",
		special: specialBuildColony,
		cost:    35,
	},
	{
		text:    "increase radius",
		special: specialIncreaseRadius,
		cost:    15,
	},
	{
		text:    "decrease radius",
		special: specialDecreaseRadius,
		cost:    5,
	},
}

var choiceOptionList = []choiceOption{
	{
		text: "resources",
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.2},
		},
	},
	{
		text: "growth",
		effects: []choiceOptionEffect{
			{priority: priorityGrowth, value: 0.2},
		},
	},
	{
		text: "security",
		effects: []choiceOptionEffect{
			{priority: prioritySecurity, value: 0.2},
		},
	},
	{
		text: "evolution",
		effects: []choiceOptionEffect{
			{priority: priorityEvolution, value: 0.2},
		},
	},

	{
		text: "resources+growth",
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.15},
			{priority: priorityGrowth, value: 0.15},
		},
	},
	{
		text: "resources+security",
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.15},
			{priority: prioritySecurity, value: 0.15},
		},
	},
	{
		text: "resources+evolution",
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.15},
			{priority: priorityEvolution, value: 0.15},
		},
	},
	{
		text: "growth+security",
		effects: []choiceOptionEffect{
			{priority: priorityGrowth, value: 0.15},
			{priority: prioritySecurity, value: 0.15},
		},
	},
	{
		text: "growth+evolution",
		effects: []choiceOptionEffect{
			{priority: priorityGrowth, value: 0.15},
			{priority: priorityEvolution, value: 0.15},
		},
	},
	{
		text: "security+evolution",
		effects: []choiceOptionEffect{
			{priority: prioritySecurity, value: 0.15},
			{priority: priorityEvolution, value: 0.15},
		},
	},
}

type choiceState int

const (
	choiceCharging choiceState = iota
	choiceReady
)

type choiceWindowNode struct {
	pos gmath.Vec

	scene *ge.Scene

	input *input.Handler

	state choiceState

	Enabled bool

	targetValue float64
	value       float64

	openSprite   *ge.Sprite
	foldedSprite *ge.Sprite

	choices []*choiceOptionSlot

	shuffledOptions []choiceOption

	beforeSpecialShuffle int
	specialChoices       []choiceOption

	chargeBar *gameui.ProgressBar

	EventChoiceSelected gsignal.Event[selectedChoice]
}

func newChoiceWindowNode(pos gmath.Vec, h *input.Handler) *choiceWindowNode {
	return &choiceWindowNode{pos: pos, input: h}
}

func (w *choiceWindowNode) Init(scene *ge.Scene) {
	w.scene = scene

	w.shuffledOptions = make([]choiceOption, len(choiceOptionList))
	copy(w.shuffledOptions, choiceOptionList)

	w.specialChoices = make([]choiceOption, len(specialChoicesList))
	copy(w.specialChoices, specialChoicesList)

	w.openSprite = scene.NewSprite(assets.ImageChoiceWindow)
	w.openSprite.Centered = false
	w.openSprite.Pos.Base = &w.pos
	scene.AddGraphics(w.openSprite)

	w.foldedSprite = scene.NewSprite(assets.ImageChoiceRechargeWindow)
	w.foldedSprite.Centered = false
	w.foldedSprite.Pos.Base = &w.pos
	w.foldedSprite.Pos.Offset.Y = 136
	scene.AddGraphics(w.foldedSprite)

	bgColor := ge.RGB(0x080c10)
	fgColor := ge.RGB(0x5994b9)
	chargeBarPos := w.foldedSprite.Pos.WithOffset(22, 16)
	w.chargeBar = gameui.NewProgressBar(chargeBarPos, 224-44, 24, bgColor, fgColor)
	scene.AddObject(w.chargeBar)

	icons := [...]resource.ImageID{
		assets.ImageButtonY,
		assets.ImageButtonB,
		assets.ImageButtonA,
		assets.ImageButtonX,
		assets.ImageButtonRB,
	}
	offsetY := 18.0
	w.choices = make([]*choiceOptionSlot, 5)
	for i := range w.choices {
		l := scene.NewLabel(assets.FontTiny)
		l.Pos.Base = &w.pos
		l.Pos.Offset.Y = offsetY
		l.Pos.Offset.X = 44
		l.AlignVertical = ge.AlignVerticalCenter
		l.Width = 224 - 60
		l.Height = 28
		bg := ge.NewRect(scene.Context(), l.Width, l.Height)
		bg.Centered = false
		bg.Pos = l.Pos.WithOffset(-2, 0)
		bg.FillColorScale.SetColor(bgColor)
		scene.AddGraphics(bg)
		scene.AddGraphics(l)
		icon := scene.NewSprite(icons[i])
		icon.Centered = false
		icon.Pos = l.Pos.WithOffset(-34, 0)
		scene.AddGraphics(icon)
		w.choices[i] = &choiceOptionSlot{
			icon:    icon,
			label:   l,
			labelBg: bg,
		}
		offsetY += l.Height + 4
	}
	// centerPos := ge.Pos{
	// 	Base: &w.pos,
	// 	// Offset: gmath.Vec{
	// 	// 	X: w.openSprite.FrameWidth * 0.5,
	// 	// 	Y: w.openSprite.FrameHeight * 0.5,
	// 	// },
	// }
	// w.choices[0].label.Pos = centerPos.WithOffset(0, -32)
	// w.choices[1].label.Pos = centerPos.WithOffset(48, 0)
	// w.choices[2].label.Pos = centerPos.WithOffset(0, 32)
	// w.choices[3].label.Pos = centerPos.WithOffset(-48, 0)

	w.startCharging(10)
}

func (w *choiceWindowNode) IsDisposed() bool {
	return false
}

func (w *choiceWindowNode) revealChoices() {
	w.openSprite.Visible = true
	w.foldedSprite.Visible = false
	w.chargeBar.Visible = false
	for _, o := range w.choices {
		o.label.Visible = true
		o.icon.Visible = true
		o.labelBg.Visible = true
	}
	w.state = choiceReady

	gmath.Shuffle(w.scene.Rand(), w.shuffledOptions)
	for i, o := range w.shuffledOptions[:4] {
		w.choices[i].option = o
		w.choices[i].label.Text = o.text
	}

	if w.beforeSpecialShuffle == 0 {
		gmath.Shuffle(w.scene.Rand(), w.specialChoices)
		w.beforeSpecialShuffle = gmath.Clamp(4, 1, len(specialChoicesList))
	}
	w.beforeSpecialShuffle--
	specialIndex := w.beforeSpecialShuffle

	specialOption := w.specialChoices[specialIndex]
	w.choices[4].option = specialOption
	w.choices[4].label.Text = specialOption.text

	w.scene.Audio().PlaySound(assets.AudioChoiceReady)
}

func (w *choiceWindowNode) startCharging(targetValue float64) {
	w.chargeBar.SetValue(0)
	w.value = 0
	w.targetValue = targetValue
	w.openSprite.Visible = false
	w.foldedSprite.Visible = true
	w.chargeBar.Visible = true
	for _, o := range w.choices {
		o.label.Visible = false
		o.icon.Visible = false
		o.labelBg.Visible = false
	}
	w.state = choiceCharging
}

func (w *choiceWindowNode) Update(delta float64) {
	switch w.state {
	case choiceCharging:
		w.value += delta
		if w.value >= w.targetValue {
			w.revealChoices()
			return
		}
		w.chargeBar.SetValue(w.value / w.targetValue)

	case choiceReady:
		actions := [...]input.Action{
			controls.ActionChoice1,
			controls.ActionChoice2,
			controls.ActionChoice3,
			controls.ActionChoice4,
			controls.ActionChoice5,
		}
		for i, a := range actions {
			info, ok := w.input.JustPressedActionInfo(a)
			if !ok {
				continue
			}
			w.activateChoice(i, info)
			break
		}
	}
}

func (w *choiceWindowNode) activateChoice(i int, info input.EventInfo) {
	if !w.Enabled {
		w.scene.Audio().PlaySound(assets.AudioError)
		return
	}

	selectedFaction := factionTag(i + 1)
	choice := selectedChoice{
		Faction:   selectedFaction,
		Option:    w.choices[i].option,
		UseCursor: info.IsMouseEvent(),
	}

	delayRoll := w.scene.Rand().FloatRange(0.8, 1.2)
	if i == 4 {
		// Special action selected.
		w.startCharging(w.choices[i].option.cost * delayRoll)
	} else {
		w.startCharging(10.0 * delayRoll)
	}

	w.scene.Audio().PlaySound(assets.AudioChoiceMade)

	w.EventChoiceSelected.Emit(choice)
}
