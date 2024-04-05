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

type ControlsTouchMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewControlsTouchMenuController(state *session.State) *ControlsTouchMenuController {
	return &ControlsTouchMenuController{state: state}
}

func (c *ControlsTouchMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *ControlsTouchMenuController) Update(delta float64) {
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *ControlsTouchMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	// TODO: use an adaptive font here as well? (e.g. state.Resources.Font1)
	smallFont := assets.Font1

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.controls"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	panel := eui.NewTextPanel(uiResources, 0, 0)
	rowContainer.AddChild(panel)

	controlsText := d.Get("menu.controls.touch.text")
	grid := eui.NewGridContainer(2, widget.GridLayoutOpts.Spacing(24, 4),
		widget.GridLayoutOpts.Stretch([]bool{true, false}, nil))

	for _, line := range strings.Split(controlsText, "\n") {
		left, right, _ := strings.Cut(line, " | ")
		leftLabel := eui.NewLabel(left, smallFont)
		grid.AddChild(leftLabel)
		rightLabel := eui.NewLabel(right, smallFont)
		grid.AddChild(rightLabel)
	}
	panel.AddChild(grid)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ControlsTouchMenuController) back() {
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
