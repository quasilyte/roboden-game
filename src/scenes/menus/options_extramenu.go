package menus

import (
	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/assets"
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
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *OptionsExtraMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.extra"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	{
		rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Value:     &options.Demo,
			Label:     d.Get("menu.options.splash_screen"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		}))
	}

	{
		rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Value:     &options.ShowFPS,
			Label:     d.Get("menu.options.show_fps"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		}))
	}

	{
		rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Value:     &options.ShowTimer,
			Label:     d.Get("menu.options.show_timer"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		}))
	}

	rowContainer.AddChild(eui.NewTransparentSeparator())

	if !c.state.Device.IsMobile {
		rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.terminal"), func() {
			c.scene.Context().ChangeScene(NewTerminalMenuController(c.state))
		}))
	}

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *OptionsExtraMenuController) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
