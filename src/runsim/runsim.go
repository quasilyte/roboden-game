package runsim

import (
	"errors"
	"time"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
)

var errTimeout = errors.New("simulation takes too long")

func NewState(ctx *ge.Context) *session.State {
	state := &session.State{
		Context: ctx,
		Persistent: session.PersistentData{
			Settings: session.GameSettings{
				Graphics: session.GraphicsSettings{
					ShadowsEnabled:       false,
					AllShadersEnabled:    false,
					VSyncEnabled:         false,
					CameraShakingEnabled: false,
				},
				MusicVolumeLevel:   0,
				EffectsVolumeLevel: 0,
			},
		},
	}
	state.CombinedInput = gameinput.Handler{
		Handler: ctx.Input.NewHandler(0, nil),
	}
	state.BoundInputs[0] = &state.CombinedInput
	return state
}

func PrepareAssets(ctx *ge.Context) {
	assetsConfig := &assets.Config{
		XM: true,
	}
	var progress float64
	assets.RegisterImageResources(ctx, assetsConfig, &progress)
	assets.RegisterRawResources(ctx)
	assets.RegisterShaderResources(ctx, assetsConfig, &progress)
}

func Run(state *session.State, levelGenChecksum, timeoutSeconds int, controller *staging.Controller) (serverapi.GameResults, error) {
	var simResult serverapi.GameResults

	runner, scene := ge.NewSimulatedScene(state.Context, controller)
	controller.Init(scene)

	if levelGenChecksum != 0 {
		if controller.GetLevelGenChecksum() != levelGenChecksum {
			return simResult, errors.New("levelgen checksum mismatch")
		}
	}

	timeout := (time.Duration(timeoutSeconds) * time.Second)

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
		if time.Since(start) >= timeout {
			return simResult, errTimeout
		}
	}
	return simResult, nil
}
