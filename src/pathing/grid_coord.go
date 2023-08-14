package pathing

type GridCoord struct {
	X int
	Y int
}

func (c GridCoord) IsZero() bool {
	return c.X == 0 && c.Y == 0
}

func (c GridCoord) Add(other GridCoord) GridCoord {
	return GridCoord{X: c.X + other.X, Y: c.Y + other.Y}
}

func (c GridCoord) reversedMove(d Direction) GridCoord {
	switch d {
	case DirRight:
		return GridCoord{X: c.X - 1, Y: c.Y}
	case DirDown:
		return GridCoord{X: c.X, Y: c.Y - 1}
	case DirLeft:
		return GridCoord{X: c.X + 1, Y: c.Y}
	case DirUp:
		return GridCoord{X: c.X, Y: c.Y + 1}
	default:
		return c
	}
}

func (c GridCoord) Move(d Direction) GridCoord {
	switch d {
	case DirRight:
		return GridCoord{X: c.X + 1, Y: c.Y}
	case DirDown:
		return GridCoord{X: c.X, Y: c.Y + 1}
	case DirLeft:
		return GridCoord{X: c.X - 1, Y: c.Y}
	case DirUp:
		return GridCoord{X: c.X, Y: c.Y - 1}
	default:
		return c
	}
}

func (c GridCoord) Dist(other GridCoord) int {
	return intabs(c.X-other.X) + intabs(c.Y-other.Y)
}

func intabs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
