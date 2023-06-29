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

type ExtraCreditsMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewExtraCreditsMenuController(state *session.State) *ExtraCreditsMenuController {
	return &ExtraCreditsMenuController{state: state}
}

func (c *ExtraCreditsMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *ExtraCreditsMenuController) Update(delta float64) {
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ExtraCreditsMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	bigFont := assets.BitmapFont3
	smallFont := assets.BitmapFont1

	d := c.scene.Context().Dict

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.credits")+" -> "+d.Get("menu.more"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	lines := []string{
		"* Hajime Hoshi - Ebitengine creator and maintainer (@hajimehoshi)",
		"* Mark Carpenter - ebitenui maintainer (@mcarpenter622)",
		"* Supportive Ebitengine community",
	}

	normalContainer := eui.NewAnchorContainer()
	label := eui.NewLabel(strings.Join(lines, "\n"), smallFont)
	normalContainer.AddChild(label)
	rowContainer.AddChild(normalContainer)

	rowContainer.AddChild(eui.NewCenteredLabel(d.Get("menu.credits.thank_player"), bigFont))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewCenteredLabel("Made with Ebitengine", smallFont))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ExtraCreditsMenuController) back() {
	c.scene.Context().ChangeScene(NewCreditsMenuController(c.state))
}
