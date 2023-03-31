package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/scenes/menus"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/userdevice"
)

func main() {
	state := getDefaultSessionState()

	flag.StringVar(&state.MemProfile, "memprofile", "", "collect app heap allocations profile")
	flag.StringVar(&state.CPUProfile, "cpuprofile", "", "collect app cpu profile")
	extraScore := flag.Int("extra-score", 0, "add extra score points to the session")
	flag.Parse()

	ctx := ge.NewContext()
	ctx.Rand.SetSeed(time.Now().Unix())
	ctx.GameName = "roboden"
	ctx.WindowTitle = "Roboden"
	ctx.WindowWidth = 1920 / 2
	ctx.WindowHeight = 1080 / 2

	assets.Register(ctx)
	controls.BindKeymap(ctx, state)

	ctx.LoadGameData("save", &state.Persistent)
	state.ReloadLanguage(ctx)

	if *extraScore != 0 {
		state.Persistent.PlayerStats.TotalScore += *extraScore
	}

	ctx.FullScreen = state.Persistent.Settings.Graphics.FullscreenEnabled

	fmt.Println("is mobile?", state.Device.IsMobile)

	if err := ge.RunGame(ctx, menus.NewMainMenuController(state)); err != nil {
		panic(err)
	}
}

func newLevelConfig(config *session.LevelConfig) *session.LevelConfig {
	config.BuildTurretActionAvailable = true
	config.AttackActionAvailable = true
	config.RadiusActionAvailable = true

	config.ExtraUI = true
	config.EliteResources = true

	config.Tier2Recipes = []gamedata.AgentMergeRecipe{
		gamedata.FindRecipe(gamedata.ClonerAgentStats),
		gamedata.FindRecipe(gamedata.FighterAgentStats),
		gamedata.FindRecipe(gamedata.RepairAgentStats),
		gamedata.FindRecipe(gamedata.FreighterAgentStats),
		gamedata.FindRecipe(gamedata.CripplerAgentStats),
		gamedata.FindRecipe(gamedata.RedminerAgentStats),
		gamedata.FindRecipe(gamedata.ServoAgentStats),
	}

	config.Resources = 2
	config.WorldSize = 2
	config.CreepDifficulty = 1
	config.BossDifficulty = 1

	return config
}

func getDefaultSessionState() *session.State {
	state := &session.State{
		ArenaLevelConfig: newLevelConfig(&session.LevelConfig{}),
		LevelConfig: newLevelConfig(&session.LevelConfig{
			EnemyBoss:     true,
			InitialCreeps: 1,
			NumCreepBases: 2,
		}),
		Persistent: session.PersistentData{
			// The default settings.
			Settings: session.GameSettings{
				EffectsVolumeLevel: 2,
				MusicVolumeLevel:   2,
				ScrollingSpeed:     2,
				EdgeScrollRange:    2,
				Debug:              false,
				Lang:               inferDefaultLang(),
				Graphics: session.GraphicsSettings{
					ShadowsEnabled:    true,
					AllShadersEnabled: true,
					FullscreenEnabled: true,
				},
			},
		},
		Device: userdevice.GetInfo(),
	}

	state.Persistent.PlayerStats.TurretsUnlocked = append(state.Persistent.PlayerStats.TurretsUnlocked, gamedata.AgentGunpoint)

	for _, recipe := range gamedata.Tier2agentMergeRecipes {
		drone := recipe.Result
		if drone.ScoreCost != 0 {
			continue
		}
		state.Persistent.PlayerStats.DronesUnlocked = append(state.Persistent.PlayerStats.DronesUnlocked, drone.Kind)
	}

	return state
}

func inferDefaultLang() string {
	languages := ge.InferLanguages()
	defaultLanguage := "en"
	selectedLanguage := ""
	for _, l := range languages {
		switch l {
		case "en", "ru":
			if selectedLanguage != defaultLanguage {
				selectedLanguage = l
			}
		}
	}
	if selectedLanguage == "" {
		selectedLanguage = defaultLanguage
	}
	return selectedLanguage
}
