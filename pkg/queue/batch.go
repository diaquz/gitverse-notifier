package queue

import (
	"context"
	"time"

	"gitverse-notifier/pkg/events"
)

type BatchesManager interface {
	Add(context.Context, *events.Event) (string, bool)
	TryProcessBucket(context.Context, string) ([]*events.Event, bool)
	PopExpired(context.Context, time.Time) []*events.Event
	PopAll(context.Context) []*events.Event
}
