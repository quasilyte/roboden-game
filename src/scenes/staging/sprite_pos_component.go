package staging

import (
	"math"

	"github.com/quasilyte/gmath"
)

type spritePosComponent struct {
	value gmath.Vec
}

func (s *spritePosComponent) UpdatePos(pos gmath.Vec) {
	// FIXME: this should be fixed in the ge package.
	s.value.X = math.Round(pos.X)
	s.value.Y = math.Round(pos.Y)
}
