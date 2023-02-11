package main

import (
	"time"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/scenes/menus"
	"github.com/quasilyte/roboden-game/session"
)

func main() {
	ctx := ge.NewContext()
	ctx.Rand.SetSeed(time.Now().Unix())
	ctx.GameName = "roboden"
	ctx.WindowTitle = "Roboden"
	ctx.WindowWidth = 1920 / 2
	ctx.WindowHeight = 1080 / 2
	ctx.FullScreen = true

	state := &session.State{
		LevelOptions: session.LevelOptions{
			Resources:  2,
			Difficulty: 2,
			WorldSize:  2,
		},
		Persistent: session.PersistentData{
			// The default settings.
			Settings: session.GameSettings{
				EffectsVolumeLevel: 2,
				MusicVolumeLevel:   2,
				ScrollingSpeed:     2,
				Debug:              false,
			},
		},
	}

	assets.Register(ctx)
	controls.BindKeymap(ctx, state)

	ctx.LoadGameData("save", &state.Persistent)

	if err := ge.RunGame(ctx, menus.NewMainMenuController(state)); err != nil {
		panic(err)
	}
}
