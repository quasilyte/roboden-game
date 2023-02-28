package staging

import (
	"image/color"
	"math"
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
	specialBuildGunpoint
	specialBuildColony
	specialAttack
	specialChoiceMoveColony
)

type selectedChoice struct {
	Faction factionTag
	Option  choiceOption
	Pos     gmath.Vec
}

type choiceOptionSlot struct {
	floppy     *ge.Sprite
	icon       *ge.Sprite
	labelLight *ge.Label
	labelDark  *ge.Label
	rect       gmath.Rect
	option     choiceOption
}

type choiceOption struct {
	text    string
	effects []choiceOptionEffect
	special specialChoiceKind
	icon    resource.ImageID
	cost    float64
}

type choiceOptionEffect struct {
	priority colonyPriority
	value    float64
}

var specialChoicesTable = [...]choiceOption{
	specialAttack: {
		text:    "attack",
		special: specialAttack,
		cost:    5,
		icon:    assets.ImageActionAttack,
	},
	specialBuildColony: {
		text:    "build_colony",
		special: specialBuildColony,
		cost:    40,
		icon:    assets.ImageActionBuildColony,
	},
	specialBuildGunpoint: {
		text:    "build_gunpoint",
		special: specialBuildGunpoint,
		cost:    15,
		icon:    assets.ImageActionBuildTurret,
	},
	specialIncreaseRadius: {
		text:    "increase_radius",
		special: specialIncreaseRadius,
		cost:    15,
		icon:    assets.ImageActionIncreaseRadius,
	},
	specialDecreaseRadius: {
		text:    "decrease_radius",
		special: specialDecreaseRadius,
		cost:    4,
		icon:    assets.ImageActionDecreaseRadius,
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

	floppyOffsetX float64
	selectedSlide float64
	selectedIndex int

	targetValue float64
	value       float64

	foldedSprite *ge.Sprite

	choices []*choiceOptionSlot

	shuffledOptions []choiceOption

	beforeSpecialShuffle int
	canBuildBase         bool
	specialChoiceKinds   []specialChoiceKind
	specialChoices       []choiceOption

	cursor *cursorNode

	chargeBar *gameui.ProgressBar

	EventChoiceSelected gsignal.Event[selectedChoice]
}

func newChoiceWindowNode(pos gmath.Vec, h *input.Handler, cursor *cursorNode) *choiceWindowNode {
	return &choiceWindowNode{
		pos:           pos,
		input:         h,
		cursor:        cursor,
		selectedIndex: -1,
	}
}

func (w *choiceWindowNode) Init(scene *ge.Scene) {
	w.scene = scene

	d := scene.Dict()

	w.shuffledOptions = make([]choiceOption, len(choiceOptionList))
	copy(w.shuffledOptions, choiceOptionList)

	translateText := func(s string) string {
		keys := strings.Split(s, "+")
		for _, k := range keys {
			s = strings.Replace(s, k, d.Get("game.choice", k), 1)
		}
		s = strings.ReplaceAll(s, "+", "\n+\n")
		s = strings.ReplaceAll(s, " ", "\n")
		return s
	}

	// Now translate the options.
	for i := range w.shuffledOptions {
		o := &w.shuffledOptions[i]
		o.text = translateText(o.text)
	}

	w.specialChoiceKinds = []specialChoiceKind{
		specialBuildColony,
		specialAttack,
		specialDecreaseRadius,
		specialIncreaseRadius,
	}

	// Now translate the special choices.
	w.specialChoices = make([]choiceOption, len(specialChoicesTable))
	copy(w.specialChoices, specialChoicesTable[:])
	for i := range w.specialChoices {
		o := &w.specialChoices[i]
		if o.text == "" {
			continue
		}
		o.text = translateText(o.text)
	}

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

	camera := w.selectedColony.world.camera
	floppies := [...]resource.ImageID{
		assets.ImageFloppyYellow,
		assets.ImageFloppyRed,
		assets.ImageFloppyGreen,
		assets.ImageFloppyBlue,
		assets.ImageFloppyGray,
	}
	fontColors := [...][2]color.RGBA{
		{ge.RGB(0x99943d), ge.RGB(0x666114)},
		{ge.RGB(0x804140), ge.RGB(0x4d1717)},
		{ge.RGB(0x40804a), ge.RGB(0x174d1f)},
		{ge.RGB(0x405680), ge.RGB(0x172a4d)},
		{ge.RGB(0x5e5a5d), ge.RGB(0x3d3a3c)},
	}
	offsetY := 8.0
	w.floppyOffsetX = camera.Rect.Width() - 144 - 8
	offset := gmath.Vec{X: w.floppyOffsetX, Y: 8}
	w.choices = make([]*choiceOptionSlot, 5)
	for i := range w.choices {
		floppy := scene.NewSprite(floppies[i])
		floppy.Centered = false
		floppy.Pos.Offset = offset
		scene.AddGraphics(floppy)

		offset.Y += floppy.ImageHeight() + offsetY

		darkLabel := scene.NewLabel(assets.FontTiny)
		darkLabel.ColorScale.SetColor(fontColors[i][1])
		darkLabel.Pos.Base = &floppy.Pos.Offset
		darkLabel.Pos.Offset = gmath.Vec{X: 48, Y: 6}
		darkLabel.AlignVertical = ge.AlignVerticalCenter
		darkLabel.AlignHorizontal = ge.AlignHorizontalCenter
		darkLabel.Width = 86
		darkLabel.Height = 62

		lightLabel := scene.NewLabel(assets.FontTiny)
		lightLabel.ColorScale.SetColor(fontColors[i][0])
		lightLabel.Pos = darkLabel.Pos
		lightLabel.Pos.Offset = lightLabel.Pos.Offset.Add(gmath.Vec{X: -2, Y: -2})
		lightLabel.AlignVertical = ge.AlignVerticalCenter
		lightLabel.AlignHorizontal = ge.AlignHorizontalCenter
		lightLabel.Width = 86 + 2
		lightLabel.Height = 62 + 2

		scene.AddGraphics(darkLabel)
		scene.AddGraphics(lightLabel)

		var icon *ge.Sprite
		if i == 4 {
			icon = ge.NewSprite(scene.Context())
			icon.Centered = false
			icon.Pos.Base = &floppy.Pos.Offset
			icon.Pos.Offset = gmath.Vec{X: 5, Y: 26}
			scene.AddGraphics(icon)
		}

		choice := &choiceOptionSlot{
			labelDark:  darkLabel,
			labelLight: lightLabel,
			floppy:     floppy,
			icon:       icon,
			rect: gmath.Rect{
				Min: floppy.Pos.Resolve(),
				Max: floppy.Pos.Resolve().Add(gmath.Vec{
					X: floppy.ImageWidth(),
					Y: floppy.ImageHeight(),
				}),
			},
		}

		w.choices[i] = choice
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
	w.foldedSprite.Visible = false
	w.chargeBar.Visible = false
	for i, o := range w.choices {
		if i == w.selectedIndex {
			o.floppy.Pos.Offset.X = w.floppyOffsetX
			continue
		}
		o.labelDark.Visible = true
		o.labelLight.Visible = true
		o.floppy.Visible = true
		if o.icon != nil {
			o.icon.Visible = true
		}
	}
	w.state = choiceReady

	gmath.Shuffle(w.scene.Rand(), w.shuffledOptions)
	for i, o := range w.shuffledOptions[:4] {
		w.choices[i].option = o
		w.choices[i].labelDark.Text = o.text
		w.choices[i].labelLight.Text = o.text
	}

	if w.beforeSpecialShuffle == 0 {
		w.canBuildBase = !w.canBuildBase
		gmath.Shuffle(w.scene.Rand(), w.specialChoiceKinds)
		w.beforeSpecialShuffle = gmath.Clamp(4, 1, len(w.specialChoiceKinds))
	}
	w.beforeSpecialShuffle--
	specialIndex := w.beforeSpecialShuffle

	specialOptionKind := w.specialChoiceKinds[specialIndex]
	if specialOptionKind == specialBuildColony {
		if !w.canBuildBase {
			specialOptionKind = specialBuildGunpoint
		}
	}
	specialOption := w.specialChoices[specialOptionKind]
	w.choices[4].option = specialOption
	w.choices[4].labelDark.Text = specialOption.text
	w.choices[4].labelLight.Text = specialOption.text
	w.choices[4].icon.SetImage(w.scene.LoadImage(specialOption.icon))

	w.scene.Audio().PlaySound(assets.AudioChoiceReady)
}

func (w *choiceWindowNode) startCharging(targetValue float64) {
	w.chargeBar.SetValue(0)
	w.value = 0
	w.targetValue = targetValue
	w.foldedSprite.Visible = true
	w.chargeBar.Visible = true
	for i, o := range w.choices {
		if i == w.selectedIndex {
			continue
		}
		o.labelDark.Visible = false
		o.labelLight.Visible = false
		o.floppy.Visible = false
		if o.icon != nil {
			o.icon.Visible = false
		}
	}
	w.state = choiceCharging
}

func (w *choiceWindowNode) Update(delta float64) {
	if w.selectedColony == nil {
		return
	}
	if w.state != choiceCharging {
		return
	}
	w.value += delta
	if w.value >= w.targetValue {
		w.revealChoices()
		return
	}
	w.chargeBar.SetValue(w.value / w.targetValue)

	const maxSelectedSlide float64 = 64
	if w.selectedSlide != -1 && w.selectedIndex != -1 {
		w.selectedSlide = gmath.ClampMax(w.selectedSlide+delta*16, maxSelectedSlide)
		w.choices[w.selectedIndex].floppy.Pos.Offset.X = math.Round(w.floppyOffsetX + w.selectedSlide)
		if w.selectedSlide == maxSelectedSlide {
			w.selectedSlide = -1
		}
	}

}

func (w *choiceWindowNode) HandleInput() {
	if w.selectedColony == nil {
		return
	}
	if w.state != choiceReady {
		return
	}
	if pos, ok := w.cursor.ClickPos(controls.ActionClick); ok {
		for i, choice := range w.choices {
			if choice.rect.Contains(pos) {
				w.activateChoice(i)
				return
			}
		}
	}
	if pos, ok := w.cursor.ClickPos(controls.ActionMoveChoice); ok {
		globalClickPos := pos.Add(w.selectedColony.world.camera.Offset)
		if globalClickPos.DistanceTo(w.selectedColony.pos) > 28 {
			w.activateMoveChoice(globalClickPos)
			return
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

func (w *choiceWindowNode) activateMoveChoice(pos gmath.Vec) {
	if !w.Enabled {
		w.scene.Audio().PlaySound(assets.AudioError)
		return
	}
	w.selectedIndex = -1
	choice := selectedChoice{
		Option: choiceOption{special: specialChoiceMoveColony},
		Pos:    pos,
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

	w.selectedIndex = i
	w.selectedSlide = 0

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
