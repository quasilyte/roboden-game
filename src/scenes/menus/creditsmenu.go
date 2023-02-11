package menus

import (
	"sort"
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type CreditsMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewCreditsMenuController(state *session.State) *CreditsMenuController {
	return &CreditsMenuController{state: state}
}

func (c *CreditsMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *CreditsMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *CreditsMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer()
	root.AddChild(rowContainer)

	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	titleLabel := eui.NewLabel(uiResources, "Credits", smallFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	testers := []string{
		"bontequero",
		"kreker",
		"NKMory",
		"BaBuwkaPride",
	}
	sort.Strings(testers)

	lines := []string{
		"A game by Iskander & Oleg",
		"quasilyte - coding, game design, testing",
		"shooQrow - graphics, testing",
		strings.Join(testers, ", ") + " - testing",
	}

	for _, l := range lines {
		label := eui.NewLabel(uiResources, l, smallFont)
		rowContainer.AddChild(label)
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewLabel(uiResources, "Made with Ebitengine", smallFont))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Back", func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *CreditsMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
