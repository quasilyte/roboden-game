package menus

import (
	"fmt"
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type ControlsGamepadMenuController struct {
	state *session.State

	id int

	updateDelay float64

	statusText *widget.Text
	leftRadar  *widget.Graphic
	rightRadar *widget.Graphic
	leftStick  *ge.Sprite
	rightStick *ge.Sprite

	scene *ge.Scene
}

func NewControlsGamepadMenuController(state *session.State, id int) *ControlsGamepadMenuController {
	return &ControlsGamepadMenuController{
		id:    id,
		state: state,
	}
}

func (c *ControlsGamepadMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
	c.updateDelay = 0.4
}

func (c *ControlsGamepadMenuController) Update(delta float64) {
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}

	c.updateDelay = gmath.ClampMin(c.updateDelay-delta, 0)
	if c.updateDelay == 0 {
		c.updateDelay = 0.05
		c.checkGamepad()
	}
}

func (c *ControlsGamepadMenuController) checkGamepad() {
	var h *gameinput.Handler
	if c.id == 0 {
		h = &c.state.FirstGamepadInput
	} else {
		h = &c.state.SecondGamepadInput
	}

	d := c.scene.Dict()

	if !h.GamepadConnected() {
		c.statusText.Label = fmt.Sprintf("%s: %s", d.Get("menu.controls.gamepad_status"), d.Get("menu.controls.gamepad_status.not_connected"))
		if c.leftStick != nil {
			c.leftStick.Dispose()
			c.rightStick.Dispose()
			c.leftStick = nil
			c.rightStick = nil
		}
		return
	}

	if c.leftStick == nil {
		leftRadarRect := c.leftRadar.GetWidget().Rect
		rightRadarRect := c.rightRadar.GetWidget().Rect
		dotOffset := gmath.Vec{
			X: float64(leftRadarRect.Dx() / 2),
			Y: float64(leftRadarRect.Dy() / 2),
		}

		leftStickPos := dotOffset
		leftStickPos.X += float64(leftRadarRect.Min.X)
		leftStickPos.Y += float64(leftRadarRect.Min.Y)
		c.leftStick = c.scene.NewSprite(assets.ImageUIGamepadRadarDot)
		c.leftStick.Pos.Base = &leftStickPos
		c.scene.AddGraphicsAbove(c.leftStick, 1)

		rightStickPos := dotOffset
		rightStickPos.X += float64(rightRadarRect.Min.X)
		rightStickPos.Y += float64(rightRadarRect.Min.Y)
		c.rightStick = c.scene.NewSprite(assets.ImageUIGamepadRadarDot)
		c.rightStick.Pos.Base = &rightStickPos
		c.scene.AddGraphicsAbove(c.rightStick, 1)
	}

	const radarScale = 40
	c.leftStick.Pos.Offset = gmath.Vec{}
	c.rightStick.Pos.Offset = gmath.Vec{}
	if info, ok := h.PressedActionInfo(controls.ActionTestLeftStick); ok {
		c.leftStick.Pos.Offset = info.Pos.Mulf(radarScale)
	}
	if info, ok := h.PressedActionInfo(controls.ActionMoveCursor); ok {
		c.rightStick.Pos.Offset = info.Pos.Mulf(radarScale)
	}

	c.statusText.Label = fmt.Sprintf("%s: %s\n\n%s",
		d.Get("menu.controls.gamepad_status"),
		d.Get("menu.controls.gamepad_status.connected"),
		d.Get("menu.controls.gamepad.calibrate"))
}

func (c *ControlsGamepadMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	smallFont := assets.BitmapFont1

	options := &c.state.Persistent.Settings

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.controls")+" -> "+d.Get("menu.controls.gamepad")+fmt.Sprintf(" %d", c.id+1), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	panelsPairContainer := eui.NewGridContainer(2, widget.GridLayoutOpts.Spacing(8, 4),
		widget.GridLayoutOpts.Stretch(nil, nil))

	calibrationPanel := eui.NewTextPanel(uiResources, 0, 0)
	panelsPairContainer.AddChild(calibrationPanel)

	panel := eui.NewTextPanel(uiResources, 0, 0)
	panelsPairContainer.AddChild(panel)

	rowContainer.AddChild(panelsPairContainer)

	h := c.state.GetInput(c.id)

	grid := eui.NewGridContainer(2, widget.GridLayoutOpts.Spacing(24, 4),
		widget.GridLayoutOpts.Stretch([]bool{true, false}, nil))

	initControlsGrid := func(s string) {
		for _, line := range strings.Split(s, "\n") {
			left, right, _ := strings.Cut(line, " | ")
			leftLabel := eui.NewLabel(left, smallFont)
			grid.AddChild(leftLabel)
			rightLabel := eui.NewLabel(right, smallFont)
			grid.AddChild(rightLabel)
		}
	}

	initControlsGrid(h.ReplaceKeyNames(d.Get("menu.controls.gamepad.text")))
	panel.AddChild(grid)

	rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
		Resources:  uiResources,
		Input:      c.state.CombinedInput,
		Value:      &options.GamepadSettings[c.id].Layout,
		Label:      d.Get("menu.controls.gamepad_layout"),
		ValueNames: []string{"Xbox", "PlayStation", "Nintendo Switch"},
		OnPressed: func() {
			h.SetGamepadLayout(gameinput.GamepadLayoutKind(options.GamepadSettings[c.id].Layout))
			grid.RemoveChildren()
			initControlsGrid(h.ReplaceKeyNames(d.Get("menu.controls.gamepad.text")))

			// TODO: update bindings text.
		},
	}))

	rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
		Resources:  uiResources,
		Input:      c.state.CombinedInput,
		Value:      &options.GamepadSettings[c.id].CursorSpeed,
		Label:      d.Get("menu.controls.gamepad_cursor_speed"),
		ValueNames: []string{"-80", "-50%", "-20%", "+0%", "+20%", "+50%", "+80%", "+100%"},
		OnPressed: func() {
			h.SetVirtualCursorSpeed(options.GamepadSettings[c.id].CursorSpeed)
		},
	}))

	rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
		Resources:  uiResources,
		Input:      c.state.CombinedInput,
		Value:      &options.GamepadSettings[c.id].DeadzoneLevel,
		Label:      d.Get("menu.controls.gamepad_deadzone"),
		ValueNames: []string{"0.05", "0.10", "0.15", "0.20", "0.25", "0.30", "0.35", "0.40", "0.45", "0.50", "0.55", "0.60"},
		OnPressed: func() {
			h.SetGamepadDeadzoneLevel(options.GamepadSettings[c.id].DeadzoneLevel)
		},
	}))

	{
		calibrationContentsContainer := eui.NewRowLayoutContainer(10, nil)

		radarsContainer := eui.NewGridContainer(2, widget.GridLayoutOpts.Spacing(10, 4),
			widget.GridLayoutOpts.Stretch([]bool{false, false, true}, nil))
		calibrationContentsContainer.AddChild(radarsContainer)

		leftRadarWidget := widget.NewGraphic(widget.GraphicOpts.Image(c.createRadarImage()))
		c.leftRadar = leftRadarWidget
		radarsContainer.AddChild(leftRadarWidget)

		rightRadarWidget := widget.NewGraphic(widget.GraphicOpts.Image(c.createRadarImage()))
		c.rightRadar = rightRadarWidget
		radarsContainer.AddChild(rightRadarWidget)

		statusText := eui.NewLabel(fmt.Sprintf("%s: %s", d.Get("menu.controls.gamepad_status"), d.Get("menu.controls.gamepad_status.checking")), smallFont)
		statusText.MaxWidth = 240
		calibrationContentsContainer.AddChild(statusText)

		c.statusText = statusText

		calibrationPanel.AddChild(calibrationContentsContainer)
	}

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ControlsGamepadMenuController) createRadarImage() *ebiten.Image {
	return c.scene.LoadImage(assets.ImageUIGamepadRadar).Data
}

func (c *ControlsGamepadMenuController) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewControlsMenuController(c.state))
}
