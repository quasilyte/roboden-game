package pathing

type priorityQueue[T any] struct {
	data []priorityQueueElem[T]
}

type priorityQueueElem[T any] struct {
	Value    T
	Priority int
}

func newPriorityQueue[T any](size int) *priorityQueue[T] {
	h := &priorityQueue[T]{
		data: make([]priorityQueueElem[T], 0, size),
	}
	return h
}

func (q *priorityQueue[T]) Len() int {
	return len(q.data)
}

func (q *priorityQueue[T]) Reset() {
	q.data = q.data[:0]
}

func (q *priorityQueue[T]) Push(priority int, value T) {
	q.data = append(q.data, priorityQueueElem[T]{
		Priority: priority,
		Value:    value,
	})

	data := q.data
	i := uint(len(data) - 1)
	for {
		j := (i - 1) / 2
		if i <= j || i >= uint(len(data)) || data[i].Priority >= data[j].Priority {
			break
		}
		data[i], data[j] = data[j], data[i]
		i = j
	}
}

func (q *priorityQueue[T]) Pop() T {
	if q.Len() == 0 {
		var zero T
		return zero
	}
	data := q.data
	size := len(data) - 1
	data[0], data[size] = data[size], data[0]

	// down(0)
	i := 0
	for {
		j := 2*i + 1
		if j >= size {
			break
		}
		if j2 := j + 1; j2 < size && data[j2].Priority < data[j].Priority {
			j = j2
		}
		if data[i].Priority < data[j].Priority {
			break
		}
		data[i], data[j] = data[j], data[i]
		i = j
	}

	value := data[size].Value
	q.data = data[:size]
	return value
}
