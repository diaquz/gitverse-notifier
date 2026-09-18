package queue

import (
	"context"
	"fmt"

	"gitverse-notifier/pkg/events"
)

type EventQueue interface {
	Enqueue(ctx context.Context, event *events.Event) error
	Events() <-chan *events.Event
	Close()
}

type MemoryQueue struct {
	ch chan *events.Event
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
	}
}

func (q *MemoryQueue) Events() <-chan *events.Event {
	return q.ch
}

func (q *MemoryQueue) Close() {
	close(q.ch)
}
