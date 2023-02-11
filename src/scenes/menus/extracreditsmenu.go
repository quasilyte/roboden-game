package menus

import (
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
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ExtraCreditsMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer()
	root.AddChild(rowContainer)

	bigFont := c.scene.Context().Loader.LoadFont(assets.FontBig).Face
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	titleLabel := eui.NewLabel(uiResources, "Special Thanks", smallFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	lines := []string{
		"* Hajime Hoshi - Ebitengine creator and maintainer *",
		"* Supportive Ebitengine community *",
	}

	for _, l := range lines {
		label := eui.NewLabel(uiResources, l, smallFont)
		rowContainer.AddChild(label)
	}

	rowContainer.AddChild(eui.NewLabel(uiResources, "And thank you, player", bigFont))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Back", func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ExtraCreditsMenuController) back() {
	c.scene.Context().ChangeScene(NewCreditsMenuController(c.state))
}
