package staging

import "github.com/quasilyte/roboden-game/pathing"

const (
	ptagFree    uint8 = 0
	ptagBlocked uint8 = 1
	ptagForest  uint8 = 2
	ptagUnused  uint8 = 3
)

var (
	layerNormal     = pathing.MakeGridLayer(1, 0, 1, 0)
	layerLandColony = pathing.MakeGridLayer(1, 0, 0, 0)
)
