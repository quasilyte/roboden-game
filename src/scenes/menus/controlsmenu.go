package menus

import (
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
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *ControlsMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	options := &c.state.Persistent.Settings

	d := c.scene.Dict()

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.controls"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.controls.keyboard"), func() {
		c.scene.Context().ChangeScene(NewControlsKeyboardMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.controls.gamepad")+" 1", func() {
		c.scene.Context().ChangeScene(NewControlsGamepadMenuController(c.state, 0))
	}))
	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.controls.gamepad")+" 2", func() {
		c.scene.Context().ChangeScene(NewControlsGamepadMenuController(c.state, 1))
	}))

	// TODO: show it for mobile devices.
	// touchButton := eui.NewButton(uiResources, c.scene, d.Get("menu.controls.touch"), func() {
	// })
	// touchButton.GetWidget().Disabled = true
	// rowContainer.AddChild(touchButton)

	if !c.state.Device.IsMobile() {
		inputMethods := []string{
			d.Get("menu.controls.method_combined"),
			d.Get("menu.controls.method_keyboard"),
			d.Get("menu.controls.method_gamepad") + " 1",
			d.Get("menu.controls.method_gamepad") + " 2",
		}
		rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
			Resources:  uiResources,
			Input:      c.state.MenuInput,
			Scene:      c.scene,
			Value:      &options.Player1InputMethod,
			ValueNames: inputMethods,
			Label:      d.Get("menu.controls.player_label") + " 1",
		}))
		rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
			Resources:  uiResources,
			Input:      c.state.MenuInput,
			Scene:      c.scene,
			Value:      &options.Player2InputMethod,
			ValueNames: inputMethods,
			Label:      d.Get("menu.controls.player_label") + " 2",
		}))
	}

	rowContainer.AddChild(eui.NewTransparentSeparator())

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ControlsMenuController) back() {
	c.state.ReloadInputs()
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
