package staging

import (
	"fmt"
	"strings"
	"time"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/contentlock"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/timeutil"
)

type resultsController struct {
	state  *session.State
	config *session.LevelConfig

	scene          *ge.Scene
	backController ge.SceneController

	highScore            bool
	newAchievements      []string
	upgradedAchievements []string
	newDrones            []gamedata.ColonyAgentKind

	results battleResults
}

type battleResults struct {
	Victory          bool
	GroundBossDefeat bool
	TimePlayed       time.Duration

	ResourcesGathered      float64
	EliteResourcesGathered float64
	DronesProduced         int
	DronesCloned           int
	CreepsDefeated         int
	CreepTotalValue        int
	CreepFragScore         int
	CreepBasesDestroyed    int

	DominatorsSurvived int

	T3created       int
	ColoniesBuilt   int
	RadiusIncreases int

	YellowFactionUsed bool
	RedFactionUsed    bool
	GreenFactionUsed  bool
	BlueFactionUsed   bool

	RedCrystalsCollected int

	ArenaLevel           int
	Score                int
	DifficultyScore      int
	DronePointsAllocated int

	OnlyTier1Military bool

	Tier3Drones []gamedata.ColonyAgentKind
}

func newResultsController(state *session.State, config *session.LevelConfig, backController ge.SceneController, results battleResults) *resultsController {
	return &resultsController{
		state:          state,
		backController: backController,
		results:        results,
		config:         config,
	}
}

func (c *resultsController) Init(scene *ge.Scene) {
	c.scene = scene
	victory := c.results.Victory ||
		(c.config.GameMode == gamedata.ModeArena && c.config.InfiniteMode)
	if victory {
		c.updateProgress()
		c.scene.Context().SaveGameData("save", c.state.Persistent)
	}
	c.initUI()
}

func (c *resultsController) updateProgress() {
	stats := &c.state.Persistent.PlayerStats

	stats.TotalPlayTime += c.results.TimePlayed

	if c.config.Tutorial != nil {
		if !xslices.Contains(stats.TutorialsCompleted, c.config.Tutorial.ID) {
			stats.TutorialsCompleted = append(stats.TutorialsCompleted, c.config.Tutorial.ID)
			stats.TotalScore += c.config.Tutorial.ScoreReward
		}
		return
	}

	t3drones := map[gamedata.ColonyAgentKind]struct{}{}
	for _, k := range stats.Tier3DronesSeen {
		t3drones[k] = struct{}{}
	}
	for _, k := range c.results.Tier3Drones {
		if _, ok := t3drones[k]; ok {
			continue
		}
		stats.Tier3DronesSeen = append(stats.Tier3DronesSeen, k)
	}

	stats.TotalScore += c.results.Score
	switch c.config.GameMode {
	case gamedata.ModeClassic:
		if stats.HighestClassicScore < c.results.Score {
			c.highScore = true
			stats.HighestClassicScore = c.results.Score
			stats.HighestClassicScoreDifficulty = c.results.DifficultyScore
		}
	case gamedata.ModeArena:
		if c.config.InfiniteMode {
			if stats.HighestInfArenaScore < c.results.Score {
				c.highScore = true
				stats.HighestInfArenaScore = c.results.Score
				stats.HighestInfArenaScoreDifficulty = c.results.DifficultyScore
			}
		} else {
			if stats.HighestArenaScore < c.results.Score {
				c.highScore = true
				stats.HighestArenaScore = c.results.Score
				stats.HighestArenaScoreDifficulty = c.results.DifficultyScore
			}
		}
	}

	contentUpdates := contentlock.Update(c.state)
	c.newAchievements, c.upgradedAchievements = c.checkAchievements()
	c.newDrones = contentUpdates.DronesUnlocked
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
		if a.Mode != gamedata.ModeAny {
			if a.Mode != c.config.GameMode {
				continue
			}
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
			unlocked = len(stats.Tier3DronesSeen) >= len(gamedata.Tier3agentMergeRecipes)
		case "t3less":
			unlocked = c.results.T3created == 0
		case "cheapbuild10":
			unlocked = c.results.DronePointsAllocated <= 10
		case "hightension":
			unlocked = c.config.WorldSize == 0 && c.results.CreepBasesDestroyed == 0
		case "solobase":
			unlocked = c.results.ColoniesBuilt == 0
		case "uiless":
			unlocked = !c.config.ExtraUI
		case "powerof3":
			unlocked = !c.results.YellowFactionUsed || !c.results.RedFactionUsed || !c.results.GreenFactionUsed || !c.results.BlueFactionUsed
		case "tinyradius":
			unlocked = c.results.RadiusIncreases == 0
		case "t1army":
			unlocked = c.results.OnlyTier1Military
		case "groundwin":
			unlocked = c.results.GroundBossDefeat
		case "turretdamage":
			// TODO

		case "infinite":
			unlocked = c.config.InfiniteMode && c.results.ArenaLevel >= 35
		case "antidominator":
			unlocked = !c.config.InfiniteMode && c.results.DominatorsSurvived == 0
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
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	d := c.scene.Dict()

	var titleString string
	switch {
	case c.results.Victory:
		titleString = d.Get("menu.results.victory") + "!"
	case c.config.GameMode == gamedata.ModeArena && c.config.InfiniteMode:
		titleString = d.Get("menu.results.the_end")
	default:
		titleString = d.Get("menu.results.defeat")
	}
	titleLabel := eui.NewCenteredLabel(titleString, smallFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	lines := []string{
		fmt.Sprintf("%s: %v", d.Get("menu.results.time_played"), timeutil.FormatDuration(d, c.results.TimePlayed)),
		fmt.Sprintf("%s: %v", d.Get("menu.results.resources_gathered"), int(c.results.ResourcesGathered)),
		fmt.Sprintf("%s: %v", d.Get("menu.results.drones_total"), c.results.DronesProduced),
		fmt.Sprintf("%s: %v", d.Get("menu.results.creeps_defeated"), c.results.CreepsDefeated),
	}
	if c.config.GameMode != gamedata.ModeTutorial {
		if (c.config.GameMode == gamedata.ModeArena && c.config.InfiniteMode) || c.results.Victory {
			if c.highScore {
				lines = append(lines, fmt.Sprintf("%s: %v (%s)", d.Get("menu.results.score"), c.results.Score, d.Get("menu.results.new_record")))
			} else {
				lines = append(lines, fmt.Sprintf("%s: %v", d.Get("menu.results.score"), c.results.Score))
			}
		}
	}
	if c.config.GameMode == gamedata.ModeArena && c.config.InfiniteMode {
		lines = append(lines, fmt.Sprintf("%s: %d", d.Get("game.wave"), c.results.ArenaLevel))
	}

	for _, a := range c.newAchievements {
		lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.results.new_achievement"), d.Get("achievement", a)))
	}
	for _, a := range c.upgradedAchievements {
		lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.results.upgraded_achievement"), d.Get("achievement", a)))
	}
	for _, kind := range c.newDrones {
		lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.results.new_drone"), d.Get("drone", strings.ToLower(kind.String()))))
	}

	label := eui.NewCenteredLabel(strings.Join(lines, "\n"), smallFont)
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
