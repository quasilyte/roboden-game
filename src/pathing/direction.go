package pathing

//go:generate stringer -type=Direction -trimprefix=Dir
type Direction int

const (
	DirRight Direction = iota
	DirDown
	DirLeft
	DirUp
	DirNone // A special sentinel value
)

func (d Direction) Reversed() Direction {
	switch d {
	case DirRight:
		return DirLeft
	case DirDown:
		return DirUp
	case DirLeft:
		return DirRight
	case DirUp:
		return DirDown
	default:
		return DirNone
	}
}
