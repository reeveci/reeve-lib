package queue

type Queue[V any] interface {
	Get() V
	Pop() V
	Push(V)
	Count() uint
}
