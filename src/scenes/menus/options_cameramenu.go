package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type OptionsCameraMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewOptionsCameraMenuController(state *session.State) *OptionsCameraMenuController {
	return &OptionsCameraMenuController{state: state}
}

func (c *OptionsCameraMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *OptionsCameraMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *OptionsCameraMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.camera"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	{
		rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
			Scene:      c.scene,
			Resources:  uiResources,
			Input:      c.state.MainInput,
			Value:      &options.ScrollingSpeed,
			Label:      d.Get("menu.options.scroll_speed"),
			ValueNames: []string{"1", "2", "3", "4", "5"},
		}))
	}

	{
		rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
			Scene:      c.scene,
			Resources:  uiResources,
			Input:      c.state.MainInput,
			Value:      &options.EdgeScrollRange,
			Label:      d.Get("menu.options.edge_scroll_range"),
			ValueNames: []string{"0", "1", "2", "3", "4"},
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

func (c *OptionsCameraMenuController) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
