package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/langs"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
)

func main() {
	replayDataBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	var replayData serverapi.GameReplay
	if err := json.Unmarshal(replayDataBytes, &replayData); err != nil {
		panic(err)
	}

	config := gamedata.LevelConfig{
		ReplayLevelConfig:          replayData.Config,
		ExecMode:                   gamedata.ExecuteSimulation,
		AttackActionAvailable:      true,
		BuildTurretActionAvailable: true,
		RadiusActionAvailable:      true,
		EliteResources:             true,
		EnemyBoss:                  replayData.Config.RawGameMode == "classic",
	}
	switch replayData.Config.RawGameMode {
	case "classic":
		config.GameMode = gamedata.ModeClassic
	case "arena", "inf_arena":
		config.GameMode = gamedata.ModeArena
	}
	ctx := ge.NewContext(ge.ContextConfig{
		Mute:       true,
		FixedDelta: true,
	})
	ctx.Loader.OpenAssetFunc = assets.MakeOpenAssetFunc(ctx, "")
	ctx.Dict = langs.NewDictionary("en", 2)

	var progress float64
	assets.RegisterImageResources(ctx, &progress)
	assets.RegisterFontResources(ctx, &progress)
	assets.RegisterRawResources(ctx)
	assets.RegisterShaderResources(ctx, &progress)

	state := &session.State{}
	state.MainInput = gameinput.Handler{
		Handler: ctx.Input.NewHandler(0, nil),
	}

	controller := staging.NewController(state, config, nil)
	controller.SetReplayActions(replayData.Actions)
	runner, scene := ge.NewSimulatedScene(ctx, controller)
	controller.Init(scene)

	var simResult serverapi.GameResults
	start := time.Now()
OuterLoop:
	for {
		for i := 0; i < 60*60; i++ {
			runner.Update(1.0 / 60.0)
			var stop bool
			simResult, stop = controller.GetSimulationResult()
			if stop {
				break OuterLoop
			}
		}
		if time.Since(start) >= (30 * time.Second) {
			panic("simulation takes too long")
		}
	}

	encodedResult, err := json.Marshal(simResult)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(encodedResult))
}
