package dispatch

import (
	"context"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/settings"
)

type ActionHandler interface {
	Name() string
	Run(context.Context, *events.Event, *settings.ActionRule) error
	Ready() bool
}
