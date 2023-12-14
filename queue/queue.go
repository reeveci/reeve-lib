package queue

func NewQueue[V any]() Queue[V] {
	return &queue[V]{}
}

type queue[V any] struct {
	head, bottom *entry[V]
	count        uint
}

type entry[V any] struct {
	Value V
	Next  *entry[V]
}

func (q *queue[V]) Get() (result V) {
	if q == nil || q.head == nil {
		return
	}

	result = q.head.Value
	return
}

func (q *queue[V]) Pop() (result V) {
	if q == nil || q.head == nil {
		return
	}

	result = q.head.Value
	q.head = q.head.Next
	if q.head == nil {
		q.bottom = nil
	}
	if q.count > 0 {
		q.count -= 1
	}
	return
}

func (q *queue[V]) Push(value V) {
	bottom := &entry[V]{value, nil}
	if q.bottom != nil {
		q.bottom.Next = bottom
	}
	q.bottom = bottom
	if q.head == nil {
		q.head = bottom
	}
	q.count += 1
}

func (q *queue[V]) Count() uint {
	return q.count
}
