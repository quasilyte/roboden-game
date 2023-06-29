package menus

import (
	"fmt"
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type ControlsGamepadMenuController struct {
	state *session.State

	id int

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
}

func (c *ControlsGamepadMenuController) Update(delta float64) {
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
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

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	controlsText := d.Get("menu.controls.gamepad.text")
	grid := eui.NewGridContainer(2, widget.GridLayoutOpts.Spacing(24, 4),
		widget.GridLayoutOpts.Stretch([]bool{true, false}, nil))

	for _, line := range strings.Split(controlsText, "\n") {
		left, right, _ := strings.Cut(line, " | ")
		leftLabel := eui.NewLabel(left, smallFont)
		grid.AddChild(leftLabel)
		rightLabel := eui.NewLabel(right, smallFont)
		grid.AddChild(rightLabel)
	}
	rowContainer.AddChild(grid)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
		Resources:  uiResources,
		Input:      c.state.CombinedInput,
		Value:      &options.GamepadSettings[c.id].DeadzoneLevel,
		Label:      d.Get("menu.controls.gamepad_deadzone"),
		ValueNames: []string{"0.05", "0.10", "0.15", "0.20", "0.25", "0.30", "0.35", "0.40"},
		OnPressed: func() {
			c.state.GetInput(c.id).SetGamepadDeadzoneLevel(options.GamepadSettings[c.id].DeadzoneLevel)
		},
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ControlsGamepadMenuController) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewControlsMenuController(c.state))
}
