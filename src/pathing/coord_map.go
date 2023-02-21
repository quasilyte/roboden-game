package pathing

// The sparse/dense idea is described here: https://research.swtch.com/sparse
//
// We use a fact that grid coords are usually small and can be packed as
// simple integers if we want to.
// So, a GridCoord{x, y} (a pair of 2 ints) can be translated to
// uint16 value which can be interpreted as a packed coordinate.

type coordMap struct {
	dense   []coordMapElem
	sparse  []uint16
	numRows int
	numCols int
}

type coordMapElem struct {
	key   uint16
	value Direction
}

func newCoordMap(numRows, numCols int) *coordMap {
	size := numRows * numCols
	return &coordMap{
		dense:   make([]coordMapElem, 0, size/8),
		sparse:  make([]uint16, size),
		numRows: numRows,
		numCols: numCols,
	}
}

func (m *coordMap) Cap() int {
	return len(m.sparse)
}

func (m *coordMap) Len() int {
	return len(m.dense)
}

func (m *coordMap) Get(k uint) Direction {
	if k < uint(len(m.sparse)) {
		i := uint(m.sparse[k])
		if i < uint(len(m.dense)) && uint(m.dense[i].key) == k {
			return m.dense[i].value
		}
	}
	return DirNone
}

func (m *coordMap) Set(k uint, d Direction) {
	sparse := m.sparse
	if k < uint(len(sparse)) {
		i := uint(sparse[k])
		if i < uint(len(m.dense)) && uint(m.dense[i].key) == k {
			m.dense[i].value = d
			return
		}
		// Insert a new value.
		m.dense = append(m.dense, coordMapElem{uint16(k), d})
		sparse[k] = uint16(len(m.dense)) - 1
	}
}

func (m *coordMap) Reset() {
	m.dense = m.dense[:0]
}

func (s *coordMap) packCoord(c GridCoord) uint {
	return uint((c.Y * s.numCols) + c.X)
}
