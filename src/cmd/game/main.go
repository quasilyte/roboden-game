package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/contentlock"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/scenes/menus"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/userdevice"
)

func main() {
	state := getDefaultSessionState()

	var gameDataFolder string
	flag.StringVar(&state.MemProfile, "memprofile", "", "collect app heap allocations profile")
	flag.StringVar(&state.CPUProfile, "cpuprofile", "", "collect app cpu profile")
	flag.StringVar(&gameDataFolder, "data", "", "a game data folder path")
	flag.Parse()

	ctx := ge.NewContext()
	ctx.Rand.SetSeed(time.Now().Unix())
	ctx.GameName = "roboden"
	ctx.WindowTitle = "Roboden"
	ctx.WindowWidth = 1920 / 2
	ctx.WindowHeight = 1080 / 2

	if gameDataFolder == "" {
		gameDataFolder = "roboden_data"
	}

	ctx.Loader.OpenAssetFunc = assets.MakeOpenAssetFunc(ctx, gameDataFolder)
	assets.RegisterRawResources(ctx)
	controls.BindKeymap(ctx, state)

	if err := ctx.LoadGameData("save", &state.Persistent); err != nil {
		fmt.Printf("can't load game data: %v", err)
		state.Persistent = contentlock.GetDefaultData()
		ctx.SaveGameData("save", state.Persistent)
	}
	state.ReloadLanguage(ctx)

	ctx.FullScreen = state.Persistent.Settings.Graphics.FullscreenEnabled

	fmt.Println("is mobile?", state.Device.IsMobile)

	if err := ge.RunGame(ctx, menus.NewBootloadController(state)); err != nil {
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
		gamedata.FindRecipe(gamedata.CripplerAgentStats),
		gamedata.FindRecipe(gamedata.RechargeAgentStats),
		gamedata.FindRecipe(gamedata.RedminerAgentStats),
		gamedata.FindRecipe(gamedata.ServoAgentStats),
	}
	config.TurretDesign = gamedata.GunpointAgentStats

	config.Resources = 2
	config.WorldSize = 2
	config.CreepDifficulty = 1
	config.BossDifficulty = 1

	return config
}

func getDefaultSessionState() *session.State {
	state := &session.State{
		ArenaLevelConfig: newLevelConfig(&session.LevelConfig{
			ArenaProgression: 1,
		}),
		LevelConfig: newLevelConfig(&session.LevelConfig{
			EnemyBoss:     true,
			InitialCreeps: 1,
			NumCreepBases: 2,
		}),
		Persistent: contentlock.GetDefaultData(),
		Device:     userdevice.GetInfo(),
	}

	for _, recipe := range gamedata.Tier2agentMergeRecipes {
		drone := recipe.Result
		if drone.ScoreCost != 0 {
			continue
		}
		state.Persistent.PlayerStats.DronesUnlocked = append(state.Persistent.PlayerStats.DronesUnlocked, drone.Kind.String())
	}

	for _, turret := range gamedata.TurretStatsList {
		if turret.ScoreCost != 0 {
			continue
		}
		state.Persistent.PlayerStats.TurretsUnlocked = append(state.Persistent.PlayerStats.TurretsUnlocked, turret.Kind.String())
	}

	return state
}
