package gameui

import (
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
)

type Cursor interface {
	ClickPos(input.Action) (gmath.Vec, bool)
}
