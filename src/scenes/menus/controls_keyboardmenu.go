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
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *ControlsKeyboardMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	var buttons []eui.Widget

	// TODO: use an adaptive font here as well? (e.g. state.Resources.Font1)
	smallFont := assets.Font1

	options := &c.state.Persistent.Settings

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.controls")+" -> "+d.Get("menu.controls.keyboard"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	panel := eui.NewTextPanel(uiResources, 0, 0)
	rowContainer.AddChild(panel)

	controlsText := d.Get("menu.controls.keyboard.text")
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

	wheelScrollSelect := eui.NewSelectButton(eui.SelectButtonConfig{
		Resources: uiResources,
		Input:     c.state.MenuInput,
		Value:     &options.WheelScrollingMode,
		Label:     d.Get("menu.controls.wheel_scroll"),
		ValueNames: []string{
			d.Get("menu.controls.wheel_scroll.drag"),
			d.Get("menu.controls.wheel_scroll.float"),
		},
	})
	c.scene.AddObject(wheelScrollSelect)
	rowContainer.AddChild(wheelScrollSelect.Widget)
	buttons = append(buttons, wheelScrollSelect.Widget)

	backButton := eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	})
	rowContainer.AddChild(backButton)
	buttons = append(buttons, backButton)

	navTree := createSimpleNavTree(buttons)
	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *ControlsKeyboardMenuController) back() {
	c.scene.Context().ChangeScene(NewControlsMenuController(c.state))
}
