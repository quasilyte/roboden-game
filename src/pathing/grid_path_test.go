package pathing

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

func makeGridPath(directions []Direction) GridPath {
	var p GridPath
	for i := len(directions) - 1; i >= 0; i-- {
		p.push(directions[i])
	}
	p.Rewind()
	return p
}

func TestGridPathString(t *testing.T) {
	tests := []string{
		"{}",
		"{Left}",
		"{Left,Right}",
		"{Right,Left}",
		"{Down,Down,Down,Up}",
		"{Left,Right,Up,Down}",
		"{Left,Right,Right,Right,Left}",
		"{Up,Up,Down,Down,Left,Left,Right,Right,Down,Down}",
		"{Up,Up,Down,Down,Left,Left,Right,Right,Down,Down,Down,Left,Up,Right}",
		"{Left,Left,Left,Left,Left,Left,Left,Left,Left,Left,Left,Left,Left,Left}",
		"{Left,Left,Left,Left,Left,Left,Left,Left,Left,Left,Left,Left,Left,Right}",
		"{Up,Left,Left,Left,Left,Left,Left,Left,Left,Left,Left,Left,Left,Left}",
	}

	parsePath := func(s string) GridPath {
		s = s[1 : len(s)-1] // Drop "{}"
		if s == "" {
			return GridPath{}
		}
		var directions []Direction
		for _, part := range strings.Split(s, ",") {
			switch part {
			case "Right":
				directions = append(directions, DirRight)
			case "Down":
				directions = append(directions, DirDown)
			case "Left":
				directions = append(directions, DirLeft)
			case "Up":
				directions = append(directions, DirUp)
			default:
				panic("unexpected part: " + part)
			}
		}
		return makeGridPath(directions)
	}

	for _, test := range tests {
		p := parsePath(test)
		if p.String() != test {
			t.Fatalf("results mismatched:\nhave: %q\nwant: %q", p.String(), test)
		}
	}
}

func TestGridPath(t *testing.T) {
	tests := [][]Direction{
		{},
		{DirLeft},
		{DirDown},
		{DirLeft, DirRight, DirUp},
		{DirLeft, DirLeft, DirLeft},
		{DirDown, DirDown, DirDown},
		{DirDown, DirUp, DirLeft, DirRight, DirLeft, DirRight},
		{DirDown, DirLeft, DirLeft, DirLeft, DirLeft, DirDown},
		{DirRight, DirRight, DirRight, DirRight, DirRight, DirRight, DirRight},
		{DirDown, DirRight, DirRight, DirDown, DirRight, DirUp, DirRight, DirLeft},
	}

	for i, directions := range tests {
		p := makeGridPath(directions)
		reconstructed := []Direction{}
		for p.HasNext() {
			reconstructed = append(reconstructed, p.Next())
		}
		if !reflect.DeepEqual(directions, reconstructed) {
			t.Fatalf("test%d paths mismatch", i)
		}
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	for i := 0; i < 100; i++ {
		size := r.Intn(20) + 10
		directions := []Direction{}
		for j := 0; j < size; j++ {
			d := r.Intn(4)
			directions = append(directions, Direction(d))
		}
		p := makeGridPath(directions)
		reconstructed := []Direction{}
		for p.HasNext() {
			reconstructed = append(reconstructed, p.Next())
		}
		if !reflect.DeepEqual(directions, reconstructed) {
			t.Fatalf("test%d paths mismatch", i)
		}

		p.Rewind()
		reconstructed = reconstructed[:0]
		for p.HasNext() {
			reconstructed = append(reconstructed, p.Next())
		}
		if !reflect.DeepEqual(directions, reconstructed) {
			t.Fatalf("test%d paths mismatch", i)
		}
	}
}
