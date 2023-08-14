package pathing

var neighborOffsets = [4]GridCoord{
	{X: 1},
	{Y: 1},
	{X: -1},
	{Y: -1},
}

type GreedyBFS struct {
	pqueue     *priorityQueue[weightedGridCoord]
	coordSlice []weightedGridCoord
	coordMap   *coordMap
}

type weightedGridCoord struct {
	Coord  GridCoord
	Weight int
}

func NewGreedyBFS(numCols, numRows int) *GreedyBFS {
	return &GreedyBFS{
		pqueue:     newPriorityQueue[weightedGridCoord](),
		coordMap:   newCoordMap(numCols, numRows),
		coordSlice: make([]weightedGridCoord, 0, 40),
	}
}

func (bfs *GreedyBFS) BuildPath(g *Grid, from, to GridCoord, l GridLayer) BuildPathResult {
	var result BuildPathResult
	if from == to {
		return result
	}

	// If we would like to use small (local) coordinates instead of global
	// ones, it will be necessary to translate both start and dest here.
	start := from
	goal := to

	frontier := bfs.pqueue
	frontier.Reset()

	hotFrontier := bfs.coordSlice[:0]
	hotFrontier = append(hotFrontier, weightedGridCoord{Coord: start})

	pathmap := bfs.coordMap
	pathmap.Reset()

	shortestDist := 0xff
	var fallbackCoord GridCoord
	foundPath := false
	for len(hotFrontier) != 0 || !frontier.IsEmpty() {
		var current weightedGridCoord
		if len(hotFrontier) != 0 {
			current = hotFrontier[len(hotFrontier)-1]
			hotFrontier = hotFrontier[:len(hotFrontier)-1]
		} else {
			current = frontier.Pop()
		}

		if current.Coord == goal {
			result.Steps = bfs.constructPath(start, goal, pathmap)
			foundPath = true
			break
		}
		if current.Weight >= gridPathMaxLen {
			result.Steps = bfs.constructPath(start, current.Coord, pathmap)
			result.Finish = current.Coord
			result.Partial = true
			foundPath = true
			break
		}

		dist := goal.Dist(current.Coord)
		if dist < shortestDist {
			shortestDist = dist
			fallbackCoord = current.Coord
		}
		for dir, offset := range &neighborOffsets {
			next := current.Coord.Add(offset)
			cx := uint(next.X)
			cy := uint(next.Y)
			if cx >= g.numCols || cy >= g.numRows {
				continue
			}
			if g.getCellValue(cx, cy, l) == 0 {
				continue
			}
			pathmapKey := pathmap.packCoord(next)
			if pathmap.Get(pathmapKey) != DirNone {
				continue
			}
			pathmap.Set(pathmapKey, Direction(dir))
			nextDist := goal.Dist(next)
			nextWeighted := weightedGridCoord{
				Coord:  next,
				Weight: current.Weight + 1,
			}
			if nextDist < dist {
				hotFrontier = append(hotFrontier, nextWeighted)
			} else {
				frontier.Push(nextDist, nextWeighted)
			}
		}
	}

	if !foundPath {
		result.Steps = bfs.constructPath(start, fallbackCoord, pathmap)
		result.Finish = fallbackCoord
		result.Partial = true
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
		d := pathmap.Get(pathmap.packCoord(pos))
		if pos == from {
			break
		}
		result.push(d)
		pos = pos.reversedMove(d)
	}
	return result
}
