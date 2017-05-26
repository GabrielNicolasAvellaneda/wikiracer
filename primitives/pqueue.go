package primitives

import (
	"context"

	pqueue "gopkg.in/oleiade/lane.v1"
)

// Q defines a generic queue interface with 2 methods: Enqueue, Dequeue.
type Q interface {
	Enqueue(item interface{}, priority int)
	Dequeue() (item interface{}, priority int)
}

// Pair represents a priority queue item which includes actual item and priority
type Pair struct {
	Item interface{}
	Priority int
}

// NewPQueue returns a new priority queue implementation.
func NewPQueue(ctx context.Context, dequeueChan chan<- interface{}) Q {
	q := &priorityQueue{
		queue: pqueue.NewPQueue(pqueue.MINPQ),
		dequeueChan: dequeueChan,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if q.queue.Size() > 0 {
				item, priority := q.Dequeue()
				dequeueChan <- &Pair{
					Item: item,
					Priority: priority,
				}
			}
		}
	}()

	return q
}

type priorityQueue struct {
	queue *pqueue.PQueue

	dequeueChan chan<- interface{}
}

// Enqueue adds item to a queue with priority.
func (q *priorityQueue) Enqueue(item interface{}, p int) {
	q.queue.Push(item, p)
}

// Dequeue gets the item from the queue.
func (q *priorityQueue) Dequeue() (interface{}, int) {
	return q.queue.Pop()
}
