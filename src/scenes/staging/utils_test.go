package staging

import (
	"testing"

	"github.com/quasilyte/gmath"
)

func TestResizedRect(t *testing.T) {
	rect := func(width, height float64) gmath.Rect {
		return gmath.Rect{
			Max: gmath.Vec{X: width, Y: height},
		}
	}
	pt := func(x, y float64) gmath.Vec {
		return gmath.Vec{X: x, Y: y}
	}

	tests := []struct {
		rect  gmath.Rect
		delta float64
		want  gmath.Rect
	}{
		{rect(0, 0), 0, rect(0, 0)},
		{rect(10, 10), 0, rect(10, 10)},
		{rect(10, 10), 1, gmath.Rect{Min: pt(-1, -1), Max: pt(11, 11)}},
		{rect(10, 10), 4, gmath.Rect{Min: pt(-4, -4), Max: pt(14, 14)}},
		{rect(10, 10), -1, gmath.Rect{Min: pt(1, 1), Max: pt(9, 9)}},
		{rect(10, 10), -4, gmath.Rect{Min: pt(4, 4), Max: pt(6, 6)}},
		{rect(20, 10), 2, gmath.Rect{Min: pt(-2, -2), Max: pt(22, 12)}},
		{rect(10, 20), 2, gmath.Rect{Min: pt(-2, -2), Max: pt(12, 22)}},
		{rect(20, 10), -2, gmath.Rect{Min: pt(2, 2), Max: pt(18, 8)}},
		{rect(10, 20), -2, gmath.Rect{Min: pt(2, 2), Max: pt(8, 18)}},
	}

	for i, test := range tests {
		have := resizedRect(test.rect, test.delta)
		if have != test.want {
			t.Fatalf("test[%d]: rect=%v delta=%.2f:\nhave: %v\nwant: %v",
				i, test.rect, test.delta, have, test.want)
		}
	}
}
