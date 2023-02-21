package pathing

type GreedyBFS struct {
	pqueue     *priorityQueue[GridCoord]
	coordSlice []GridCoord
	coordMap   *coordMap
}

func NewGreedyBFS(numRows, numCols int) *GreedyBFS {
	return &GreedyBFS{
		pqueue:     newPriorityQueue[GridCoord](80),
		coordMap:   newCoordMap(numRows, numCols),
		coordSlice: make([]GridCoord, 0, 40),
	}
}

func (bfs *GreedyBFS) BuildPath(g *Grid, from, to GridCoord) BuildPathResult {
	var result BuildPathResult
	if from == to {
		result.Complete = true
		return result
	}

	// Translate the world (grid) coordinates to our local coordinates.
	// Start will be centered.
	// The goal coordinate will be truncated, if necessary.
	start := from
	goal := to

	frontier := bfs.pqueue
	frontier.Reset()

	hotFrontier := bfs.coordSlice[:0]
	hotFrontier = append(hotFrontier, start)

	pathmap := bfs.coordMap
	pathmap.Reset()

	for len(hotFrontier)+frontier.Len() != 0 {
		var current GridCoord
		if len(hotFrontier) != 0 {
			current = hotFrontier[len(hotFrontier)-1]
			hotFrontier = hotFrontier[:len(hotFrontier)-1]
		} else {
			current = frontier.Pop()
		}

		if current == goal {
			result.Steps = bfs.constructPath(start, goal, pathmap)
			result.Complete = true
			break
		}
		dist := goal.Dist(current)
		for dir := DirRight; dir <= DirUp; dir++ {
			// TODO: handle origin point.
			next := current.Move(dir)
			if !g.CellIsFree(next) {
				continue
			}
			if pathmap.Get(next) != DirNone {
				continue
			}

			pathmap.Set(next, dir)
			nextDist := goal.Dist(next)
			if nextDist < dist {
				hotFrontier = append(hotFrontier, next)
			} else {
				frontier.Push(nextDist, next)
			}
		}
	}

	// In case if that slice was growing due to appends,
	// save that extra capacity for later.
	bfs.coordSlice = hotFrontier[:0]

	return result
}

func (bfs *GreedyBFS) constructPath(from, to GridCoord, pathmap *coordMap) GridPath {
	// We walk from the finish point towards the start.
	// The directions are pushed in that order and would lead
	// to a reversed path, but since GridPath does its iteration
	// in reversed order itself, we don't need to do any
	// post-build reversal here.
	var result GridPath
	pos := to
	for {
		d := pathmap.Get(pos)
		if pos == from {
			break
		}
		result.push(d)
		pos = pos.Move(d.Reversed())
	}
	return result
}
