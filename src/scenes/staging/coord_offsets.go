package staging

import (
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/pathing"
)

var tankColonyTeleportOffsets = []gmath.Vec{
	{X: -70},
	{Y: -70},
	{X: 70},
	{Y: 70},
}

// ? ? ?
// ? x ?
// ? ? ?
var resourceNearOffsets = []pathing.GridCoord{
	{X: -1, Y: -1},
	{X: 0, Y: -1},
	{X: 1, Y: -1},
	{X: 1, Y: -1},
	{X: 1, Y: 1},
	{X: 0, Y: 1},
	{X: -1, Y: 1},
	{X: -1, Y: 0},
}

// >     ?
// >   ? ? ?
// > ? ? x ? ?
// >   ? ? ?
// >     ?
var smallColonyNearCellOffsets = []pathing.GridCoord{
	{X: -1, Y: -1},
	{X: 0, Y: -1},
	{X: 1, Y: -1},
	{X: -1, Y: 0},
	{X: 1, Y: 0},
	{X: -1, Y: 1},
	{X: 0, Y: 1},
	{X: 1, Y: 1},

	{X: 0, Y: -2},
	{X: -2, Y: 0},
	{X: 2, Y: 0},
	{X: 0, Y: 2},
}

// ? ? ? ?
// ? o o ?
// ? o x ?
// ? ? ? ?
var colonyNearCellOffsets = []pathing.GridCoord{
	{X: -2, Y: -2},
	{X: -1, Y: -2},
	{X: 0, Y: -2},
	{X: 1, Y: -2},
	{X: 1, Y: -1},
	{X: 1, Y: 0},
	{X: 1, Y: 1},
	{X: 0, Y: 1},
	{X: -1, Y: 1},
	{X: -2, Y: 1},
	{X: -2, Y: 0},
	{X: -2, Y: -1},
}

// ? ? ? ? ?
// ? o o . ?
// ? o x . ?
// ? . . . ?
// ? ? ? ? ?
var colonyNear2x2CellOffsets = []pathing.GridCoord{
	{X: -2, Y: -2},
	{X: -1, Y: -2},
	{X: 0, Y: -2},
	{X: 1, Y: -2},
	{X: 2, Y: -2},
	{X: 2, Y: -1},
	{X: 2, Y: 0},
	{X: 2, Y: 1},
	{X: 2, Y: 2},
	{X: 1, Y: 2},
	{X: 0, Y: 2},
	{X: -1, Y: 2},
	{X: -2, Y: 2},
	{X: -2, Y: 1},
	{X: -2, Y: 0},
	{X: -2, Y: -1},
}
