package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type ControlsKeyboardMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewControlsKeyboardMenuController(state *session.State) *ControlsKeyboardMenuController {
	return &ControlsKeyboardMenuController{state: state}
}

func (c *ControlsKeyboardMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *ControlsKeyboardMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ControlsKeyboardMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	titleLabel := eui.NewCenteredLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.settings")+" -> "+d.Get("menu.options.controls")+" -> "+d.Get("menu.controls.keyboard"), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	controlsText := d.Get("menu.controls.keyboard.text")

	normalContainer := eui.NewAnchorContainer()
	label := eui.NewLabel(uiResources, controlsText, smallFont)
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

func (c *ControlsKeyboardMenuController) back() {
	c.scene.Context().ChangeScene(NewControlsMenuController(c.state))
}
