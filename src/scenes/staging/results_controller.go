package staging

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/contentlock"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/timeutil"
)

type resultsController struct {
	state  *session.State
	config *gamedata.LevelConfig

	scene          *ge.Scene
	backController ge.SceneController

	resultTag string
	highScore bool
	rewards   *gameRewards

	results battleResults
}

type battleResults struct {
	Victory           bool
	GroundBossDefeat  bool
	BossDefeated      bool
	GroundControl     bool
	AtomicBombVictory bool
	TimePlayed        time.Duration
	Ticks             int

	ResourcesGathered      float64
	EliteResourcesGathered float64
	DronesProduced         int
	CreepsDefeated         int
	CreepsStomped          int
	CreepTotalValue        int
	CreepFragScore         int
	CreepBasesDestroyed    int

	DominatorsSurvived int

	T3created       int
	ColoniesBuilt   int
	RadiusIncreases int

	EnemyColonyDamage            float64
	EnemyColonyDamageFromTurrets float64

	YellowFactionUsed  bool
	RedFactionUsed     bool
	GreenFactionUsed   bool
	BlueFactionUsed    bool
	OpenedEvolutionTab bool
	Paused             bool

	RedCrystalsCollected int

	ArenaLevel           int
	Score                int
	DifficultyScore      int
	DronePointsAllocated int

	OnlyTier1Military bool

	Replay [][]serverapi.PlayerAction

	Tier3Drones []gamedata.ColonyAgentKind
}

func newResultsController(state *session.State, config *gamedata.LevelConfig, backController ge.SceneController, results battleResults, rewards *gameRewards) *resultsController {
	return &resultsController{
		state:          state,
		backController: backController,
		results:        results,
		config:         config,
		rewards:        rewards,
	}
}

func (c *resultsController) Init(scene *ge.Scene) {
	c.scene = scene
	eui.AddBackground(c.state.BackgroundImage, scene)

	firstTime := false
	if c.rewards == nil {
		firstTime = true
		c.rewards = &gameRewards{}
	}

	victory := c.results.Victory || c.config.GameMode == gamedata.ModeInfArena
	if victory && firstTime {
		c.updateProgress()
		c.scene.Context().SaveGameData("save", c.state.Persistent)
	}

	c.initUI()
}

func (c *resultsController) makeGameReplay() serverapi.GameReplay {
	var replay serverapi.GameReplay
	replay.GameVersion = gamedata.BuildNumber
	replay.Config = c.config.ReplayLevelConfig
	replay.Actions = c.results.Replay
	replay.Results.Score = c.results.Score
	replay.Results.Victory = c.results.Victory
	replay.Results.Time = int(math.Floor(c.results.TimePlayed.Seconds()))
	replay.Results.Ticks = c.results.Ticks
	return replay
}

func (c *resultsController) updateProgress() {
	stats := &c.state.Persistent.PlayerStats

	stats.TotalPlayTime += c.results.TimePlayed

	if c.config.GameMode == gamedata.ModeTutorial {
		if !stats.TutorialCompleted {
			stats.TotalScore += c.results.Score
			if !xslices.Contains(stats.ModesUnlocked, "classic") {
				stats.ModesUnlocked = append(stats.ModesUnlocked, "classic")
			}
			stats.TutorialCompleted = true
		}
		return
	}

	t3drones := map[gamedata.ColonyAgentKind]struct{}{}
	for _, name := range stats.Tier3DronesSeen {
		t3drones[gamedata.DroneKindByName[name]] = struct{}{}
	}
	for _, k := range c.results.Tier3Drones {
		if _, ok := t3drones[k]; ok {
			continue
		}
		stats.Tier3DronesSeen = append(stats.Tier3DronesSeen, k.String())
	}

	stats.TotalScore += c.results.Score
	switch c.config.GameMode {
	case gamedata.ModeClassic:
		if stats.HighestClassicScore < c.results.Score {
			c.highScore = true
			stats.HighestClassicScore = c.results.Score
			stats.HighestClassicScoreDifficulty = c.results.DifficultyScore
		}
	case gamedata.ModeInfArena:
		if stats.HighestInfArenaScore < c.results.Score {
			c.highScore = true
			stats.HighestInfArenaScore = c.results.Score
			stats.HighestInfArenaScoreDifficulty = c.results.DifficultyScore
		}
	case gamedata.ModeArena:
		if stats.HighestArenaScore < c.results.Score {
			c.highScore = true
			stats.HighestArenaScore = c.results.Score
			stats.HighestArenaScoreDifficulty = c.results.DifficultyScore
		}
	case gamedata.ModeReverse:
		if stats.HighestReverseScore < c.results.Score {
			c.highScore = true
			stats.HighestReverseScore = c.results.Score
			stats.HighestReverseScoreDifficulty = c.results.DifficultyScore
		}
	}

	contentUpdates := contentlock.Update(c.state)
	c.rewards.newAchievements, c.rewards.upgradedAchievements = c.checkAchievements()
	c.rewards.newCores = contentUpdates.CoresUnlocked
	c.rewards.newDrones = contentUpdates.DronesUnlocked
	c.rewards.newTurrets = contentUpdates.TurretsUnlocked
	c.rewards.newOptions = contentUpdates.OptionsUnlocked
	c.rewards.newModes = contentUpdates.ModesUnlocked
}

func (c *resultsController) Update(delta float64) {
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		if c.rewards.IsEmpty() {
			c.back()
		} else {
			c.claimRewards()
		}
		return
	}
}

func (c *resultsController) checkAchievements() ([]string, []string) {
	var newAchievements []string
	var upgradedAchievements []string

	if c.config.PlayersMode != serverapi.PmodeSinglePlayer {
		return nil, nil
	}

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
	if c.results.DifficultyScore >= 160 {
		difficultyLevel = 2
	}

	needSave := false
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
		case "t3engineer":
			unlocked = len(stats.Tier3DronesSeen) >= len(gamedata.Tier3agentMergeRecipes) && c.config.GameMode != gamedata.ModeReverse
		case "trample":
			unlocked = c.results.CreepsStomped != 0 && c.config.GameMode != gamedata.ModeReverse
		case "nopeeking":
			unlocked = !c.results.OpenedEvolutionTab && c.config.GameMode != gamedata.ModeReverse
		case "nonstop":
			unlocked = !c.results.Paused && c.config.GameMode != gamedata.ModeReverse

		case "impossible":
			unlocked = c.results.DifficultyScore > 200
		case "speedrunning":
			unlocked = c.results.TimePlayed.Minutes() < 15
		case "victorydrag":
			unlocked = c.results.TimePlayed.Hours() >= 2
		case "t3less":
			unlocked = c.results.T3created == 0
		case "cheapbuild10":
			unlocked = c.results.DronePointsAllocated <= 10
		case "hightension":
			unlocked = c.config.WorldSize == 0 &&
				c.config.NumCreepBases != 0 &&
				c.results.CreepBasesDestroyed == 0
		case "solobase":
			unlocked = c.results.ColoniesBuilt == 0
		case "uiless":
			unlocked = c.config.InterfaceMode == 0
		case "powerof3":
			unlocked = !c.results.YellowFactionUsed || !c.results.RedFactionUsed || !c.results.GreenFactionUsed || !c.results.BlueFactionUsed
		case "tinyradius":
			unlocked = c.results.RadiusIncreases == 0
		case "t1army":
			unlocked = c.results.OnlyTier1Military
		case "groundwin":
			unlocked = c.results.GroundBossDefeat
		case "turretdamage":
			unlocked = c.results.EnemyColonyDamageFromTurrets >= (c.results.EnemyColonyDamage * 0.25)
		case "leet":
			unlocked = c.config.Seed == 1337

		case "antidominator":
			unlocked = c.results.DominatorsSurvived == 0

		case "infinite":
			unlocked = c.results.ArenaLevel >= 35

		case "colonyhunter":
			unlocked = c.results.ColoniesBuilt >= 3
		case "groundcontrol":
			unlocked = c.results.GroundControl
		case "atomicfinisher":
			unlocked = c.results.AtomicBombVictory
		}

		if !unlocked {
			continue
		}

		elite := difficultyLevel == 2 || a.OnlyElite
		if _, ok := alreadyAchieved[a.Name]; ok {
			stats.Achievements = xslices.Remove(stats.Achievements, session.Achievement{
				Name:  a.Name,
				Elite: alreadyAchieved[a.Name] == 2,
			})
			upgradedAchievements = append(upgradedAchievements, a.Name)
		} else {
			newAchievements = append(newAchievements, a.Name)
		}

		updated := c.state.UnlockAchievement(session.Achievement{
			Name:  a.Name,
			Elite: elite,
		})
		if updated {
			needSave = true
		}
	}

	if needSave {
		c.scene.Context().SaveGameData("save", c.state.Persistent)
	}

	return newAchievements, upgradedAchievements
}

func (c *resultsController) calcResultTag() (string, bool) {
	if c.config.GameMode == gamedata.ModeReverse && c.config.PlayersMode == serverapi.PmodeTwoPlayers {
		if c.results.BossDefeated {
			return "menu.results.player2_win", false
		}
		return "menu.results.player1_win", false
	}
	switch {
	case c.results.Victory:
		return "menu.results.victory", true
	case c.config.GameMode == gamedata.ModeInfArena:
		return "menu.results.the_end", false
	default:
		return "menu.results.defeat", false
	}
}

func (c *resultsController) initUI() {
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(320, 10, nil)
	root.AddChild(rowContainer)

	smallFont := assets.BitmapFont1

	d := c.scene.Dict()

	var victory bool
	c.resultTag, victory = c.calcResultTag()
	titleString := d.Get(c.resultTag)
	if victory {
		titleString += "!"
	}

	titleLabel := eui.NewCenteredLabel(titleString, assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	panel := eui.NewTextPanel(uiResources, 0, 0)

	grid := eui.NewGridContainer(2, widget.GridLayoutOpts.Spacing(24, 4),
		widget.GridLayoutOpts.Stretch([]bool{true, false}, nil))
	panel.AddChild(grid)

	itoa := strconv.Itoa

	lines := [][2]string{
		{d.Get("menu.results.time_played"), timeutil.FormatDuration(d, c.results.TimePlayed)},
	}
	if c.config.GameMode != gamedata.ModeReverse {
		lines = append(lines, [2]string{d.Get("menu.results.resources_gathered"), itoa(int(c.results.ResourcesGathered))})
	}
	lines = append(lines,
		[2]string{d.Get("menu.results.drones_total"), itoa(c.results.DronesProduced)},
		[2]string{d.Get("menu.results.creeps_defeated"), itoa(c.results.CreepsDefeated)},
	)
	if c.config.GameMode != gamedata.ModeTutorial {
		if (c.config.GameMode == gamedata.ModeInfArena) || c.results.Victory {
			if c.highScore {
				lines = append(lines, [2]string{d.Get("menu.results.score"), fmt.Sprintf("%v (%s)", c.results.Score, d.Get("menu.results.new_record"))})
			} else {
				lines = append(lines, [2]string{d.Get("menu.results.score"), itoa(c.results.Score)})
			}
		}
	}
	if c.config.GameMode == gamedata.ModeInfArena {
		lines = append(lines, [2]string{d.Get("game.wave"), itoa(c.results.ArenaLevel)})
	}

	for _, pair := range lines {
		grid.AddChild(eui.NewLabel(pair[0], smallFont))
		grid.AddChild(eui.NewLabel(pair[1], smallFont))
	}

	rowContainer.AddChild(panel)

	replay := c.makeGameReplay()
	if c.config.GameMode != gamedata.ModeTutorial && c.highScore {
		c.state.SentHighscores = false
		key := c.config.RawGameMode + "_highscore"
		c.scene.Context().SaveGameData(key, replay)
	}
	if gamedata.IsRunnableReplay(replay) {
		r := session.SavedReplay{
			Date:      time.Now(),
			ResultTag: c.resultTag,
			Replay:    replay,
		}
		recentReplayKey := c.state.ReplayDataKey(0)
		c.scene.Context().SaveGameData(recentReplayKey, r)

		saved := false
		rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.save_replay"), func() {
			// Let the user press button several times, but don't do any extra
			// work if we already saved the replay data.
			if saved {
				return
			}
			saved = true
			k := c.state.ReplayDataKey(c.state.FindNextReplayIndex())
			c.scene.Context().SaveGameData(k, r)
		}))
	}
	if gamedata.IsSendableReplay(replay) {
		rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.publish_score"), func() {
			if c.state.Persistent.PlayerName == "" {
				backController := newResultsController(c.state, c.config, c.backController, c.results, c.rewards)
				userNameScene := c.state.SceneRegistry.UserNameMenu(backController)
				c.scene.Context().ChangeScene(userNameScene)
				return
			}
			nextController := c.backController
			if !c.rewards.IsEmpty() {
				nextController = newRewardsController(c.state, *c.rewards, c.backController)
			}
			submitController := c.state.SceneRegistry.SubmitScreen(nextController, []serverapi.GameReplay{replay})
			c.scene.Context().ChangeScene(submitController)
		}))
	}

	if c.rewards.IsEmpty() {
		rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.lobby_back"), func() {
			c.back()
		}))
	} else {
		rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.results.claim_rewards"), func() {
			c.claimRewards()
		}))
	}

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *resultsController) claimRewards() {
	c.scene.Context().ChangeScene(newRewardsController(c.state, *c.rewards, c.backController))
}

func (c *resultsController) back() {
	c.scene.Context().ChangeScene(c.backController)
}
