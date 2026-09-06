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

func (d *Dispatcher) Dispatch(ctx context.Context, ev Event) {
	if ev.Type == Unknown {
		logger.Warnf("skip unknown gitverse event repo=%s", ev.Repository)
		return
	}

	rules := d.manager.ActionsFor(ev.Repository, ev)
	if len(rules) == 0 {
		logger.Debugf("no actions for event=%s repo=%s", ev.Type, ev.Repository)
		return
	}

	for _, rule := range rules {
		handler, ok := d.handlers[rule.Action]
		if !ok {
			logger.Errorf("no handler registered for action %q (event=%s repo=%s)", rule.Action, ev.Type, ev.Repository)
			continue
		}
		if err := handler.Run(ctx, ev, rule); err != nil {
			logger.Errorf("action %s failed for event=%s repo=%s: %v", rule.Action, ev.Type, ev.Repository, err)
		}
	}
}

func (d *Dispatcher) Register(handler ActionHandler) error {
	if d.handlers == nil {
		d.handlers = make(map[string]ActionHandler)
	}
	d.handlers[handler.Name()] = handler
	return nil
}
