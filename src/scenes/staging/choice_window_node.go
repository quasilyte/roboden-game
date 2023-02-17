package staging

import (
	"strings"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui"
)

type specialChoiceKind int

const (
	specialChoiceNone specialChoiceKind = iota
	specialDecreaseRadius
	specialIncreaseRadius
	specialBuildColony
	specialAttack
	specialChoiceMoveColony
)

type selectedChoice struct {
	Faction factionTag
	Option  choiceOption
}

type choiceOptionSlot struct {
	icon    *ge.Sprite
	iconBg  *ge.Sprite
	label   *ge.Label
	labelBg *ge.Rect
	rect    gmath.Rect
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
		text:    "attack",
		special: specialAttack,
		cost:    5,
	},
	{
		text:    "build_colony",
		special: specialBuildColony,
		cost:    40,
	},
	{
		text:    "increase_radius",
		special: specialIncreaseRadius,
		cost:    15,
	},
	{
		text:    "decrease_radius",
		special: specialDecreaseRadius,
		cost:    4,
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

	selectedColony *colonyCoreNode

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

	d := scene.Dict()

	w.shuffledOptions = make([]choiceOption, len(choiceOptionList))
	copy(w.shuffledOptions, choiceOptionList)

	// Now translate the options.
	for i := range w.shuffledOptions {
		o := &w.shuffledOptions[i]
		keys := strings.Split(o.text, "+")
		for _, k := range keys {
			o.text = strings.Replace(o.text, k, d.Get("game.choice", k), 1)
		}
	}

	w.specialChoices = make([]choiceOption, len(specialChoicesList))
	copy(w.specialChoices, specialChoicesList)

	// Now translate the special choices.
	for i := range w.specialChoices {
		o := &w.specialChoices[i]
		o.text = strings.Replace(o.text, o.text, d.Get("game.choice", o.text), 1)
	}

	w.openSprite = scene.NewSprite(assets.ImageChoiceWindow)
	w.openSprite.Centered = false
	w.openSprite.Pos.Base = &w.pos
	scene.AddGraphics(w.openSprite)

	w.foldedSprite = scene.NewSprite(assets.ImageChoiceRechargeWindow)
	w.foldedSprite.Centered = false
	w.foldedSprite.Pos.Base = &w.pos
	w.foldedSprite.Pos.Offset.Y = 136 + 8
	scene.AddGraphics(w.foldedSprite)

	bgColor := ge.RGB(0x080c10)
	fgColor := ge.RGB(0x5994b9)
	chargeBarPos := w.foldedSprite.Pos.WithOffset(22, 16)
	w.chargeBar = gameui.NewProgressBar(chargeBarPos, 232-44, 24, bgColor, fgColor)
	scene.AddObject(w.chargeBar)

	icons := [...]resource.ImageID{
		assets.ImageYellowLogo,
		assets.ImageRedLogo,
		assets.ImageGreenLogo,
		assets.ImageBlueLogo,
	}
	offsetY := 18.0
	w.choices = make([]*choiceOptionSlot, 5)
	for i := range w.choices {
		l := scene.NewLabel(assets.FontTiny)
		l.ColorScale.SetColor(ge.RGB(0x9dd793))
		l.Pos.Base = &w.pos
		l.Pos.Offset.Y = offsetY
		l.Pos.Offset.X = 50
		l.AlignVertical = ge.AlignVerticalCenter
		l.Width = 224 - 60
		l.Height = 28
		bg := ge.NewRect(scene.Context(), l.Width, l.Height)
		bg.Centered = false
		bg.Pos = l.Pos.WithOffset(-2, 0)
		bg.FillColorScale.SetColor(bgColor)
		scene.AddGraphics(bg)
		scene.AddGraphics(l)
		choice := &choiceOptionSlot{
			label:   l,
			labelBg: bg,
			rect: gmath.Rect{
				Min: bg.AnchorPos().Resolve(),
				Max: bg.AnchorPos().Resolve().Add(gmath.Vec{X: l.Width, Y: l.Height}),
			},
		}
		if i < len(icons) {
			iconBg := scene.NewSprite(assets.ImageLogoBg)
			iconBg.Centered = false
			iconBg.Pos = l.Pos.WithOffset(-36, -2)
			scene.AddGraphics(iconBg)
			choice.iconBg = iconBg

			icon := scene.NewSprite(icons[i])
			icon.Centered = false
			icon.Pos = iconBg.Pos
			scene.AddGraphics(icon)
			choice.icon = icon
		}
		w.choices[i] = choice
		offsetY += l.Height + 6
	}

	w.startCharging(10)
}

func (w *choiceWindowNode) IsDisposed() bool {
	return false
}

func (w *choiceWindowNode) ForceRefresh() {
	if w.state == choiceReady {
		return
	}
	w.value = w.targetValue
}

func (w *choiceWindowNode) revealChoices() {
	w.openSprite.Visible = true
	w.foldedSprite.Visible = false
	w.chargeBar.Visible = false
	for _, o := range w.choices {
		o.label.Visible = true
		if o.icon != nil {
			o.icon.Visible = true
			o.iconBg.Visible = true
		}
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
		o.labelBg.Visible = false
		if o.icon != nil {
			o.icon.Visible = false
			o.iconBg.Visible = false
		}
	}
	w.state = choiceCharging
}

func (w *choiceWindowNode) Update(delta float64) {
	if w.selectedColony == nil {
		return
	}

	switch w.state {
	case choiceCharging:
		w.value += delta
		if w.value >= w.targetValue {
			w.revealChoices()
			return
		}
		w.chargeBar.SetValue(w.value / w.targetValue)

	case choiceReady:
		if info, ok := w.input.JustPressedActionInfo(controls.ActionMoveChoice); ok {
			globalClickPos := info.Pos.Add(w.selectedColony.world.camera.Offset)
			if globalClickPos.DistanceTo(w.selectedColony.pos) > 28 {
				w.activateMoveChoice()
				return
			}
		}
		if info, ok := w.input.JustPressedActionInfo(controls.ActionClick); ok {
			for i, choice := range w.choices {
				if choice.rect.Contains(info.Pos) {
					w.activateChoice(i)
					return
				}
			}
		}
		actions := [...]input.Action{
			controls.ActionChoice1,
			controls.ActionChoice2,
			controls.ActionChoice3,
			controls.ActionChoice4,
			controls.ActionChoice5,
		}
		for i, a := range actions {
			if w.input.ActionIsJustPressed(a) {
				w.activateChoice(i)
				return
			}
		}
	}
}

func (w *choiceWindowNode) activateMoveChoice() {
	if !w.Enabled {
		w.scene.Audio().PlaySound(assets.AudioError)
		return
	}
	choice := selectedChoice{
		Option: choiceOption{special: specialChoiceMoveColony},
	}
	delayRoll := w.scene.Rand().FloatRange(0.8, 1.2)
	w.startCharging(20.0 * delayRoll)
	w.scene.Audio().PlaySound(assets.AudioChoiceMade)
	w.EventChoiceSelected.Emit(choice)
}

func (w *choiceWindowNode) activateChoice(i int) {
	if !w.Enabled {
		w.scene.Audio().PlaySound(assets.AudioError)
		return
	}

	selectedFaction := factionTag(i + 1)
	choice := selectedChoice{
		Faction: selectedFaction,
		Option:  w.choices[i].option,
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
