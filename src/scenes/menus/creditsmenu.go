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
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Context().Dict

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	titleLabel := eui.NewLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.credits"), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	testers := []string{
		"bontequero",
		"yukki",
		"NKMory",
		"BaBuwkaPride",
	}
	sort.Strings(testers)

	lines := []string{
		"[" + d.Get("menu.credits.crew") + "]",
		"    quasilyte (Iskander senpai) - coding, game design, sfx, testing",
		"    shooQrow (Oleg) - graphics, co-game design, testing",
		"    " + strings.Join(testers, ", ") + " - testing",
		"[" + d.Get("menu.credits.assets") + "]",
		"    DROZERiX - Crush and War Path music tracks",
		"    JAM - Deadly Windmills music track",
		"    unTied Games - pixel art explosions free asset pack",
	}

	normalContainer := eui.NewAnchorContainer()
	label := eui.NewLabel(uiResources, strings.Join(lines, "\n"), smallFont)
	normalContainer.AddChild(label)
	rowContainer.AddChild(normalContainer)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.more"), func() {
		c.scene.Context().ChangeScene(NewExtraCreditsMenuController(c.state))
	}))

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
