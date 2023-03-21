package staging

import (
	"fmt"
	"strings"
	"time"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/timeutil"
)

type resultsController struct {
	state *session.State

	scene          *ge.Scene
	backController ge.SceneController

	results battleResults
}

type battleResults struct {
	Victory         bool
	TimePlayed      time.Duration
	SurvivingDrones int

	ResourcesGathered      float64
	EliteResourcesGathered float64
	DronesProduced         int
	CreepsDefeated         int

	RedCrystalsCollected int

	Score                int
	DifficultyScore      int
	DronePointsAllocated int
}

func newResultsController(state *session.State, backController ge.SceneController, results battleResults) *resultsController {
	return &resultsController{
		state:          state,
		backController: backController,
		results:        results,
	}
}

func (c *resultsController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()

	stats := &c.state.Persistent.PlayerStats
	stats.TotalPlayTime += c.results.TimePlayed
	if stats.HighestScore < c.results.Score {
		stats.HighestScore = c.results.Score
		stats.HighestScoreDifficulty = c.results.DifficultyScore
	}
	c.scene.Context().SaveGameData("save", c.state.Persistent)
}

func (c *resultsController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *resultsController) checkAchievements() ([]string, []string) {
	var newAchievements []string
	var upgradedAchievements []string

	stats := &c.state.Persistent.PlayerStats

	alreadyAchieved := map[string]int{}
	for _, a := range stats.Achievements {
		level := 1
		if a.Elite {
			level = 2
		}
		alreadyAchieved[a.Name] = level
	}

	difficultyLevel := 1
	if c.results.DifficultyScore >= 150 {
		difficultyLevel = 2
	}

	for _, a := range gamedata.AchievementList {
		if alreadyAchieved[a.Name] >= difficultyLevel {
			continue
		}
		unlocked := false

		switch a.Name {
		case "impossible":
			unlocked = c.results.DifficultyScore > 200
		case "speedrunning":
			unlocked = c.results.TimePlayed.Minutes() < 10
		case "victorydrag":
			unlocked = c.results.TimePlayed.Hours() >= 2
		case "t3engineer":
			// TODO
		case "t3less":
			// TODO
		case "cheapbuild10":
			unlocked = c.results.DronePointsAllocated <= 10
		case "hightension":
			// TODO
		case "solobase":
			// TODO
		case "uiless":
			// TODO
		case "powerof3":
			// TODO
		case "tinyradius":
			// TODO
		case "t1army":
			// TODO
		case "groundwin":
			// TODO
		}

		if !unlocked {
			continue
		}

		elite := difficultyLevel == 2
		if _, ok := alreadyAchieved[a.Name]; ok {
			stats.Achievements = xslices.Remove(stats.Achievements, session.Achievement{
				Name:  a.Name,
				Elite: alreadyAchieved[a.Name] == 2,
			})
			upgradedAchievements = append(upgradedAchievements, a.Name)
		} else {
			newAchievements = append(newAchievements, a.Name)
		}
		stats.Achievements = append(stats.Achievements, session.Achievement{
			Name:  a.Name,
			Elite: elite,
		})
	}

	return newAchievements, upgradedAchievements
}

func (c *resultsController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	d := c.scene.Dict()

	titleString := d.Get("menu.results.defeat")
	if c.results.Victory {
		titleString = d.Get("menu.results.victory") + "!"
	}
	titleLabel := eui.NewCenteredLabel(uiResources, titleString, smallFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	lines := []string{
		fmt.Sprintf("%s: %v", d.Get("menu.results.time_played"), timeutil.FormatDuration(d, c.results.TimePlayed)),
		fmt.Sprintf("%s: %v", d.Get("menu.results.resources_gathered"), int(c.results.ResourcesGathered)),
		fmt.Sprintf("%s: %v", d.Get("menu.results.drone_survivors"), c.results.SurvivingDrones),
		fmt.Sprintf("%s: %v", d.Get("menu.results.drones_total"), c.results.DronesProduced),
		fmt.Sprintf("%s: %v", d.Get("menu.results.creeps_defeated"), c.results.CreepsDefeated),
	}
	if c.results.Victory {
		if c.results.Score > c.state.Persistent.PlayerStats.HighestScore {
			lines = append(lines, fmt.Sprintf("%s: %v (%s)", d.Get("menu.results.score"), c.results.Score, d.Get("menu.results.new_record")))
		} else {
			lines = append(lines, fmt.Sprintf("%s: %v", d.Get("menu.results.score"), c.results.Score))
		}
	}

	newAchievements, upgradedAchievements := c.checkAchievements()
	for _, a := range newAchievements {
		lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.results.new_achievement"), d.Get("achievement", a)))
	}
	for _, a := range upgradedAchievements {
		lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.results.upgraded_achievement"), d.Get("achievement", a)))
	}

	label := eui.NewCenteredLabel(uiResources, strings.Join(lines, "\n"), smallFont)
	rowContainer.AddChild(label)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.lobby_back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *resultsController) back() {
	c.scene.Context().ChangeScene(c.backController)
}
