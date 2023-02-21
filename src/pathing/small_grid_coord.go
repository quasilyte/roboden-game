package pathing

type smallGridCoord struct {
	X int8
	Y int8
}

func (c smallGridCoord) ToGridCoord() GridCoord {
	return GridCoord{X: int(c.X), Y: int(c.Y)}
}

func (c smallGridCoord) Add(other smallGridCoord) smallGridCoord {
	return smallGridCoord{X: c.X + other.X, Y: c.Y + other.Y}
}

func (c smallGridCoord) Move(d Direction) smallGridCoord {
	switch d {
	case DirRight:
		return smallGridCoord{X: c.X + 1, Y: c.Y}
	case DirDown:
		return smallGridCoord{X: c.X, Y: c.Y + 1}
	case DirLeft:
		return smallGridCoord{X: c.X - 1, Y: c.Y}
	case DirUp:
		return smallGridCoord{X: c.X, Y: c.Y - 1}
	default:
		return c
	}
}

func (c smallGridCoord) Dist(other smallGridCoord) uint8 {
	return int8uabs(c.X-other.X) + int8uabs(c.Y-other.Y)
}

func int8uabs(x int8) uint8 {
	if x < 0 {
		return uint8(-x)
	}
	return uint8(x)
}
