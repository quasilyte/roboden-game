package pathing_test

import (
	"testing"

	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/pathing"
)

func TestEmptyGrid(t *testing.T) {
	p := pathing.NewGrid(0, 0)
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

	for _, pos := range positions {
		if p.CellIsFree(p.PosToCoord(pos)) {
			t.Fatalf("empty grid reported %v as free", pos)
		}
	}
}

func TestGridOutOfBounds(t *testing.T) {
	p := pathing.NewGrid(4*pathing.CellSize, 4*pathing.CellSize)
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

	for _, coord := range coords {
		if p.CellIsFree(coord) {
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

	for i, test := range tests {
		parsed := testParseGrid(t, test)
		for row := 0; row < parsed.numRows; row++ {
			for col := 0; col < parsed.numCols; col++ {
				marker := test[row][col]
				cell := pathing.GridCoord{X: col, Y: row}
				switch marker {
				case 'x':
					if parsed.grid.CellIsFree(cell) {
						t.Fatalf("test%d: x cell is reported as free", i)
					}
				case '.', ' ':
					if !parsed.grid.CellIsFree(cell) {
						t.Fatalf("test%d: empty/free cell is reported as marked", i)
					}
				}
			}
		}
	}
}

func TestSmallGrid(t *testing.T) {
	p := pathing.NewGrid(9*pathing.CellSize, 6*pathing.CellSize)

	numCols, numRows := p.Size()
	if numCols != 9 || numRows != 6 {
		t.Fatalf("expected [9,6] size, got [%d,%d]", numCols, numRows)
	}

	numCells := numCols * numRows
	for y := 0; y < numRows; y++ {
		for x := 0; x < numCols; x++ {
			c := pathing.GridCoord{X: x, Y: y}
			if !p.CellIsFree(c) {
				t.Fatalf("empty grid (size %d) reports in-bounds %v as marked", numCells, c)
			}
		}
	}

	for y := 0; y < numRows; y++ {
		for x := 0; x < numCols; x++ {
			c := pathing.GridCoord{X: x, Y: y}
			p.MarkCell(c)
		}
	}

	for y := 0; y < numRows; y++ {
		for x := 0; x < numCols; x++ {
			c := pathing.GridCoord{X: x, Y: y}
			if p.CellIsFree(c) {
				t.Fatalf("fully-marked grid (size %d) reports in-bounds %v as unmarked", numCells, c)
			}
		}
	}
}

func TestGrid(t *testing.T) {
	p := pathing.NewGrid(1856, 1856)

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

	for i, test := range tests {
		if !p.CellIsFree(test) {
			t.Fatalf("CheckCell(%d, %d) returned true before it was set", test.X, test.Y)
		}
		p.MarkCell(test)
		if p.CellIsFree(test) {
			t.Fatalf("CheckCell(%d, %d) returned false after it was set", test.X, test.Y)
		}
		for j := i + 1; j < len(tests); j++ {
			otherTest := tests[j]
			if !p.CellIsFree(otherTest) {
				t.Fatalf("unrelated CheckCell(%d, %d) returned true after (%d, %d) was set", otherTest.X, otherTest.Y, test.X, test.Y)
			}
		}
	}
}
