package queue

import (
	"context"
	"time"

	"gitverse-notifier/pkg/events"
)

type EventGroupsManager interface {
	Add(context.Context, *events.Event) (string, bool)
	TryProcessGroup(context.Context, string) ([]*events.Event, bool)
	PopExpiredGroups(context.Context, time.Time) []*events.Event
	PopAllGroups(context.Context) []*events.Event
}
