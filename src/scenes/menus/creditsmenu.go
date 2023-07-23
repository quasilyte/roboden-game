package menus

import (
	"sort"
	"strings"

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
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *CreditsMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Context().Dict

	smallFont := assets.BitmapFont1

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.credits"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	panel := eui.NewTextPanel(uiResources, 0, 0)
	rowContainer.AddChild(panel)

	testers := []string{
		"bontequero",
		"yukki",
		"NKMory",
		"BaBuwkaPride",
	}
	sort.Strings(testers)

	lines := []string{
		"[" + d.Get("menu.credits.crew") + "]",
		"    quasilyte (Iskander senpai) - game maker",
		"    shooQrow (Oleg) - graphics, co-game design, testing",
		"    " + strings.Join(testers, ", ") + " - testing",
		"",
		"[" + d.Get("menu.credits.assets") + "]",
		"    DROZERiX - Crush, War Path and Sexxxy Bit 3 music tracks",
		"    JAM - Deadly Windmills music track",
		"    unTied Games - super pixel effects packs (1, 2 & 3)",
		"",
		"[" + d.Get("menu.credits.special_thanks") + "]",
		"    Hajime Hoshi - Ebitengine creator and maintainer (@hajimehoshi)",
		"    Mark Carpenter - ebitenui maintainer (@mcarpenter622)",
		"",
		d.Get("menu.credits.thank_player"),
	}

	label := eui.NewLabel(strings.Join(lines, "\n"), smallFont)
	panel.AddChild(label)

	secretScreen := eui.NewButton(uiResources, c.scene, "???", func() {
		c.back()
	})
	rowContainer.AddChild(secretScreen)
	secretScreen.GetWidget().Disabled = true

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *CreditsMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
