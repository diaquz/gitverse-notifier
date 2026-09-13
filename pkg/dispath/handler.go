package dispath

import (
	"context"
	"gitverse-notifier/pkg/events"
)

type ActionHandler interface {
	Name() string
	Run(ctx context.Context, ev events.Event, rule *events.ActionRule) error
	Ready() bool
}
