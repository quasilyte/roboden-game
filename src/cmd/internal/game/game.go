package game

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gdata"
	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/buildinfo"
	"github.com/quasilyte/roboden-game/contentlock"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/scenes/menus"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/userdevice"
)

func Main() {
	state := getDefaultSessionState()

	{
		m, err := gdata.Open(gdata.Config{AppName: "ge_game_roboden"})
		if err != nil {
			// It's OK to continue, we'll just have to avoid referencing
			// gamedata object if it's nil.
			state.Logf("failed to initialize game data storage: %v", err)
		} else {
			state.GameData = m
			if runtime.GOARCH == "wasm" {
				// We were using "save" key before, now it's "save.json";
				// migrate the old save item to a new key if localStorage
				// contains the old key.
				maybeMigrateSaves(state)
			}
		}
	}

	{
		deviceInfo, err := userdevice.GetInfo()
		switch {
		case err != nil:
			if deviceInfo.Steam.Enabled {
				state.Logf("failed to get Steam info: %v", err)
			} else {
				panic(fmt.Sprintf("unexpected error: %v", err))
			}
		case !deviceInfo.Steam.Enabled:
			state.Logf("running a non-Steam build")
		default:
			steamDeckSuffix := ""
			if deviceInfo.IsSteamDeck() {
				steamDeckSuffix = " (Steam Deck)"
			}
			state.Logf("Steam SDK initialized successfully" + steamDeckSuffix)
		}
		state.Device = deviceInfo
	}

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
		state.Logf("server proto=%q host=%q path=%q", state.ServerProtocol, state.ServerHost, state.ServerPath)
	} else {
		// On wasm (inside the browser) we're hardcoding the server data for now.
		state.ServerProtocol = "https"
		state.ServerHost = "quasilyte.tech"
		state.ServerPath = "/roboden/api"
		// Also, force XM tracks when running a web build.
		// We don't bundle OGG tracks for it.
		state.Persistent.Settings.XM = true
	}

	if runtime.GOOS == "android" {
		state.Persistent.Settings.XM = true
	}

	ctx := ge.NewContext(ge.ContextConfig{
		FixedDelta: true,
	})
	ctx.Rand.SetSeed(time.Now().Unix())
	ctx.GameName = "roboden"
	ctx.WindowTitle = "Roboden"

	if gameDataFolder == "" {
		if runtime.GOARCH == "wasm" {
			gameDataFolder = "roboden_data"
		} else {
			gameLocation, err := os.Executable()
			if err != nil {
				state.Logf("error getting executable path: %v", err)
				gameLocation = os.Args[0]
			}
			gameLocation = filepath.Dir(gameLocation)
			state.Logf("game location: %q", gameLocation)
			gameDataFolder = filepath.Join(gameLocation, "roboden_data")
		}
	}

	state.Logf("buildinfo distribution tag: %s", buildinfo.Distribution)

	ctx.Loader.OpenAssetFunc = assets.MakeOpenAssetFunc(ctx, gameDataFolder)
	assets.RegisterRawResources(ctx)

	keymaps := controls.BindKeymap(ctx)
	state.TouchInput = gameinput.Handler{InputMethod: gameinput.InputMethodTouch, Handler: ctx.Input.NewHandler(0, keymaps.TouchKeymap)}
	state.CombinedInput = gameinput.Handler{InputMethod: gameinput.InputMethodCombined, Handler: ctx.Input.NewHandler(0, keymaps.CombinedKeymap)}
	state.KeyboardInput = gameinput.Handler{InputMethod: gameinput.InputMethodKeyboard, Handler: ctx.Input.NewHandler(0, keymaps.KeyboardKeymap)}
	state.FirstGamepadInput = gameinput.Handler{InputMethod: gameinput.InputMethodGamepad1, Handler: ctx.Input.NewHandler(0, keymaps.FirstGamepadKeymap)}
	state.SecondGamepadInput = gameinput.Handler{InputMethod: gameinput.InputMethodGamepad2, Handler: ctx.Input.NewHandler(1, keymaps.SecondGamepadKeymap)}

	if state.CheckGameItem("save.json") {
		if err := state.LoadGameItem("save.json", &state.Persistent); err != nil {
			state.Logf("can't load game data: %v", err)
			state.Persistent = contentlock.GetDefaultData()
			contentlock.Update(state)
			state.SaveGameItem("save.json", state.Persistent)
		} else {
			// Loaded without errors.
			if state.Persistent.FirstLaunch && state.Persistent.PlayerStats.TotalPlayTime > 0 {
				// We introduced the first launch flag after the release.
				// Some players may have already played the game,
				// so we don't want them to be marked as first-timers.
				state.Persistent.FirstLaunch = false
				state.SaveGameItem("save.json", state.Persistent)
			}
			// Re-check the content.
			contentlock.Update(state)
		}
	} else {
		// This is a first launch.
		state.Logf("save data does not exist")
		state.SaveGameItem("save.json", state.Persistent)
	}
	if state.Device.IsMobile() {
		// For mobile devices, it's always a touch control.
		state.Persistent.Settings.Player1InputMethod = int(gameinput.InputMethodTouch)
		// You can't play on mobiles without those.
		state.Persistent.Settings.ScreenButtons = true
	}
	state.ReloadInputs()
	state.ReloadLanguage(ctx)

	displayRatio := gamedata.SupportedDisplayRatios[state.Persistent.Settings.Graphics.AspectRatio]
	ctx.WindowWidth = displayRatio.Width
	ctx.WindowHeight = displayRatio.Height
	ctx.ScreenWidth = displayRatio.Width
	ctx.ScreenHeight = displayRatio.Height

	if state.Device.IsMobile() {
		state.MenuInput = &state.TouchInput
	} else {
		state.MenuInput = &state.CombinedInput
	}

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
	state.GameCommitHash = CommitHash

	state.Logf("is mobile? %v", state.Device.IsMobile())
	state.Logf("game commit version: %v", CommitHash)

	ctx.NewPanicController = func(panicInfo *ge.PanicInfo) ge.SceneController {
		return menus.NewPanicController(panicInfo)
	}

	gamedata.Validate()

	ebiten.SetVsyncEnabled(state.Persistent.Settings.Graphics.VSyncEnabled)

	if err := ge.RunGame(ctx, menus.NewBootloadController(state)); err != nil {
		panic(err)
	}
}

func maybeMigrateSaves(state *session.State) {
	if !state.GameData.ItemExists("save") {
		state.Logf("using an updated localStorage library")
		return
	}

	state.Logf("an old-style localStorage save detected; migrating it to a new slot")
	data, err := state.GameData.LoadItem("save")
	if err != nil {
		state.Logf("saves migration failed: %v", err)
		return
	}
	if err := state.GameData.SaveItem("save.json", data); err != nil {
		state.Logf("saves migration failed: %v", err)
		return
	}
	// If everything is fine, remove the old-style save.
	state.GameData.DeleteItem("save")
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
	config.CoreDesign = gamedata.DenCoreStats.Name

	config.Relicts = true
	config.GoldEnabled = true
	config.OilRegenRate = 2
	config.Terrain = 1
	config.GameSpeed = 1
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
				Teleporters:           1,
				RawGameMode:           "reverse",
				TechProgressRate:      6,
				ReverseSuperCreepRate: 3,
				DronesPower:           1,
				InitialCreeps:         1,
				BossDifficulty:        2,
				AtomicBomb:            true,
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
		BlitzLevelConfig: newLevelConfig(&gamedata.LevelConfig{
			ReplayLevelConfig: serverapi.ReplayLevelConfig{
				SuperCreeps:       false,
				InitialCreeps:     0,
				NumCreepBases:     2,
				CreepSpawnRate:    1,
				Teleporters:       1,
				RawGameMode:       "blitz",
				DronesPower:       1,
				StartingResources: true,
			},
		}),
		Persistent: contentlock.GetDefaultData(),
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
		config.StartingResources = true
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
		config.Relicts = false
		config.StartingResources = false
		config.Teleporters = 1
		config.InterfaceMode = 2
		config.InitialCreeps = 0
		config.EliteResources = true
		config.EnemyBoss = false
		config.CreepDifficulty = 0
		config.BossDifficulty = 0
		config.NumCreepBases = 0
		config.Environment = int(gamedata.EnvInferno)

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

	for _, core := range gamedata.CoreStatsList {
		if core.ScoreCost != 0 {
			continue
		}
		state.Persistent.PlayerStats.CoresUnlocked = append(state.Persistent.PlayerStats.CoresUnlocked, core.Name)
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
