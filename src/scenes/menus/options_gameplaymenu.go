package menus

import (
	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type OptionsGameplayMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewOptionsGameplayMenuController(state *session.State) *OptionsGameplayMenuController {
	return &OptionsGameplayMenuController{state: state}
}

func (c *OptionsGameplayMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *OptionsGameplayMenuController) Update(delta float64) {
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *OptionsGameplayMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(520, 10, nil)
	root.AddChild(rowContainer)

	var buttons []eui.Widget

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.gameplay"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	{
		hintModeSelect := eui.NewSelectButton(eui.SelectButtonConfig{
			Resources: uiResources,
			Input:     c.state.MenuInput,
			Value:     &options.HintMode,
			Label:     d.Get("menu.options.hint_mode"),
			ValueNames: []string{
				d.Get("menu.option.none"),
				d.Get("menu.option.some"),
				d.Get("menu.option.all"),
			},
		})
		c.scene.AddObject(hintModeSelect)
		rowContainer.AddChild(hintModeSelect.Widget)
		buttons = append(buttons, hintModeSelect.Widget)
	}

	if !c.state.Device.IsMobile() {
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			Resources: uiResources,
			Input:     c.state.MenuInput,
			BoolValue: &options.ScreenButtons,
			Label:     d.Get("menu.options.screen_buttons"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		})
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
		buttons = append(buttons, b.Widget)
	}

	{
		scrollSpeedSelect := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound:  true,
			Resources:  uiResources,
			Input:      c.state.MenuInput,
			Value:      &options.ScrollingSpeed,
			Label:      d.Get("menu.options.scroll_speed"),
			ValueNames: []string{"1", "2", "3", "4", "5"},
		})
		c.scene.AddObject(scrollSpeedSelect)
		rowContainer.AddChild(scrollSpeedSelect.Widget)
		buttons = append(buttons, scrollSpeedSelect.Widget)
	}

	if !c.state.Device.IsMobile() {
		edgeScrollRangeSelect := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound:  true,
			Resources:  uiResources,
			Input:      c.state.MenuInput,
			Value:      &options.EdgeScrollRange,
			Label:      d.Get("menu.options.edge_scroll_range"),
			ValueNames: []string{"0", "1", "2", "3", "4"},
		})
		c.scene.AddObject(edgeScrollRangeSelect)
		rowContainer.AddChild(edgeScrollRangeSelect.Widget)
		buttons = append(buttons, edgeScrollRangeSelect.Widget)
	}

	if !c.state.Device.IsMobile() {
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			Resources: uiResources,
			Input:     c.state.MenuInput,
			BoolValue: &options.NoPauseSpeedToggle,
			Label:     d.Get("menu.options.pause_speed_toggle"),
			ValueNames: []string{
				// It's a reverse (negated) option, so this order makes sense.
				d.Get("menu.option.on"),
				d.Get("menu.option.off"),
			},
		})
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
		buttons = append(buttons, b.Widget)
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

func (c *OptionsGameplayMenuController) back() {
	c.state.SaveGameItem("save.json", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
