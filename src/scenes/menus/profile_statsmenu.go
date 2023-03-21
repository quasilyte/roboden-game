package menus

import (
	"fmt"
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/timeutil"
)

type ProfileStatsMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewProfileStatsMenuController(state *session.State) *ProfileStatsMenuController {
	return &ProfileStatsMenuController{state: state}
}

func (c *ProfileStatsMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *ProfileStatsMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ProfileStatsMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	helpLabel := eui.NewLabel(uiResources, "", tinyFont)
	helpLabel.MaxWidth = 340

	titleLabel := eui.NewCenteredLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.profile")+" -> "+d.Get("menu.profile.stats"), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face
	stats := c.state.Persistent.PlayerStats
	lines := []string{
		fmt.Sprintf("%s: %v", d.Get("menu.results.time_played"), timeutil.FormatDuration(d, stats.TotalPlayTime)),
		fmt.Sprintf("%s: %v (%d%%)", d.Get("menu.profile.stats.highscore"), stats.HighestScore, stats.HighestScoreDifficulty),
	}
	rowContainer.AddChild(eui.NewCenteredLabel(uiResources, strings.Join(lines, "\n"), smallFont))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Back", func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ProfileStatsMenuController) back() {
	c.scene.Context().ChangeScene(NewProfileMenuController(c.state))
}
