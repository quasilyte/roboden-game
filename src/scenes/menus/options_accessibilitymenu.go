package menus

import (
	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type OptionsAccessibilityMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewOptionsAccessibilityMenuController(state *session.State) *OptionsAccessibilityMenuController {
	return &OptionsAccessibilityMenuController{state: state}
}

func (c *OptionsAccessibilityMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *OptionsAccessibilityMenuController) Update(delta float64) {
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *OptionsAccessibilityMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.accessibility"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	var buttons []eui.Widget

	largeDiodesSelect := eui.NewSelectButton(eui.SelectButtonConfig{
		PlaySound: true,
		Resources: uiResources,
		Input:     c.state.MenuInput,
		BoolValue: &options.LargeDiodes,
		Label:     d.Get("menu.options.large_diodes"),
		ValueNames: []string{
			d.Get("menu.option.off"),
			d.Get("menu.option.on"),
		},
	})
	c.scene.AddObject(largeDiodesSelect)
	rowContainer.AddChild(largeDiodesSelect.Widget)
	buttons = append(buttons, largeDiodesSelect.Widget)

	largerFontSelect := eui.NewSelectButton(eui.SelectButtonConfig{
		PlaySound: true,
		Resources: uiResources,
		Input:     c.state.MenuInput,
		BoolValue: &options.LargerFont,
		Label:     d.Get("menu.options.larger_font"),
		ValueNames: []string{
			d.Get("menu.option.off"),
			d.Get("menu.option.on"),
		},
		OnPressed: func() {
			c.state.AdjustTextSize(options.LargerFont)
		},
	})
	c.scene.AddObject(largerFontSelect)
	rowContainer.AddChild(largerFontSelect.Widget)
	buttons = append(buttons, largerFontSelect.Widget)

	rowContainer.AddChild(eui.NewTransparentSeparator())

	backButton := eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	})
	rowContainer.AddChild(backButton)
	buttons = append(buttons, backButton)

	navTree := createSimpleNavTree(buttons)
	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *OptionsAccessibilityMenuController) back() {
	c.state.SaveGameItem("save.json", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
