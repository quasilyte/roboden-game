package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/langs"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/runsim"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
)

func main() {
	runSimBinary := flag.String("runsim", "",
		"a path to a runsim executable")
	outputDir := flag.String("o", "",
		"an output directory")
	flag.Parse()

	if *runSimBinary == "" {
		panic("runsim binary path can't be empty")
	}
	if *outputDir == "" {
		panic("the output directory should be specified")
	}

	ctx := ge.NewContext(ge.ContextConfig{
		Mute:       true,
		FixedDelta: true,
	})
	ctx.Loader.OpenAssetFunc = assets.MakeOpenAssetFunc(ctx, "")
	ctx.Dict = langs.NewDictionary("en", 2)

	var rng gmath.Rand
	rng.SetSeed(time.Now().UnixNano())

	runsim.PrepareAssets(ctx)

	generation := int(time.Now().Unix())

	rstate := &runnerState{
		ctx:          ctx,
		runSimBinary: *runSimBinary,
		session:      runsim.NewState(ctx),
		rng:          &rng,
	}
	for i := 0; i < 1000; i++ {
		fmt.Printf("Running simulation #%d\n", i)
		results, config := runSimulation(rstate, "classic")
		filename := filepath.Join(*outputDir, fmt.Sprintf("%s_%d_%d.json", config.RawGameMode, generation, i))
		data, err := json.Marshal(runResults{
			Seed:    int(config.Seed),
			Env:     config.Environment,
			Victory: results.Victory,
			Score:   results.Score,
			Time:    results.Time,
			Mode:    config.RawGameMode,
			Drones:  config.Tier2Recipes,
			Turret:  config.TurretDesign,
			Core:    config.CoreDesign,
		})
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(filename, data, os.ModePerm); err != nil {
			panic(err)
		}
	}
}

type runnerState struct {
	ctx          *ge.Context
	runSimBinary string
	session      *session.State
	rng          *gmath.Rand
}

type runResults struct {
	Seed    int
	Env     int
	Victory bool
	Score   int
	Time    int
	Mode    string

	Drones []string
	Turret string
	Core   string
}

func runSimulation(rstate *runnerState, mode string) (serverapi.GameResults, gamedata.LevelConfig) {
	var replayConfig serverapi.ReplayLevelConfig
	replayConfig.RawGameMode = mode
	replayConfig.Seed = rstate.rng.PositiveInt64()

	// Create a random bot build.
	replayConfig.Tier2Recipes = gamedata.CreateDroneBuild(rstate.rng)
	replayConfig.TurretDesign = gamedata.PickTurretDesign(rstate.rng)
	replayConfig.PlayersMode = serverapi.PmodeSingleBot
	if rstate.rng.Bool() {
		replayConfig.CoreDesign = "den"
	} else {
		replayConfig.CoreDesign = "ark"
	}
	replayConfig.Environment = rstate.rng.IntRange(0, 1)

	// Some default settings.
	replayConfig.DronesPower = 1
	replayConfig.SuperCreeps = false
	replayConfig.Teleporters = 1
	replayConfig.OilRegenRate = 2
	replayConfig.Terrain = 1
	replayConfig.Resources = 2
	replayConfig.WorldSize = 2
	replayConfig.CreepDifficulty = 3

	// Mode-specific settings.
	switch mode {
	case "classic":
		replayConfig.InitialCreeps = 1
		replayConfig.NumCreepBases = 2
		replayConfig.CreepSpawnRate = 1
		replayConfig.BossDifficulty = 1
	default:
		panic("unexpected game mode")
	}

	config := gamedata.MakeLevelConfig(gamedata.ExecuteSimulation, replayConfig)
	config.Finalize()

	controller := staging.NewController(rstate.session, config, nil)
	simResult, err := runsim.Run(rstate.session, 35, controller)
	if err != nil {
		panic(err)
	}

	return simResult, config
}
