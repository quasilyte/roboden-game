package main

import (
	"time"

	"github.com/quasilyte/colony-game/assets"
	"github.com/quasilyte/colony-game/controls"
	"github.com/quasilyte/colony-game/scenes/staging"
	"github.com/quasilyte/colony-game/session"
	"github.com/quasilyte/ge"
)

func main() {
	ctx := ge.NewContext()
	ctx.Rand.SetSeed(time.Now().Unix())
	ctx.GameName = "colony_game"
	ctx.WindowTitle = "Colony"
	ctx.WindowWidth = 1920 / 2
	ctx.WindowHeight = 1080 / 2
	ctx.FullScreen = true

	state := &session.State{}

	assets.Register(ctx)
	controls.BindKeymap(ctx, state)

	if err := ge.RunGame(ctx, staging.NewController(state)); err != nil {
		panic(err)
	}
}
