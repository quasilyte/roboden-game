package menus

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/serverapi"
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
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ProfileStatsMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(280, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	tinyFont := assets.BitmapFont1

	helpLabel := eui.NewLabel("", tinyFont)
	helpLabel.MaxWidth = 340

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.profile")+" -> "+d.Get("menu.profile.stats"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	smallFont := assets.BitmapFont1
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
	}
	if stats.TotalScore >= gamedata.InfArenaModeCost {
		lines = append(lines, [2]string{d.Get("menu.profile.stats.inf_arena_highscore"), fmt.Sprintf("%v (%d%%)", stats.HighestInfArenaScore, stats.HighestInfArenaScoreDifficulty)})
	}
	if stats.TotalScore >= gamedata.ReverseModeCost {
		lines = append(lines, [2]string{d.Get("menu.profile.stats.reverse_highscore"), fmt.Sprintf("%v (%d%%)", stats.HighestReverseScore, stats.HighestReverseScoreDifficulty)})
	}
	for _, pair := range lines {
		grid.AddChild(eui.NewLabel(pair[0], smallFont))
		grid.AddChild(eui.NewLabel(pair[1], smallFont))
	}
	rowContainer.AddChild(grid)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	var sendScoreButton *widget.Button
	sendScoreButton = eui.NewButton(uiResources, c.scene, d.Get("menu.publish_high_score"), func() {
		if c.state.Persistent.PlayerName == "" {
			backController := NewProfileStatsMenuController(c.state)
			userNameScene := c.state.SceneRegistry.UserNameMenu(backController)
			c.scene.Context().ChangeScene(userNameScene)
			return
		}
		c.state.SentHighscores = true
		sendScoreButton.GetWidget().Disabled = true
		replays := c.prepareHighscoreReplays()
		if len(replays) != 0 {
			backController := NewProfileStatsMenuController(c.state)
			submitController := c.state.SceneRegistry.SubmitScreen(backController, replays)
			c.scene.Context().ChangeScene(submitController)
			return
		}
	})
	rowContainer.AddChild(sendScoreButton)
	sendScoreButton.GetWidget().Disabled = c.state.SentHighscores ||
		(c.state.Persistent.PlayerStats.HighestClassicScore == 0 &&
			c.state.Persistent.PlayerStats.HighestArenaScore == 0 &&
			c.state.Persistent.PlayerStats.HighestInfArenaScore == 0 &&
			c.state.Persistent.PlayerStats.HighestReverseScore == 0)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ProfileStatsMenuController) prepareHighscoreReplays() []serverapi.GameReplay {
	keys := []string{
		"classic_highscore",
		"arena_highscore",
		"inf_arena_highscore",
		"reverse_highscore",
	}
	var replays []serverapi.GameReplay
	for _, key := range keys {
		if !c.scene.Context().CheckGameData(key) {
			continue
		}
		var replay serverapi.GameReplay
		if err := c.scene.Context().LoadGameData(key, &replay); err != nil {
			fmt.Printf("load %q highscore data: %v\n", key, err)
			continue
		}
		if gamedata.IsSendableReplay(replay) && gamedata.IsValidReplay(replay) {
			replays = append(replays, replay)
		}
	}
	return replays
}

func (c *ProfileStatsMenuController) back() {
	c.scene.Context().ChangeScene(NewProfileMenuController(c.state))
}
