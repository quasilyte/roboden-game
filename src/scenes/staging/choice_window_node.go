package staging

import (
	"math"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/gameui"
	"github.com/quasilyte/roboden-game/viewport"
)

type choiceWindowNode struct {
	scene *ge.Scene

	cam *viewport.Camera

	input gameinput.Handler

	Enabled bool

	charging    bool
	targetValue float64
	value       float64

	floppyOffsetX float64
	selectedIndex int

	choices []*choiceOptionSlot

	world *worldState

	cursor *gameui.CursorNode
}

type choiceOptionSlot struct {
	flipAnim *ge.Animation
	floppy   *ge.Sprite
	icon     *ge.Sprite
	label1   *ge.Sprite
	label2   *ge.Sprite
	rect     gmath.Rect
	option   choiceOption
}

func newChoiceWindowNode(cam *viewport.Camera, world *worldState, h gameinput.Handler, cursor *gameui.CursorNode) *choiceWindowNode {
	return &choiceWindowNode{
		cam:           cam,
		input:         h,
		cursor:        cursor,
		selectedIndex: -1,
		world:         world,
	}
}

func (w *choiceWindowNode) Init(scene *ge.Scene) {
	w.scene = scene

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
	offsetY := 8.0
	w.floppyOffsetX = (w.cam.Rect.Width() - 86 - 8)
	offset := gmath.Vec{X: w.floppyOffsetX, Y: 8}
	w.choices = make([]*choiceOptionSlot, 5)
	for i := range w.choices {
		floppy := scene.NewSprite(floppies[i])
		floppy.Centered = false
		floppy.Pos.Offset = offset
		w.cam.UI.AddGraphics(floppy)

		flipSprite := scene.NewSprite(flipSprites[i])
		flipSprite.Centered = false
		flipSprite.Pos.Offset = offset
		flipSprite.Visible = false
		w.cam.UI.AddGraphics(flipSprite)

		offset.Y += floppy.ImageHeight() + offsetY

		label1 := scene.NewSprite(assets.ImagePriorityIcons)
		label1.Pos.Base = &floppy.Pos.Offset
		label1.Centered = false
		label1.Visible = false

		label2 := scene.NewSprite(assets.ImagePriorityIcons)
		label2.Pos.Base = &floppy.Pos.Offset
		label2.Centered = false
		label2.Visible = false

		w.cam.UI.AddGraphics(label1)
		w.cam.UI.AddGraphics(label2)

		var icon *ge.Sprite
		if i == 4 {
			icon = ge.NewSprite(scene.Context())
			icon.Centered = false
			icon.Pos.Base = &floppy.Pos.Offset
			icon.Pos.Offset = gmath.Vec{X: 46, Y: 24}
			w.cam.UI.AddGraphics(icon)
		}

		choice := &choiceOptionSlot{
			flipAnim: ge.NewAnimation(flipSprite, -1),
			label1:   label1,
			label2:   label2,
			floppy:   floppy,
			icon:     icon,
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

func (w *choiceWindowNode) RevealChoices(selection choiceSelection) {
	w.charging = false

	for _, o := range w.choices {
		o.floppy.Pos.Offset.X = w.floppyOffsetX
		o.floppy.Visible = true
		o.label1.Visible = false
		o.label2.Visible = false
		o.flipAnim.Sprite().Visible = false
		if o.icon != nil {
			o.icon.Visible = true
		}
	}

	for i, o := range selection.cards {
		faction := gamedata.FactionTag(i + 1)
		choice := w.choices[i]
		choice.option = o
		if len(o.effects) == 1 {
			choice.label1.Visible = true
			choice.label1.Pos.Offset = gmath.Vec{X: 55, Y: 32}
			setPriorityIconFrame(choice.label1, o.effects[0].priority, faction)
		} else {
			choice.label1.Visible = true
			choice.label1.Pos.Offset = gmath.Vec{X: 55, Y: 32 - 10}
			setPriorityIconFrame(choice.label1, o.effects[0].priority, faction)
			choice.label2.Visible = true
			choice.label2.Pos.Offset = gmath.Vec{X: 55, Y: 32 + 10}
			setPriorityIconFrame(choice.label2, o.effects[1].priority, faction)
		}
	}

	w.choices[4].option = selection.special
	w.choices[4].icon.SetImage(w.scene.LoadImage(selection.special.icon))

	w.scene.Audio().PlaySound(assets.AudioChoiceReady)
}

func (w *choiceWindowNode) StartCharging(targetValue float64, cardIndex int) {
	w.charging = true
	w.targetValue = targetValue
	w.value = 0
	w.selectedIndex = cardIndex

	for i, o := range w.choices {
		if i == w.selectedIndex {
			continue
		}
		o.flipAnim.Rewind()
		o.flipAnim.Sprite().Visible = true
		o.floppy.Visible = false
		o.label1.Visible = false
		o.label2.Visible = false
		if o.icon != nil {
			o.icon.Visible = false
		}
	}
}

func (w *choiceWindowNode) Update(delta float64) {
	if !w.charging {
		return
	}
	w.value += delta

	percentage := w.value / w.targetValue
	const maxSlideOffset float64 = 86 + 8
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

func (w *choiceWindowNode) HandleInput() int {
	if pos, ok := w.cursor.ClickPos(controls.ActionClick); ok {
		for i, choice := range w.choices {
			if choice.rect.Contains(pos) {
				return i
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
			return i
		}
	}

	return -1
}
