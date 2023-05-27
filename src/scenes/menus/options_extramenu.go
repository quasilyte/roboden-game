package menus

import (
	"github.com/ebitenui/ebitenui/widget"
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
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *OptionsExtraMenuController) initUI() {
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face

	d := c.scene.Dict()
	titleLabel := eui.NewLabel(d.Get("menu.main.title")+" -> "+d.Get("menu.main.settings")+" -> "+d.Get("menu.options.extra"), normalFont)
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

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	if !c.state.Device.IsMobile {
		rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.terminal"), func() {
			c.scene.Context().ChangeScene(NewTerminalMenuController(c.state))
		}))

		rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.set_user_name"), func() {
			c.scene.Context().ChangeScene(NewUserNameMenuController(c.state, c))
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
