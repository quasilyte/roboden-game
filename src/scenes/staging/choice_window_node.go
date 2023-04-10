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
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui"
	"github.com/quasilyte/roboden-game/session"
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
	Faction gamedata.FactionTag
	Option  choiceOption
	Pos     gmath.Vec
}

type choiceOptionSlot struct {
	flipAnim   *ge.Animation
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
		cost:    25,
		icon:    assets.ImageActionBuildColony,
	},
	specialBuildGunpoint: {
		text:    "build_gunpoint",
		special: specialBuildGunpoint,
		cost:    10,
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

	choices []*choiceOptionSlot

	shuffledOptions []choiceOption

	beforeSpecialShuffle int
	buildTurret          bool
	specialChoiceKinds   []specialChoiceKind
	specialChoices       []choiceOption

	config *session.LevelConfig
	world  *worldState

	cursor *gameui.CursorNode

	EventChoiceSelected gsignal.Event[selectedChoice]
}

func newChoiceWindowNode(pos gmath.Vec, world *worldState, h *input.Handler, cursor *gameui.CursorNode) *choiceWindowNode {
	return &choiceWindowNode{
		pos:           pos,
		input:         h,
		cursor:        cursor,
		selectedIndex: -1,
		config:        world.config,
		world:         world,
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
	}

	if w.config.AttackActionAvailable {
		w.specialChoiceKinds = append(w.specialChoiceKinds, specialAttack)
	}
	if w.config.RadiusActionAvailable {
		w.specialChoiceKinds = append(w.specialChoiceKinds,
			specialDecreaseRadius,
			specialIncreaseRadius)
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

	camera := w.selectedColony.world.camera
	floppies := [...]resource.ImageID{
		assets.ImageFloppyYellow,
		assets.ImageFloppyRed,
		assets.ImageFloppyGreen,
		assets.ImageFloppyBlue,
		assets.ImageFloppyGray,
	}
	flipSprites := [...]resource.ImageID{
		assets.ImageFloppyYellowFlip,
		assets.ImageFloppyRedFlip,
		assets.ImageFloppyGreenFlip,
		assets.ImageFloppyBlueFlip,
		assets.ImageFloppyGrayFlip,
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
		scene.AddGraphicsAbove(floppy, 1)

		flipSprite := scene.NewSprite(flipSprites[i])
		flipSprite.Centered = false
		flipSprite.Pos.Offset = offset
		flipSprite.Visible = false
		scene.AddGraphicsAbove(flipSprite, 1)

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

		scene.AddGraphicsAbove(darkLabel, 1)
		scene.AddGraphicsAbove(lightLabel, 1)

		var icon *ge.Sprite
		if i == 4 {
			icon = ge.NewSprite(scene.Context())
			icon.Centered = false
			icon.Pos.Base = &floppy.Pos.Offset
			icon.Pos.Offset = gmath.Vec{X: 5, Y: 26}
			scene.AddGraphicsAbove(icon, 1)
		}

		choice := &choiceOptionSlot{
			flipAnim:   ge.NewAnimation(flipSprite, -1),
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
		// Hide it behind the camera before Update() starts to drag it in.
		floppy.Pos.Offset.X += 640

		w.choices[i] = choice
	}
}

func (w *choiceWindowNode) IsDisposed() bool {
	return false
}

func (w *choiceWindowNode) revealChoices() {
	for _, o := range w.choices {
		o.floppy.Pos.Offset.X = w.floppyOffsetX
		o.floppy.Visible = true
		o.labelDark.Visible = true
		o.labelLight.Visible = true
		o.flipAnim.Sprite().Visible = false
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
		w.buildTurret = !w.buildTurret
		gmath.Shuffle(w.scene.Rand(), w.specialChoiceKinds)
		w.beforeSpecialShuffle = len(w.specialChoiceKinds)
	}
	w.beforeSpecialShuffle--
	specialIndex := w.beforeSpecialShuffle

	specialOptionKind := w.specialChoiceKinds[specialIndex]
	if specialOptionKind == specialBuildColony {
		if w.buildTurret && w.config.BuildTurretActionAvailable {
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
	w.value = 0
	w.targetValue = targetValue
	for i, o := range w.choices {
		if i == w.selectedIndex {
			continue
		}
		o.flipAnim.Rewind()
		o.flipAnim.Sprite().Visible = true
		o.floppy.Visible = false
		o.labelDark.Visible = false
		o.labelLight.Visible = false
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

	percentage := w.value / w.targetValue
	const maxSlideOffset float64 = 144 + 8
	for i, o := range w.choices {
		if i == w.selectedIndex {
			o.floppy.Pos.Offset.X = math.Round(w.floppyOffsetX + maxSlideOffset*(1.05*percentage))
			continue
		}

		if o.flipAnim.Tick(delta) {
			o.flipAnim.Sprite().Visible = false
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
	if w.world.movementEnabled {
		if pos, ok := w.cursor.ClickPos(controls.ActionMoveChoice); ok {
			globalClickPos := pos.Add(w.selectedColony.world.camera.Offset)
			if globalClickPos.DistanceTo(w.selectedColony.pos) > 28 {
				w.activateMoveChoice(globalClickPos)
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
	w.startCharging(8.0)
	w.EventChoiceSelected.Emit(choice)
}

func (w *choiceWindowNode) activateChoice(i int) {
	if !w.Enabled {
		w.scene.Audio().PlaySound(assets.AudioError)
		return
	}

	w.selectedIndex = i
	w.selectedSlide = 0

	selectedFaction := gamedata.FactionTag(i + 1)
	choice := selectedChoice{
		Faction: selectedFaction,
		Option:  w.choices[i].option,
	}

	if i == 4 {
		// Special action selected.
		w.startCharging(w.choices[i].option.cost)
	} else {
		w.startCharging(10.0)
	}

	w.EventChoiceSelected.Emit(choice)
}
