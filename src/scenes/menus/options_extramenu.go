package menus

import (
	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type OptionsExtraMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewOptionsExtraMenuController(state *session.State) *OptionsExtraMenuController {
	return &OptionsExtraMenuController{state: state}
}

func (c *OptionsExtraMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *OptionsExtraMenuController) Update(delta float64) {
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *OptionsExtraMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.extra"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	var buttons []eui.Widget

	{
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound: true,
			Resources: uiResources,
			Input:     c.state.MenuInput,
			BoolValue: &options.Demo,
			Label:     d.Get("menu.options.splash_screen"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		})
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
		buttons = append(buttons, b.Widget)
	}

	if !c.state.Device.IsMobile() {
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound: true,
			Resources: uiResources,
			Input:     c.state.MenuInput,
			BoolValue: &options.ShowFPS,
			Label:     d.Get("menu.options.show_fps"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		})
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
		buttons = append(buttons, b.Widget)
	}

	if !c.state.Device.IsMobile() {
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound: true,
			Resources: uiResources,
			Input:     c.state.MenuInput,
			BoolValue: &options.ShowTimer,
			Label:     d.Get("menu.options.show_timer"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		})
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
		buttons = append(buttons, b.Widget)
	}

	rowContainer.AddChild(eui.NewTransparentSeparator())

	terminalButton := eui.NewButton(uiResources, c.scene, d.Get("menu.terminal"), func() {
		c.scene.Context().ChangeScene(NewTerminalMenuController(c.state))
	})
	rowContainer.AddChild(terminalButton)
	buttons = append(buttons, terminalButton)

	backButton := eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	})
	rowContainer.AddChild(backButton)
	buttons = append(buttons, backButton)

	navTree := createSimpleNavTree(buttons)
	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *OptionsExtraMenuController) back() {
	c.state.SaveGameItem("save.json", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
