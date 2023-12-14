package queue

import "sync"

type blockedQueue[V any] struct {
	queue       Queue[V]
	lock, rLock sync.Mutex
	waitChan    chan bool
}

func Blocked[V any](q Queue[V]) Queue[V] {
	return &blockedQueue[V]{queue: q, waitChan: make(chan bool)}
}

func (q *blockedQueue[V]) Get() V {
	q.rLock.Lock()
	defer q.rLock.Unlock()

	q.lock.Lock()
	defer q.lock.Unlock()

	if q.queue.Count() > 0 {
		return q.queue.Get()
	}

	for {
		q.lock.Unlock()
		<-q.waitChan

		q.lock.Lock()
		if q.queue.Count() > 0 {
			return q.queue.Get()
		}
	}
}

func (q *blockedQueue[V]) Pop() V {
	q.rLock.Lock()
	defer q.rLock.Unlock()

	q.lock.Lock()
	defer q.lock.Unlock()

	if q.queue.Count() > 0 {
		return q.queue.Pop()
	}

	for {
		q.lock.Unlock()
		<-q.waitChan

		q.lock.Lock()
		if q.queue.Count() > 0 {
			return q.queue.Pop()
		}
	}
}

func (q *blockedQueue[V]) Push(value V) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.queue.Push(value)

	select {
	case q.waitChan <- true:
	default:
	}
}

func (q *blockedQueue[V]) Count() uint {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.queue.Count()
}
