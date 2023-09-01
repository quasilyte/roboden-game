package menus

import (
	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/assets"
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
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionMenuBack) {
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

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.gameplay"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	{
		rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
			Resources: uiResources,
			Input:     c.state.CombinedInput,
			Value:     &options.HintMode,
			Label:     d.Get("menu.options.hint_mode"),
			ValueNames: []string{
				d.Get("menu.option.none"),
				d.Get("menu.option.some"),
				d.Get("menu.option.all"),
			},
		}))
	}

	{
		rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
			Resources: uiResources,
			Value:     &options.ScreenButtons,
			Label:     d.Get("menu.options.screen_buttons"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		}))
	}

	{
		rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
			Scene:      c.scene,
			Resources:  uiResources,
			Input:      c.state.CombinedInput,
			Value:      &options.ScrollingSpeed,
			Label:      d.Get("menu.options.scroll_speed"),
			ValueNames: []string{"1", "2", "3", "4", "5"},
		}))
	}

	{
		rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
			Scene:      c.scene,
			Resources:  uiResources,
			Input:      c.state.CombinedInput,
			Value:      &options.EdgeScrollRange,
			Label:      d.Get("menu.options.edge_scroll_range"),
			ValueNames: []string{"0", "1", "2", "3", "4"},
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

func (c *OptionsGameplayMenuController) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
