package staging

import (
	"image/color"
	"math"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/gameui"
)

type choiceWindowNode struct {
	pos gmath.Vec

	scene *ge.Scene

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
	flipAnim   *ge.Animation
	floppy     *ge.Sprite
	icon       *ge.Sprite
	labelLight *ge.Label
	labelDark  *ge.Label
	rect       gmath.Rect
	option     choiceOption
}

func newChoiceWindowNode(pos gmath.Vec, world *worldState, h gameinput.Handler, cursor *gameui.CursorNode) *choiceWindowNode {
	return &choiceWindowNode{
		pos:           pos,
		input:         h,
		cursor:        cursor,
		selectedIndex: -1,
		world:         world,
	}
}

func (w *choiceWindowNode) Init(scene *ge.Scene) {
	w.scene = scene

	camera := w.world.camera
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
		w.world.uiLayer.AddGraphics(floppy)

		flipSprite := scene.NewSprite(flipSprites[i])
		flipSprite.Centered = false
		flipSprite.Pos.Offset = offset
		flipSprite.Visible = false
		w.world.uiLayer.AddGraphics(flipSprite)

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

		w.world.uiLayer.AddGraphics(darkLabel)
		w.world.uiLayer.AddGraphics(lightLabel)

		var icon *ge.Sprite
		if i == 4 {
			icon = ge.NewSprite(scene.Context())
			icon.Centered = false
			icon.Pos.Base = &floppy.Pos.Offset
			icon.Pos.Offset = gmath.Vec{X: 5, Y: 26}
			w.world.uiLayer.AddGraphics(icon)
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

func (w *choiceWindowNode) RevealChoices(selection choiceSelection) {
	w.charging = false

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

	for i, o := range selection.cards {
		w.choices[i].option = o
		w.choices[i].labelDark.Text = o.text
		w.choices[i].labelLight.Text = o.text
	}

	w.choices[4].option = selection.special
	w.choices[4].labelDark.Text = selection.special.text
	w.choices[4].labelLight.Text = selection.special.text
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
		o.labelDark.Visible = false
		o.labelLight.Visible = false
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
