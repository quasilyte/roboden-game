package pathing_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/pathing"
)

func TestEmptyGrid(t *testing.T) {
	p := pathing.NewGrid(0, 0, 0)
	cols, rows := p.Size()
	if rows != 0 || cols != 0 {
		t.Fatalf("expected [0,0] size, got [%d,%d]", cols, rows)
	}

	positions := []gmath.Vec{
		{X: 0, Y: 0},
		{X: 98, Y: 0},
		{X: 0, Y: 98},
		{X: -98, Y: 0},
		{X: 0, Y: -98},
		{X: 2045, Y: 3525},
		{X: -2045, Y: -3525},
	}

	l := pathing.MakeGridLayer(1, 0, 1, 1)
	for _, pos := range positions {
		if p.GetCellValue(p.PosToCoord(pos), l) != 0 {
			t.Fatalf("empty grid reported %v as free", pos)
		}
	}
}

func TestGridOutOfBounds(t *testing.T) {
	p := pathing.NewGrid(4*pathing.CellSize, 4*pathing.CellSize, 0)
	cols, rows := p.Size()
	if rows != 4 || cols != 4 {
		t.Fatalf("expected [4,4] size, got [%d,%d]", cols, rows)
	}

	coords := []pathing.GridCoord{
		{X: 0, Y: -1},
		{X: -1, Y: -1},
		{X: -1, Y: 0},
		{X: -40, Y: -40},

		{X: 4, Y: 0},
		{X: 5, Y: 0},
		{X: 50, Y: 0},
		{X: 0, Y: 4},
		{X: 0, Y: 5},
		{X: 0, Y: 50},
		{X: 4, Y: 4},
		{X: 5, Y: 5},
		{X: 50, Y: 50},

		{X: 2, Y: 10},
		{X: 3, Y: 10},
		{X: 10, Y: 2},
		{X: 10, Y: 3},
		{X: 2, Y: -10},
		{X: 3, Y: -10},
		{X: -10, Y: 2},
		{X: -10, Y: 3},
	}

	l := pathing.MakeGridLayer(1, 0, 1, 1)
	for _, coord := range coords {
		if p.GetCellValue(coord, l) != 0 {
			t.Fatalf("grid reported out-of-bounds %v as free", coord)
		}
	}
}

func TestGridMaps(t *testing.T) {
	tests := [][]string{
		{
			"A.............x...............................x.....x.....x....",
			"..............x.......x......xxxxxxxxxx.......x..x..x..x..x....",
			"....xxxxxxxxxxx.......x...............x.......x..x..x..x..x....",
			"......................x...............x..........x.....x......B",
		},

		{
			"A.............x..........x................x.x..x...x.x...x.....x....",
			"..............x........x.x....xxxxxxxxxx..x.xxxx...x.xx..x..x..x....",
			"....xxxxxxxxxxx...xx...x.x.............x..x.x..x...x.xx..x..x..x....",
			"..................xx...x.x.............x..x.x..x.....xx.....x......B",
		},
	}

	l := pathing.MakeGridLayer(1, 0, 1, 1)
	for i, test := range tests {
		parsed := testParseGrid(t, test)
		for row := 0; row < parsed.numRows; row++ {
			for col := 0; col < parsed.numCols; col++ {
				marker := test[row][col]
				cell := pathing.GridCoord{X: col, Y: row}
				switch marker {
				case 'x':
					if parsed.grid.GetCellValue(cell, l) != 0 {
						t.Fatalf("test%d: x cell is reported as free", i)
					}
				case '.', ' ':
					if parsed.grid.GetCellValue(cell, l) != 1 {
						t.Fatalf("test%d: empty/free cell is reported as marked", i)
					}
				}
			}
		}
	}
}

func TestRandFillGrid(t *testing.T) {
	p := pathing.NewGrid(10*pathing.CellSize, 10*pathing.CellSize, 0)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	layers := make([][]uint8, 10)
	for i := range layers {
		layers[i] = make([]uint8, 10)
		for j := range layers[i] {
			layers[i][j] = uint8(rng.Int63n(4))
		}
	}

	values := []uint8{10, 0, 20, 30}
	values2 := []uint8{0, 1, 2, 3}
	l := pathing.MakeGridLayer(values[0], values[1], values[2], values[3])
	l2 := pathing.MakeGridLayer(values2[0], values2[1], values2[2], values2[3])
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			c := pathing.GridCoord{X: x, Y: y}
			tag := layers[y][x]
			p.SetCellTag(c, tag)
			v := p.GetCellValue(c, l)
			if v != values[tag] {
				t.Fatalf("grid[%d][%d] value mismatch: have %v, want %v", y, x, v, values[tag])
			}
			v2 := p.GetCellValue(c, l2)
			if v2 != values2[tag] {
				t.Fatalf("grid[%d][%d] value2 mismatch: have %v, want %v", y, x, v2, values2[tag])
			}
		}
	}
}

func TestGridValueChange(t *testing.T) {
	p := pathing.NewGrid(4*pathing.CellSize, 4*pathing.CellSize, 0)
	layerValues := []uint8{1, 0, 5, 10}
	l := pathing.MakeGridLayer(layerValues[0], layerValues[1], layerValues[2], layerValues[3])
	coord := pathing.GridCoord{X: 1, Y: 1}

	probes := []uint8{
		3,
		3,
		0,
		1,
		3,
		2,
		1,
		0,
		3,
		3,
		0,
		0,
		1,
		0,
	}

	for _, probe := range probes {
		want := layerValues[probe]
		p.SetCellTag(coord, probe)
		if have := p.GetCellValue(coord, l); have != want {
			t.Fatalf("SetCellTag(%v, %v): have %v, want %v", coord, probe, have, want)
		}
	}
}

func TestSmallGrid(t *testing.T) {
	p := pathing.NewGrid(9*pathing.CellSize, 6*pathing.CellSize, 0)

	numCols, numRows := p.Size()
	if numCols != 9 || numRows != 6 {
		t.Fatalf("expected [9,6] size, got [%d,%d]", numCols, numRows)
	}

	values := []uint8{10, 0, 20, 30}
	l := pathing.MakeGridLayer(values[0], values[1], values[2], values[3])
	numCells := numCols * numRows
	for y := 0; y < numRows; y++ {
		for x := 0; x < numCols; x++ {
			c := pathing.GridCoord{X: x, Y: y}
			if p.GetCellValue(c, l) != 10 {
				t.Fatalf("empty grid (size %d) reports in-bounds %v as marked", numCells, c)
			}
		}
	}

	for y := 0; y < numRows; y++ {
		for x := 0; x < numCols; x++ {
			c := pathing.GridCoord{X: x, Y: y}
			tag := uint8((y*numCols + x) % 4)
			p.SetCellTag(c, tag)
		}
	}

	for y := 0; y < numRows; y++ {
		for x := 0; x < numCols; x++ {
			c := pathing.GridCoord{X: x, Y: y}
			tag := uint8((y*numCols + x) % 4)
			v := values[tag]
			if actual := p.GetCellValue(c, l); actual != v {
				t.Fatalf("expected %v value to be %v, got %v", c, v, actual)
			}
		}
	}
}

func TestGrid(t *testing.T) {
	p := pathing.NewGrid(1856, 1856, 0)

	tests := []pathing.GridCoord{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
		{X: 4, Y: 0},
		{X: 0, Y: 4},
		{X: 8, Y: 0},
		{X: 0, Y: 8},
		{X: 9, Y: 0},
		{X: 0, Y: 9},
		{X: 9, Y: 9},
		{X: 30, Y: 31},
		{X: 31, Y: 30},
		{X: 0, Y: 14},
		{X: 14, Y: 0},
	}

	l := pathing.MakeGridLayer(0, 1, 2, 3)
	for i, test := range tests {
		if p.GetCellValue(test, l) != 0 {
			t.Fatalf("GetCellValue(%d, %d) returned true before it was set", test.X, test.Y)
		}
		p.SetCellTag(test, 1)
		if p.GetCellValue(test, l) != 1 {
			t.Fatalf("GetCellValue(%d, %d) returned false after it was set", test.X, test.Y)
		}
		for j := i + 1; j < len(tests); j++ {
			otherTest := tests[j]
			if p.GetCellValue(otherTest, l) != 0 {
				t.Fatalf("unrelated GetCellValue(%d, %d) returned true after (%d, %d) was set", otherTest.X, otherTest.Y, test.X, test.Y)
			}
		}
	}
}
