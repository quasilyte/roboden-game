package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type OptionsGraphicsMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewOptionsGraphicsMenuController(state *session.State) *OptionsGraphicsMenuController {
	return &OptionsGraphicsMenuController{state: state}
}

func (c *OptionsGraphicsMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *OptionsGraphicsMenuController) Update(delta float64) {
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *OptionsGraphicsMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	normalFont := assets.BitmapFont3

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.graphics"), normalFont)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	{
		rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Value:     &options.Graphics.ShadowsEnabled,
			Label:     d.Get("menu.options.graphics.shadows"),
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
			Value:     &options.Graphics.AllShadersEnabled,
			Label:     d.Get("menu.options.graphics.shaders"),
			ValueNames: []string{
				d.Get("menu.option.mandatory"),
				d.Get("menu.option.all"),
			},
		}))
	}

	{
		rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Value:     &options.Graphics.FullscreenEnabled,
			Label:     d.Get("menu.options.graphics.fullscreen"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
			OnPressed: func() {
				ebiten.SetFullscreen(options.Graphics.FullscreenEnabled)
			},
		}))
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *OptionsGraphicsMenuController) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
