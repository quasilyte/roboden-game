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
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer()
	root.AddChild(rowContainer)

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	titleLabel := eui.NewLabel(uiResources, "Main Menu -> Comtrols", normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	lines := []string{
		"[Pan Camera]",
		"    Mouse: middle button + drag, edge scroll",
		"    Keyboard: arrow keys",
		"[Move Colony]",
		"    Mouse: right mouse button click on the destination",
		"[Select Colony]",
		"    Mouse: left mouse button click on the colony",
		"    Keyboard: tab to switch between the colonies",
		"[Choice Select]",
		"    Mouse: left mouse button click on the option",
		"    Keyboard: 1, 2, 3, 4, 5, q, w, e, r, t",
		"[Exit/Back]",
		"    Keyboard: escape",
	}

	// for _, l := range lines {
	// 	label := eui.NewLabel(uiResources, l, smallFont)
	// 	rowContainer.AddChild(label)
	// }

	normalContainer := eui.NewAnchorContainer()
	label := eui.NewLabel(uiResources, strings.Join(lines, "\n"), smallFont)
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

func (c *ControlsMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
