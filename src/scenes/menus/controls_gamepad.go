package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type ControlsGamepadMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewControlsGamepadMenuController(state *session.State) *ControlsGamepadMenuController {
	return &ControlsGamepadMenuController{state: state}
}

func (c *ControlsGamepadMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *ControlsGamepadMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ControlsGamepadMenuController) initUI() {
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.title")+" -> "+d.Get("menu.main.settings")+" -> "+d.Get("menu.options.controls")+" -> "+d.Get("menu.controls.gamepad"), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	controlsText := d.Get("menu.controls.gamepad.text")

	normalContainer := eui.NewAnchorContainer()
	label := eui.NewLabel(controlsText, smallFont)
	normalContainer.AddChild(label)
	rowContainer.AddChild(normalContainer)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Back", func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ControlsGamepadMenuController) back() {
	c.scene.Context().ChangeScene(NewControlsMenuController(c.state))
}
