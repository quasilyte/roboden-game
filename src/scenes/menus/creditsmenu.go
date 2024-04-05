package menus

import (
	"sort"
	"strings"
	"time"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
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
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *CreditsMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	var buttons []eui.Widget

	d := c.scene.Context().Dict

	smallFont := c.state.Resources.Font2

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.credits"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	panel := eui.NewTextPanel(uiResources, 640, 92*2)
	rowContainer.AddChild(panel)

	testers := []string{
		"bontequero",
		"yukki",
		"BaBuwkaPride",
	}
	sort.Strings(testers)

	pages := []string{
		strings.Join([]string{
			"[" + d.Get("menu.credits.crew") + "]",
			"  quasilyte (Iskander senpai) - game maker",
			"  shooQrow (Oleg) - graphics, co-game design",
			"  " + strings.Join(testers, ", ") + " - testing",
		}, "\n"),
		strings.Join([]string{
			"[" + d.Get("menu.credits.assets") + "]",
			"  DROZERiX - Crush, War Path, Sexy Bit 3 tracks",
			"  JAM - Deadly Windmills track",
			"  unTied Games - pixel effects packs 1-3",
		}, "\n"),
		strings.Join([]string{
			"[" + d.Get("menu.credits.special_thanks") + "]",
			"  Hajime Hoshi - Ebitengine creator",
			"  Mark Carpenter - ebitenui",
		}, "\n"),
		strings.Join([]string{
			d.Get("menu.credits.thank_player"),
		}, "\n"),
	}

	label := eui.NewLabel(pages[0], smallFont)
	panel.AddChild(label)

	secretScreen := eui.NewButton(uiResources, c.scene, "???", func() {
		c.scene.Context().ChangeScene(NewSecretMenuController(c.state))
	})
	rowContainer.AddChild(secretScreen)
	buttons = append(buttons, secretScreen)
	secretTime := time.Duration(7*time.Hour + 7*time.Minute + 7*time.Second)
	secretScreen.GetWidget().Disabled = c.state.Persistent.PlayerStats.TotalPlayTime < secretTime

	var pageSlider gmath.Slider
	pageSlider.SetBounds(0, len(pages)-1)
	nextButton := eui.NewButton(uiResources, c.scene, d.Get("menu.next"), func() {
		pageSlider.Inc()
		label.Label = pages[pageSlider.Value()]
	})
	rowContainer.AddChild(nextButton)
	buttons = append(buttons, nextButton)

	backButton := eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	})
	rowContainer.AddChild(backButton)
	buttons = append(buttons, backButton)

	navTree := createSimpleNavTree(buttons)
	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *CreditsMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
