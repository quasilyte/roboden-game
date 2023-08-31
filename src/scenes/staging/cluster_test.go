package staging

import (
	"testing"

	"github.com/quasilyte/gmath"
)

func TestFindSearchClusters(t *testing.T) {
	tests := []struct {
		clusterSize float64
		pos         gmath.Vec
		r           float64
		want        [4]int
	}{
		{10, gmath.Vec{X: 5, Y: 5}, 2, [4]int{0, 0, 0, 0}},
		{10, gmath.Vec{X: 0, Y: 0}, 9.5, [4]int{0, 0, 0, 0}},
		{10, gmath.Vec{X: 0, Y: 0}, 10, [4]int{0, 0, 0, 0}},

		{10, gmath.Vec{X: 0, Y: 0}, 10.1, [4]int{0, 0, 1, 1}},
		{10, gmath.Vec{X: 5, Y: 5}, 6, [4]int{0, 0, 1, 1}},
		{10, gmath.Vec{X: 2, Y: 5}, 6, [4]int{0, 0, 0, 1}},
		{10, gmath.Vec{X: 5, Y: 2}, 6, [4]int{0, 0, 1, 0}},

		{10, gmath.Vec{X: 10, Y: 10}, 10, [4]int{0, 0, 1, 1}},
		{10, gmath.Vec{X: 10, Y: 10}, 10.1, [4]int{0, 0, 2, 2}},
		{10, gmath.Vec{X: 10.1, Y: 10.1}, 10.1, [4]int{0, 0, 2, 2}},
		{10, gmath.Vec{X: 10.1, Y: 10.1}, 6, [4]int{0, 0, 1, 1}},
		{10, gmath.Vec{X: 19.9, Y: 19.9}, 6, [4]int{1, 1, 2, 2}},
		{10, gmath.Vec{X: 15, Y: 15}, 6, [4]int{0, 0, 2, 2}},
		{10, gmath.Vec{X: 12, Y: 15}, 6, [4]int{0, 0, 1, 2}},
		{10, gmath.Vec{X: 15, Y: 12}, 6, [4]int{0, 0, 2, 1}},

		{10, gmath.Vec{X: 15, Y: 15}, 10, [4]int{0, 0, 2, 2}},
		{10, gmath.Vec{X: 19, Y: 19}, 5, [4]int{1, 1, 2, 2}},
		{10, gmath.Vec{X: 19, Y: 19}, 20, [4]int{0, 0, 3, 3}},

		{10, gmath.Vec{X: 35, Y: 35}, 30, [4]int{0, 0, 6, 6}},
		{10, gmath.Vec{X: 45, Y: 45}, 30, [4]int{1, 1, 7, 7}},

		{10, gmath.Vec{X: 35, Y: 35}, 50, [4]int{0, 0, 7, 7}},
		{10, gmath.Vec{X: 45, Y: 45}, 50, [4]int{0, 0, 7, 7}},
		{10, gmath.Vec{X: 35, Y: 35}, 100, [4]int{0, 0, 7, 7}},
		{10, gmath.Vec{X: 45, Y: 45}, 100, [4]int{0, 0, 7, 7}},
		{10, gmath.Vec{X: 40, Y: 40}, 100, [4]int{0, 0, 7, 7}},
		{10, gmath.Vec{X: 2, Y: 5}, 100, [4]int{0, 0, 7, 7}},
		{10, gmath.Vec{X: 2, Y: 5}, 500, [4]int{0, 0, 7, 7}},
		{10, gmath.Vec{X: 2, Y: 5}, 1000, [4]int{0, 0, 7, 7}},
	}

	for i, test := range tests {
		world := &worldState{
			creepClusterSize: test.clusterSize,
		}
		world.creepClusterMultiplier = 1.0 / world.creepClusterSize
		startX, startY, endX, endY := world.findSearchClusters(test.pos, test.r)
		have := [4]int{startX, startY, endX, endY}
		if test.want != have {
			t.Fatalf("test[%d] size=%f pos=%v r=%f\nhave: %v\nwant: %v",
				i, test.clusterSize, test.pos, test.r, have, test.want)
		}
	}
}
