package menus

import (
	"strings"

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
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.title")+" -> "+d.Get("menu.main.settings")+" -> "+d.Get("menu.options.controls")+" -> "+d.Get("menu.controls.keyboard"), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	controlsText := d.Get("menu.controls.keyboard.text")
	grid := eui.NewGridContainer(2, widget.GridLayoutOpts.Spacing(24, 4))

	for _, line := range strings.Split(controlsText, "\n") {
		left, right, _ := strings.Cut(line, " | ")
		leftLabel := eui.NewLabel(left, smallFont)
		grid.AddChild(leftLabel)
		rightLabel := eui.NewLabel(right, smallFont)
		grid.AddChild(rightLabel)
	}
	rowContainer.AddChild(grid)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ControlsKeyboardMenuController) back() {
	c.scene.Context().ChangeScene(NewControlsMenuController(c.state))
}
