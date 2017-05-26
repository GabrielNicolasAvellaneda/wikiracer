package primitives

import (
	"context"

	queue "github.com/scryner/lfreequeue"
)

// Q defines a generic queue interface with 2 methods: Enqueue, Dequeue.
type Q interface {
	Enqueue(interface{})
	Dequeue() interface{}
}

// NewWatchIterQueue is a Queue constructor. ctx is used to control queue iterator, dequeueChan is used to send back the
// items as them appear in the queue.
func NewWatchIterQueue(ctx context.Context, dequeueChan chan<- interface{}) Q {
	q := &Queue{
		queue.NewQueue(),
		dequeueChan,
	}

	go func() {
		watchIterator := q.queue.NewWatchIterator()
		iter := watchIterator.Iter()
		defer watchIterator.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case item := <-iter:
				dequeueChan <- item
			}
		}
	}()
	return q
}

// Queue is a simple async FIFO queue data structure which implements Q interface.
// If Queue is constructed with NewWatchIterQueue, the user can receive the data asynchronously with dequeueChan.
type Queue struct {
	queue *queue.Queue

	dequeueChan chan<- interface{}
}

// Enqueue appends an item to current queue.
func (q *Queue) Enqueue(item interface{}) {
	q.queue.Enqueue(item)
}

// Dequeue returns an item from a queue.
func (q *Queue) Dequeue() interface{} {
	return q.Dequeue()
}
