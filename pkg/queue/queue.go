package queue

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gitverse-notifier/pkg/events"
)

var ErrQueueFull = errors.New("event queue is full")

type EventQueue interface {
	Enqueue(ctx context.Context, event *events.Event) error
	EnqueueBlocking(ctx context.Context, event *events.Event) error
	Events() <-chan *events.Event
	Close()
}

type MemoryQueue struct {
	ch        chan *events.Event
	closeOnce sync.Once
}

func NewMemoryQueue(size int) *MemoryQueue {
	return &MemoryQueue{ch: make(chan *events.Event, size)}
}

func (q *MemoryQueue) Enqueue(ctx context.Context, event *events.Event) error {
	select {
	case q.ch <- event:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("enqueue cancelled: %w", ctx.Err())
	default:
		return ErrQueueFull
	}
}

func (q *MemoryQueue) EnqueueBlocking(ctx context.Context, event *events.Event) error {
	select {
	case q.ch <- event:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("enqueue cancelled: %w", ctx.Err())
	}
}

func (q *MemoryQueue) Events() <-chan *events.Event {
	return q.ch
}

func (q *MemoryQueue) Close() {
	q.closeOnce.Do(func() {
		close(q.ch)
	})
}
