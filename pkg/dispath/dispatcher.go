package dispath

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/repositories"
)

type Dispatcher struct {
	manager  *repositories.RepositoriesManager
	handlers map[string]ActionHandler
}

func NewDispatcher(manager *repositories.RepositoriesManager, handlers ...ActionHandler) *Dispatcher {
	d := &Dispatcher{
		manager:  manager,
		handlers: make(map[string]ActionHandler, len(handlers)),
	}
	for _, h := range handlers {
		d.handlers[h.Name()] = h
	}
	return d
}

func (d *Dispatcher) PotentialyDispathable(event events.Event) bool {
	return d.manager.HasPotentialActions(event.Repository, event)
}

func (d *Dispatcher) Dispatch(ctx context.Context, event events.Event) {
	rules := d.manager.ActionsFor(event.Repository, event)
	if len(rules) == 0 {
		logger.Info(ctx, "no actions for event", "event", event.Type, "repository", event.Repository)
		return
	}

	for _, rule := range rules {
		handler, ok := d.handlers[rule.Action]
		if !ok {
			logger.Error(ctx, "no handlers registered for action",
				"action", rule.Action, "event", event.Type, "repository", event.Repository)
			continue
		}

		logger.Info(ctx, "processing action for event", "action", handler.Name(), "event", event.Type)
		if !handler.Ready() {
			logger.Warn(ctx, "handler is not configured, skipping", "action", rule.Action)
			continue
		}

		if err := handler.Run(ctx, event, &rule); err != nil {
			logger.Error(ctx, "action failed",
				"action", rule.Action, "event", event.Type, "repository", event.Repository, "err", err)
		}
	}
}

func (d *Dispatcher) Register(handler ActionHandler) error {
	d.handlers[handler.Name()] = handler
	return nil
}
