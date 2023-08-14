package pathing

import (
	"math/bits"
)

type priorityQueue[T any] struct {
	buckets [64][]T
	mask    uint64
}

func newPriorityQueue[T any]() *priorityQueue[T] {
	h := &priorityQueue[T]{}
	for i := range &h.buckets {
		// Start with some small capacity for every bucket.
		h.buckets[i] = make([]T, 0, 4)
	}
	return h
}

func (q *priorityQueue[T]) Reset() {
	buckets := &q.buckets

	// Reslice storage slices back.
	// To avoid traversing all len(q.buckets),
	// we have some offset to skip uninteresting (already empty) buckets.
	// We also stop when mask is 0 meaning all remaining buckets are empty too.
	// In other words, it would only touch slices between min and max non-empty priorities.
	mask := q.mask
	offset := uint(bits.TrailingZeros64(mask))
	mask >>= offset
	i := offset
	for mask != 0 {
		if i < uint(len(buckets)) {
			buckets[i] = buckets[i][:0]
		}
		mask >>= 1
		i++
	}

	q.mask = 0
}

func (q *priorityQueue[T]) IsEmpty() bool {
	return q.mask == 0
}

func (q *priorityQueue[T]) Push(priority int, value T) {
	// No bound checks since compiler knows that i will never exceed 64.
	// We also get a cool truncation of values above 64 to store them
	// in our biggest bucket.
	i := uint(priority) & 0b111111
	q.buckets[i] = append(q.buckets[i], value)
	q.mask |= 1 << i
}

func (q *priorityQueue[T]) Pop() T {
	buckets := &q.buckets

	// Using uints here and explicit len check to avoid the
	// implicitly inserted bound check.
	i := uint(bits.TrailingZeros64(q.mask))
	if i < uint(len(buckets)) {
		e := buckets[i][len(buckets[i])-1]
		buckets[i] = buckets[i][:len(buckets[i])-1]
		if len(buckets[i]) == 0 {
			q.mask &^= 1 << i
		}
		return e
	}

	// A queue is empty?
	var x T
	return x
}
