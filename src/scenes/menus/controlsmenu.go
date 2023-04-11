package menus

import (
	"github.com/ebitenui/ebitenui/widget"
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
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ControlsMenuController) initUI() {
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face

	titleLabel := eui.NewCenteredLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.settings")+" -> "+d.Get("menu.options.controls"), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.controls.keyboard"), func() {
		c.scene.Context().ChangeScene(NewControlsKeyboardMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.controls.gamepad"), func() {
		c.scene.Context().ChangeScene(NewControlsGamepadMenuController(c.state))
	}))

	touchButton := eui.NewButton(uiResources, c.scene, d.Get("menu.controls.touch"), func() {
	})
	touchButton.GetWidget().Disabled = true
	rowContainer.AddChild(touchButton)

	rowContainer.AddChild(eui.NewCenteredLabel(uiResources, d.Get("menu.controls.notice"), normalFont))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ControlsMenuController) back() {
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
