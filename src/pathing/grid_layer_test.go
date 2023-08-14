package pathing

import "testing"

func TestGridLayer(t *testing.T) {
	tests := [][]uint8{
		{0, 0, 0, 0},
		{0, 0, 0, 1},
		{1, 0, 0, 0},
		{1, 1, 1, 1},
		{10, 0, 10, 0},
		{1, 2, 3, 4},
		{4, 3, 2, 1},
		{0xff, 0xff, 0xff, 0xff},
		{100, 0xff, 0xff, 100},
		{24, 53, 21, 99},
		{99, 145, 9, 0},
	}

	for _, test := range tests {
		l := MakeGridLayer(test[0], test[1], test[2], test[3])
		for i := uint8(0); i <= 3; i++ {
			want := test[i]
			have := l.Get(i)
			if want != have {
				t.Fatalf("(%v).Get(%d): have %v, want %v", test, i, have, want)
			}
		}
	}
}
