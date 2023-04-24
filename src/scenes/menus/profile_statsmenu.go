package menus

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
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
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	helpLabel := eui.NewLabel("", tinyFont)
	helpLabel.MaxWidth = 340

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.title")+" -> "+d.Get("menu.main.profile")+" -> "+d.Get("menu.profile.stats"), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face
	stats := c.state.Persistent.PlayerStats

	grid := eui.NewGridContainer(2, widget.GridLayoutOpts.Spacing(24, 4),
		widget.GridLayoutOpts.Stretch([]bool{true, false}, nil))
	lines := [][2]string{
		{d.Get("menu.results.time_played"), fmt.Sprintf("%v", timeutil.FormatDuration(d, stats.TotalPlayTime))},
		{d.Get("menu.profile.stats.totalscore"), fmt.Sprintf("%v", stats.TotalScore)},
		{d.Get("menu.profile.stats.classic_highscore"), fmt.Sprintf("%v (%d%%)", stats.HighestClassicScore, stats.HighestClassicScoreDifficulty)},
	}
	if stats.TotalScore >= gamedata.ArenaModeCost {
		lines = append(lines, [2]string{d.Get("menu.profile.stats.arena_highscore"), fmt.Sprintf("%v (%d%%)", stats.HighestArenaScore, stats.HighestArenaScoreDifficulty)})
		lines = append(lines, [2]string{d.Get("menu.profile.stats.inf_arena_highscore"), fmt.Sprintf("%v (%d%%)", stats.HighestInfArenaScore, stats.HighestInfArenaScoreDifficulty)})
	}
	for _, pair := range lines {
		grid.AddChild(eui.NewLabel(pair[0], smallFont))
		grid.AddChild(eui.NewLabel(pair[1], smallFont))
	}
	rowContainer.AddChild(grid)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ProfileStatsMenuController) back() {
	c.scene.Context().ChangeScene(NewProfileMenuController(c.state))
}
