package dispath

import (
	"context"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/settings"
)

type ActionHandler interface {
	Name() string
	Run(ctx context.Context, ev events.Event, rule *settings.ActionRule) error
	Ready() bool
}
