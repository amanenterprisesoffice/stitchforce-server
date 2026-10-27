package queue

import (
	"context"
	"errors"
	"sync"
)

var ErrClosed = errors.New("queue is closed")

type MemoryQueue struct {
	ch     chan Job
	mu     sync.RWMutex
	closed bool
}

func NewMemory(size int) *MemoryQueue {
	return &MemoryQueue{ch: make(chan Job, size)}
}

func (q *MemoryQueue) Enqueue(ctx context.Context, job Job) error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return ErrClosed
	}

	select {
	case q.ch <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *MemoryQueue) Consume(ctx context.Context) (<-chan Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return nil, ErrClosed
	}

	return q.ch, nil
}

func (q *MemoryQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil
	}
	q.closed = true
	close(q.ch)
	return nil
}
