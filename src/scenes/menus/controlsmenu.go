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

	var buttons []eui.Widget

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.controls"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	{
		b := eui.NewButton(uiResources, c.scene, d.Get("menu.controls.keyboard"), func() {
			c.scene.Context().ChangeScene(NewControlsKeyboardMenuController(c.state))
		})
		rowContainer.AddChild(b)
		buttons = append(buttons, b)
	}

	{
		b := eui.NewButton(uiResources, c.scene, d.Get("menu.controls.gamepad")+" 1", func() {
			c.scene.Context().ChangeScene(NewControlsGamepadMenuController(c.state, 0))
		})
		rowContainer.AddChild(b)
		buttons = append(buttons, b)
	}
	{
		b := eui.NewButton(uiResources, c.scene, d.Get("menu.controls.gamepad")+" 2", func() {
			c.scene.Context().ChangeScene(NewControlsGamepadMenuController(c.state, 1))
		})
		rowContainer.AddChild(b)
		buttons = append(buttons, b)
	}

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

		player1inputSelect := eui.NewSelectButton(eui.SelectButtonConfig{
			Resources:  uiResources,
			Input:      c.state.MenuInput,
			PlaySound:  true,
			Value:      &options.Player1InputMethod,
			ValueNames: inputMethods,
			Label:      d.Get("menu.controls.player_label") + " 1",
		})
		c.scene.AddObject(player1inputSelect)
		rowContainer.AddChild(player1inputSelect.Widget)
		buttons = append(buttons, player1inputSelect.Widget)

		player2inputSelect := eui.NewSelectButton(eui.SelectButtonConfig{
			Resources:  uiResources,
			Input:      c.state.MenuInput,
			PlaySound:  true,
			Value:      &options.Player2InputMethod,
			ValueNames: inputMethods,
			Label:      d.Get("menu.controls.player_label") + " 2",
		})
		c.scene.AddObject(player2inputSelect)
		rowContainer.AddChild(player2inputSelect.Widget)
		buttons = append(buttons, player2inputSelect.Widget)
	}

	rowContainer.AddChild(eui.NewTransparentSeparator())

	backButton := eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	})
	rowContainer.AddChild(backButton)
	buttons = append(buttons, backButton)

	navTree := createSimpleNavTree(buttons)
	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *ControlsMenuController) back() {
	c.state.ReloadInputs()
	c.state.SaveGameItem("save.json", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
