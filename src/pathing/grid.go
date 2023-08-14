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

func NewGrid(worldWidth, worldHeight float64, defaultTag uint8) *Grid {
	g := &Grid{
		worldWidth:  worldWidth,
		worldHeight: worldHeight,
	}

	g.numCols = uint(g.worldWidth / CellSize)
	g.numRows = uint(g.worldHeight / CellSize)

	numCells := g.numCols * g.numRows
	numBytes := numCells / 4
	if numCells%4 != 0 {
		numBytes++
	}
	b := make([]byte, numBytes)

	defaultTag &= 0b11
	if defaultTag != 0 {
		v := uint8(0)
		switch defaultTag {
		case 1:
			v = 0b01010101
		case 2:
			v = 0b10101010
		case 3:
			v = 0b11111111
		}
		for i := range b {
			b[i] = v
		}
	}

	g.bytes = b

	return g
}

func (g *Grid) Size() (numCols, numRows int) {
	return int(g.numCols), int(g.numRows)
}

func (g *Grid) SetCellTag(c GridCoord, tag uint8) {
	i := uint(c.Y)*g.numCols + uint(c.X)
	byteIndex := i / 4
	if byteIndex < uint(len(g.bytes)) {
		shift := (i % 4) * 2
		b := g.bytes[byteIndex]
		b &^= 0b11 << shift        // Clear the two data bits
		b |= (tag & 0b11) << shift // Mix it with provided bits
		g.bytes[byteIndex] = b
	}
}

func (g *Grid) GetCellValue(c GridCoord, l GridLayer) uint8 {
	x := uint(c.X)
	y := uint(c.Y)
	if x >= g.numCols || y >= g.numRows {
		// Consider out of bound cells as blocked.
		return 0
	}
	return g.getCellValue(x, y, l)
}

func (g *Grid) getCellValue(x, y uint, l GridLayer) uint8 {
	i := y*g.numCols + x
	byteIndex := i / 4
	shift := (i % 4) * 2
	tag := ((readByte(g.bytes, byteIndex)) >> shift) & 0b11
	return l.getFast(tag)
}

func (g *Grid) AlignPos(pos gmath.Vec) gmath.Vec {
	return g.CoordToPos(g.PosToCoord(pos))
}

func (g *Grid) AlignPos2x2(pos gmath.Vec) gmath.Vec {
	alignedPos := g.AlignPos(pos)
	remX := int(pos.X) % int(CellSize)
	remY := int(pos.Y) % int(CellSize)
	if remX < int(CellSize)/2 {
		alignedPos.X -= 16
	} else {
		alignedPos.X += 16
	}
	if remY < int(CellSize)/2 {
		alignedPos.Y -= 16
	} else {
		alignedPos.Y += 16
	}
	return alignedPos
}

func (g *Grid) IndexToCoord(index int) GridCoord {
	u32 := uint32(index)
	x := int(u32 & 0xffff)
	y := int(u32 >> 16)
	return GridCoord{X: x, Y: y}
}

func (g *Grid) CoordToIndex(cell GridCoord) int {
	u32 := uint32(cell.X) | uint32(cell.Y<<16)
	return int(u32)
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
