//go:build android

package mobilegame

import (
	"github.com/quasilyte/roboden-game/cmd/internal/game"
)

func init() {
	game.Main()
}

func Dummy() {}
