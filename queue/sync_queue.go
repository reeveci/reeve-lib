package queue

import "sync"

type syncQueue[V any] struct {
	queue Queue[V]
	lock  sync.Mutex
}

func Sync[V any](q Queue[V]) Queue[V] {
	return &syncQueue[V]{queue: q}
}

func (q *syncQueue[V]) Get() V {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.queue.Get()
}

func (q *syncQueue[V]) Pop() V {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.queue.Pop()
}

func (q *syncQueue[V]) Push(value V) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.queue.Push(value)
}

func (q *syncQueue[V]) Count() uint {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.queue.Count()
}
