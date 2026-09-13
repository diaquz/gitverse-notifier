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

func (d *Dispatcher) Dispatch(ctx context.Context, event events.Event) {
	if event.Type == events.Unknown {
		logger.Warnf("[EventDispath] skip unknown event for repository=%s", event.Repository)
		return
	}

	rules := d.manager.ActionsFor(event.Repository, event)
	if len(rules) == 0 {
		logger.Debugf("[EventDispath] no actions for event=%s repository=%s", event.Type, event.Repository)
		return
	}

	for _, rule := range rules {
		handler, ok := d.handlers[rule.Action]
		if !ok {
			logger.Errorf("no handler registered for action %q event=%s repository=%s", rule.Action, event.Type, event.Repository)
			continue
		}

		logger.Debugf("[EventDispather] processing %s action for %s", handler.Name(), event.Type)
		if err := handler.Run(ctx, event, &rule); err != nil {
			logger.Errorf("action %s failed for event=%s repository=%s: %v", rule.Action, event.Type, event.Repository, err)
		}
	}
}

func (d *Dispatcher) Register(handler ActionHandler) error {
	d.handlers[handler.Name()] = handler
	return nil
}
