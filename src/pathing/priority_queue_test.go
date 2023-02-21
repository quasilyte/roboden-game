package pathing

import (
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"
)

func BenchmarkPriorityQueue(b *testing.B) {
	q := newPriorityQueue[int](20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Reset()
		q.Push(1, 10)
		q.Push(2, 10)
		q.Push(3, 10)
		q.Push(4, 10)
		q.Push(0, 10)
		q.Push(1, 10)
		q.Push(10, 10)
		q.Push(0, 10)
		q.Push(100, 10)
		q.Push(100, 10)
	}
}

func TestPriorityQueue(t *testing.T) {
	var pqueue priorityQueue[byte]

	tests := []struct {
		input  string
		output string
	}{
		{"a", "a"},
		{"aaa", "aaa"},
		{"aa", "aa"},
		{"ba", "ab"},
		{"baa", "aab"},
		{"ab", "ab"},
		{"aaba", "aaab"},
		{"abcd", "abcd"},
		{"dcab", "abcd"},
	}

	for _, test := range tests {
		pqueue.Reset()
		for i := 0; i < len(test.input); i++ {
			b := test.input[i]
			pqueue.Push(int(b), b)
		}
		var output strings.Builder
		for pqueue.Len() != 0 {
			output.WriteByte(pqueue.Pop())
		}
		if output.String() != test.output {
			t.Fatalf("input=%q:\nhave: %q\nwant: %q", test.input, output.String(), test.output)
		}
	}

	{
		var q priorityQueue[int]
		for i := 0; i < 50; i++ {
			q.Push(i, i)
		}
		for i := 0; i < 50; i++ {
			result := q.Pop()
			if i != result {
				t.Fatal("invalid result in push+pop pair")
			}
		}
	}

	for i := 0; i < 64; i++ {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		var values []int
		num := r.Intn(96) + 6
		for i := 0; i < num; i++ {
			values = append(values, r.Int())
		}
		sortedValues := make([]int, len(values))
		copy(sortedValues, values)
		sort.SliceStable(sortedValues, func(i, j int) bool {
			return sortedValues[i] < sortedValues[j]
		})
		var q priorityQueue[int]
		for _, x := range values {
			q.Push(x, x)
		}
		for _, x := range sortedValues {
			y := q.Pop()
			if x != y {
				t.Fatal("invalid result in push+pop pair")
			}
		}
	}
}
