package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type ControlsMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewControlsMenuController(state *session.State) *ControlsMenuController {
	return &ControlsMenuController{state: state}
}

func (c *ControlsMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *ControlsMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ControlsMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	smallFont := assets.BitmapFont1

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.controls"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.controls.keyboard"), func() {
		c.scene.Context().ChangeScene(NewControlsKeyboardMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.controls.gamepad"), func() {
		c.scene.Context().ChangeScene(NewControlsGamepadMenuController(c.state))
	}))

	touchButton := eui.NewButton(uiResources, c.scene, d.Get("menu.controls.touch"), func() {
	})
	touchButton.GetWidget().Disabled = true
	rowContainer.AddChild(touchButton)

	rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
		Scene:     c.scene,
		Resources: uiResources,
		Value:     &c.state.Persistent.Settings.SwapGamepads,
		Label:     d.Get("menu.controls_swap_gamepads"),
		ValueNames: []string{
			d.Get("menu.option.off"),
			d.Get("menu.option.on"),
		},
	}))

	rowContainer.AddChild(eui.NewCenteredLabel(d.Get("menu.controls.notice"), smallFont))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ControlsMenuController) back() {
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
