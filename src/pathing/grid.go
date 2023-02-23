package pathing

import (
	"github.com/quasilyte/gmath"
)

const (
	CellSize float64 = 32
)

type Grid struct {
	worldWidth  float64
	worldHeight float64

	numCols uint
	numRows uint

	bytes []byte
}

func NewGrid(worldWidth, worldHeight float64) *Grid {
	g := &Grid{
		worldWidth:  worldWidth,
		worldHeight: worldHeight,
	}

	g.numCols = uint(g.worldWidth / CellSize)
	g.numRows = uint(g.worldHeight / CellSize)

	numCells := g.numCols * g.numRows
	numBytes := numCells / 8
	if numCells%8 != 0 {
		numBytes++
	}
	g.bytes = make([]byte, numBytes)

	return g
}

func (g *Grid) Size() (numCols, numRows int) {
	return int(g.numCols), int(g.numRows)
}

func (g *Grid) SetCell(c GridCoord, v bool) {
	if v {
		g.MarkCell(c)
	} else {
		g.UnmarkCell(c)
	}
}

func (g *Grid) UnmarkCell(c GridCoord) {
	i := uint(c.Y)*g.numCols + uint(c.X)
	byteIndex := i / 8
	bitIndex := i % 8
	if byteIndex < uint(len(g.bytes)) {
		g.bytes[byteIndex] &^= 1 << bitIndex
	}
}

func (g *Grid) MarkCell(c GridCoord) {
	i := uint(c.Y)*g.numCols + uint(c.X)
	byteIndex := i / 8
	bitIndex := i % 8
	if byteIndex < uint(len(g.bytes)) {
		g.bytes[byteIndex] |= 1 << bitIndex
	}
}

func (g *Grid) CellIsFree(c GridCoord) bool {
	x := uint(c.X)
	y := uint(c.Y)
	if x >= g.numCols || y >= g.numRows {
		return false
	}

	i := y*g.numCols + x
	byteIndex := i / 8
	bitIndex := i % 8
	if byteIndex < uint(len(g.bytes)) {
		return (g.bytes[byteIndex] & (1 << bitIndex)) == 0
	}
	// Consider out of bound cells as marked.
	return false
}

func (g *Grid) AlignPos(pos gmath.Vec) gmath.Vec {
	return g.CoordToPos(g.PosToCoord(pos))
}

func (g *Grid) PosToCoord(pos gmath.Vec) GridCoord {
	x := int(pos.X) / int(CellSize)
	y := int(pos.Y) / int(CellSize)
	return GridCoord{x, y}
}

func (g *Grid) CoordToPos(cell GridCoord) gmath.Vec {
	return gmath.Vec{
		X: (float64(cell.X) * CellSize) + (CellSize / 2),
		Y: (float64(cell.Y) * CellSize) + (CellSize / 2),
	}
}
