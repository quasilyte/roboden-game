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
	"github.com/quasilyte/roboden-game/gameinput"
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
	keymaps := controls.BindKeymap(ctx)
	state.CombinedInput = keymaps.CombinedInput
	state.KeyboardInput = keymaps.KeyboardInput
	state.FirstGamepadInput = keymaps.FirstGamepadInput
	state.SecondGamepadInput = keymaps.SecondGamepadInput

	if err := ctx.LoadGameData("save", &state.Persistent); err != nil {
		fmt.Printf("can't load game data: %v", err)
		state.Persistent = contentlock.GetDefaultData()
		contentlock.Update(state)
		ctx.SaveGameData("save", state.Persistent)
	} else {
		contentlock.Update(state)
	}
	state.ReloadInputs()
	state.ReloadLanguage(ctx)

	state.CombinedInput.SetGamepadDeadzoneLevel(state.Persistent.Settings.GamepadSettings[0].DeadzoneLevel)
	state.FirstGamepadInput.SetGamepadDeadzoneLevel(state.Persistent.Settings.GamepadSettings[0].DeadzoneLevel)
	state.SecondGamepadInput.SetGamepadDeadzoneLevel(state.Persistent.Settings.GamepadSettings[1].DeadzoneLevel)

	state.CombinedInput.SetVirtualCursorSpeed(state.Persistent.Settings.GamepadSettings[0].CursorSpeed)
	state.FirstGamepadInput.SetVirtualCursorSpeed(state.Persistent.Settings.GamepadSettings[0].CursorSpeed)
	state.SecondGamepadInput.SetVirtualCursorSpeed(state.Persistent.Settings.GamepadSettings[1].CursorSpeed)

	state.CombinedInput.SetGamepadLayout(gameinput.GamepadLayoutKind(state.Persistent.Settings.GamepadSettings[0].Layout))
	state.FirstGamepadInput.SetGamepadLayout(gameinput.GamepadLayoutKind(state.Persistent.Settings.GamepadSettings[0].Layout))
	state.SecondGamepadInput.SetGamepadLayout(gameinput.GamepadLayoutKind(state.Persistent.Settings.GamepadSettings[1].Layout))

	ctx.FullScreen = state.Persistent.Settings.Graphics.FullscreenEnabled

	registerScenes(state)
	state.Context = ctx

	fmt.Println("is mobile?", state.Device.IsMobile)
	fmt.Println("game commit version:", CommitHash)

	ctx.NewPanicController = func(panicInfo *ge.PanicInfo) ge.SceneController {
		return menus.NewPanicController(panicInfo)
	}

	gamedata.Validate()

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

func newLevelConfig(options *gamedata.LevelConfig) *gamedata.LevelConfig {
	config := gamedata.MakeLevelConfig(gamedata.ExecuteNormal, options.ReplayLevelConfig)

	config.InterfaceMode = 2

	config.PlayersMode = serverapi.PmodeSinglePlayer

	config.Tier2Recipes = []string{
		gamedata.ClonerAgentStats.Kind.String(),
		gamedata.FighterAgentStats.Kind.String(),
		gamedata.RepairAgentStats.Kind.String(),
		gamedata.CripplerAgentStats.Kind.String(),
		gamedata.RechargerAgentStats.Kind.String(),
		gamedata.RedminerAgentStats.Kind.String(),
		gamedata.ServoAgentStats.Kind.String(),
	}
	config.TurretDesign = gamedata.GunpointAgentStats.Kind.String()

	config.OilRegenRate = 2
	config.Terrain = 1
	config.Resources = 2
	config.WorldSize = 2
	config.CreepDifficulty = 3
	if config.BossDifficulty == 0 {
		config.BossDifficulty = 1
	}

	return &config
}

func getDefaultSessionState() *session.State {
	state := &session.State{
		ReverseLevelConfig: newLevelConfig(&gamedata.LevelConfig{
			ReplayLevelConfig: serverapi.ReplayLevelConfig{
				Teleporters:      1,
				RawGameMode:      "reverse",
				TechProgressRate: 5,
				DronesPower:      1,
				InitialCreeps:    1,
				BossDifficulty:   2,
			},
		}),
		ArenaLevelConfig: newLevelConfig(&gamedata.LevelConfig{
			ReplayLevelConfig: serverapi.ReplayLevelConfig{
				ArenaProgression: 1,
				Teleporters:      1,
				RawGameMode:      "arena",
				DronesPower:      1,
			},
		}),
		InfArenaLevelConfig: newLevelConfig(&gamedata.LevelConfig{
			ReplayLevelConfig: serverapi.ReplayLevelConfig{
				ArenaProgression: 1,
				Teleporters:      1,
				RawGameMode:      "inf_arena",
				DronesPower:      1,
			},
		}),
		ClassicLevelConfig: newLevelConfig(&gamedata.LevelConfig{
			ReplayLevelConfig: serverapi.ReplayLevelConfig{
				SuperCreeps:    false,
				InitialCreeps:  1,
				NumCreepBases:  2,
				CreepSpawnRate: 1,
				Teleporters:    1,
				RawGameMode:    "classic",
				DronesPower:    1,
			},
		}),
		Persistent: contentlock.GetDefaultData(),
		Device:     userdevice.GetInfo(),
	}

	{
		config := state.ClassicLevelConfig.Clone()
		config.WorldSize = 2
		config.Resources = 4
		config.DronesPower = 1
		config.CreepDifficulty = 1
		config.BossDifficulty = 0
		config.NumCreepBases = 1
		config.FogOfWar = false
		config.StartingResources = 2
		config.InitialCreeps = 2
		config.Teleporters = 2

		state.SplashLevelConfig = &config
	}

	{
		config := state.ClassicLevelConfig.Clone()
		config.RawGameMode = "tutorial"
		state.TutorialLevelConfig = &config
		config.WorldSize = 0
		config.Resources = 1
		config.StartingResources = 0
		config.Teleporters = 1
		config.InterfaceMode = 2
		config.InitialCreeps = 0
		config.EliteResources = true
		config.EnemyBoss = false
		config.CreepDifficulty = 0
		config.BossDifficulty = 0
		config.NumCreepBases = 0

		config.ExtraDrones = []*gamedata.AgentStats{}
		for i := 0; i < 2; i++ {
			config.ExtraDrones = append(config.ExtraDrones, gamedata.ServoAgentStats)
		}
		for i := 0; i < 5; i++ {
			config.ExtraDrones = append(config.ExtraDrones, gamedata.WorkerAgentStats)
		}
		for i := 0; i < 3; i++ {
			config.ExtraDrones = append(config.ExtraDrones, gamedata.ScoutAgentStats)
		}

		config.Finalize()
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
