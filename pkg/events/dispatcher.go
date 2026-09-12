package events

import (
	"context"
	"gitverse-notifier/pkg/logger"
)

type Dispatcher struct {
	manager  *ActionsManager
	handlers map[string]ActionHandler
}

func NewDispatcher(manager *ActionsManager, handlers ...ActionHandler) *Dispatcher {
	d := &Dispatcher{
		manager:  manager,
		handlers: make(map[string]ActionHandler, len(handlers)),
	}
	for _, h := range handlers {
		d.handlers[h.Name()] = h
	}
	return d
}

func (d *Dispatcher) Dispatch(ctx context.Context, event Event) {
	if event.Type == Unknown {
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

		if err := handler.Run(ctx, event, rule); err != nil {
			logger.Errorf("action %s failed for event=%s repository=%s: %v", rule.Action, event.Type, event.Repository, err)
		}
	}
}

func (d *Dispatcher) Register(handler ActionHandler) error {
	d.handlers[handler.Name()] = handler
	return nil
}
