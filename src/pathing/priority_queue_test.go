package pathing

import (
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"
)

func ensureEmpty[T any](t *testing.T, q *priorityQueue[T]) {
	t.Helper()
	if !q.IsEmpty() || q.mask != 0 {
		t.Fatal("queue is not empty")
	}
	for i, b := range &q.buckets {
		if len(b) != 0 {
			t.Fatalf("buckets[%d] is not empty", i)
		}
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
		{"aaaaaaa", "aaaaaaa"},
		{"aaaaaab", "aaaaaab"},
		{"baaaaaa", "aaaaaab"},
		{"baaaaab", "aaaaabb"},
		{"aaaaaaaaaaaaaaa", "aaaaaaaaaaaaaaa"},
		{"ababababab", "aaaaabbbbb"},
		{"abcabcabcabc", "aaaabbbbcccc"},
		{"abcdabcdabcdabcd", "aaaabbbbccccdddd"},
		{"aabbccddaabbccdd", "aaaabbbbccccdddd"},
	}

	for _, test := range tests {
		pqueue.Reset()
		for i := 0; i < len(test.input); i++ {
			b := test.input[i]
			priority := int(b) - 'a'
			pqueue.Push(priority, b)
		}
		var output strings.Builder
		for !pqueue.IsEmpty() {
			output.WriteByte(pqueue.Pop())
		}
		ensureEmpty(t, &pqueue)
		if output.String() != test.output {
			t.Fatalf("input=%q:\nhave: %q\nwant: %q", test.input, output.String(), test.output)
		}
	}

	{
		q := newPriorityQueue[int]()
		ensureEmpty(t, q)
		for i := 0; i < 50; i++ {
			q.Push(i, i)
		}
		for i := 0; i < 50; i++ {
			result := q.Pop()
			if i != result {
				t.Fatal("invalid result in push+pop pair")
			}
		}
		ensureEmpty(t, q)
	}

	for i := 0; i < 64; i++ {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		var values []int
		num := r.Intn(96) + 6
		for i := 0; i < num; i++ {
			maxValue := gridPathMaxLen
			minValue := 0
			v := rand.Intn(maxValue-minValue) + minValue
			values = append(values, v)
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
		ensureEmpty(t, &q)
	}
}
