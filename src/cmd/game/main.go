package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/contentlock"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/scenes/menus"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/userdevice"
)

func main() {
	state := getDefaultSessionState()

	var gameDataFolder string
	var serverAddress string
	flag.StringVar(&state.MemProfile, "memprofile", "", "collect app heap allocations profile")
	flag.StringVar(&state.CPUProfile, "cpuprofile", "", "collect app cpu profile")
	flag.StringVar(&gameDataFolder, "data", "", "a game data folder path")
	flag.StringVar(&serverAddress, "server", DefaultServerAddr, "leaderboard server address")
	flag.Parse()

	if runtime.GOARCH != "wasm" {
		// It's possible to use a localhost server on desktops.
		// Or alternative leaderboad servers for what it's worth.
		parsedAddress, err := url.Parse(serverAddress)
		if err != nil {
			state.ServerProtocol = "http"
			state.ServerHost = "127.0.0.1:8080"
			state.ServerPath = ""
		} else {
			if parsedAddress.Scheme == "" {
				state.ServerProtocol = "http"
			} else {
				state.ServerProtocol = parsedAddress.Scheme
			}
			state.ServerHost = parsedAddress.Host
			state.ServerPath = parsedAddress.Path
		}
		fmt.Printf("server proto=%q host=%q path=%q\n", state.ServerProtocol, state.ServerHost, state.ServerPath)
	} else {
		// On wasm (inside the browser) we're hardcoding the server data for now.
		state.ServerProtocol = "https"
		state.ServerHost = "quasilyte.tech"
		state.ServerPath = "/roboden/api"
	}

	ctx := ge.NewContext(ge.ContextConfig{
		FixedDelta: true,
	})
	ctx.Rand.SetSeed(time.Now().Unix())
	ctx.GameName = "roboden"
	ctx.WindowTitle = "Roboden"
	ctx.WindowWidth = 1920 / 2
	ctx.WindowHeight = 1080 / 2

	if gameDataFolder == "" {
		if runtime.GOARCH == "wasm" {
			gameDataFolder = "roboden_data"
		} else {
			gameLocation, err := os.Executable()
			if err != nil {
				fmt.Printf("error getting executable path: %v\n", err)
				gameLocation = os.Args[0]
			}
			gameLocation = filepath.Dir(gameLocation)
			fmt.Printf("game location: %q\n", gameLocation)
			gameDataFolder = filepath.Join(gameLocation, "roboden_data")
		}
	}

	ctx.Loader.OpenAssetFunc = assets.MakeOpenAssetFunc(ctx, gameDataFolder)
	assets.RegisterRawResources(ctx)
	state.MainInput = controls.BindKeymap(ctx)

	if err := ctx.LoadGameData("save", &state.Persistent); err != nil {
		fmt.Printf("can't load game data: %v", err)
		state.Persistent = contentlock.GetDefaultData()
		contentlock.Update(state)
		ctx.SaveGameData("save", state.Persistent)
	} else {
		contentlock.Update(state)
	}
	state.ReloadLanguage(ctx)

	ctx.FullScreen = state.Persistent.Settings.Graphics.FullscreenEnabled

	registerScenes(state)
	state.Context = ctx

	fmt.Println("is mobile?", state.Device.IsMobile)
	fmt.Println("game commit version:", CommitHash)

	if err := ge.RunGame(ctx, menus.NewBootloadController(state)); err != nil {
		panic(err)
	}
}

func registerScenes(state *session.State) {
	state.SceneRegistry.UserNameMenu = func(backController ge.SceneController) ge.SceneController {
		return menus.NewUserNameMenuController(state, backController)
	}
	state.SceneRegistry.SubmitScreen = func(backController ge.SceneController, replays []serverapi.GameReplay) ge.SceneController {
		return menus.NewSubmitScreenController(state, backController, replays)
	}
}

func newLevelConfig(config *gamedata.LevelConfig) *gamedata.LevelConfig {
	config.BuildTurretActionAvailable = true
	config.AttackActionAvailable = true
	config.RadiusActionAvailable = true

	config.ExtraUI = true
	config.EliteResources = true

	config.Tier2Recipes = []string{
		gamedata.ClonerAgentStats.Kind.String(),
		gamedata.FighterAgentStats.Kind.String(),
		gamedata.RepairAgentStats.Kind.String(),
		gamedata.CripplerAgentStats.Kind.String(),
		gamedata.RechargeAgentStats.Kind.String(),
		gamedata.RedminerAgentStats.Kind.String(),
		gamedata.ServoAgentStats.Kind.String(),
	}
	config.TurretDesign = gamedata.GunpointAgentStats.Kind.String()

	config.Resources = 2
	config.WorldSize = 2
	config.CreepDifficulty = 1
	config.BossDifficulty = 1

	return config
}

func getDefaultSessionState() *session.State {
	state := &session.State{
		ArenaLevelConfig: newLevelConfig(&gamedata.LevelConfig{
			ReplayLevelConfig: serverapi.ReplayLevelConfig{
				ArenaProgression: 1,
				Teleporters:      1,
			},
		}),
		LevelConfig: newLevelConfig(&gamedata.LevelConfig{
			EnemyBoss: true,
			ReplayLevelConfig: serverapi.ReplayLevelConfig{
				InitialCreeps:  1,
				NumCreepBases:  2,
				CreepSpawnRate: 1,
				Teleporters:    1,
			},
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
