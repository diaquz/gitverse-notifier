package pipeline

import (
	"context"
	"sync"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/handlers"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/queue"
	"gitverse-notifier/pkg/queue/groups"
	"gitverse-notifier/pkg/settings"
)

type Pipeline struct {
	manager   *settings.SettingsManager
	enrichers []events.Enricher
	handlers  map[string]handlers.ActionHandler
	queue     queue.EventQueue
	groups    queue.EventGroupsManager

	wg        sync.WaitGroup
	closeOnce sync.Once
}

func BuildNewPipeline(
	manager *settings.SettingsManager,
	enrichers []events.Enricher,
	actionHandlers ...handlers.ActionHandler,
) *Pipeline {
	cfg := config.GlobalConfig

	pipeline := &Pipeline{
		queue:     queue.NewMemoryQueue(cfg.EventQueueSize),
		groups:    groups.NewInMemoryEventGroupsManager(manager),
		enrichers: enrichers,
		manager:   manager,
		handlers:  make(map[string]handlers.ActionHandler, len(actionHandlers)),
	}

	for i := range actionHandlers {
		pipeline.handlers[actionHandlers[i].Name()] = actionHandlers[i]
	}

	return pipeline
}

// Enqueue ставит событие в очередь, если событие нужно группировать - пытается добавить в группу
func (p *Pipeline) Enqueue(ctx context.Context, event *events.Event) error {
	logger.Info(ctx, "enqueuing event", "action", "pipeline.enqueue", "event", event.Type, "repository", event.Repository)

	key, ok := p.groups.Add(ctx, event)
	if !ok {
		return p.queue.Enqueue(ctx, event)
	}

	if events, ok := p.groups.TryProcessGroup(ctx, key); ok {
		for _, event := range events {
			if err := p.queue.Enqueue(ctx, event); err != nil {
				logger.Error(ctx, "failed to enqueue group event",
					"action", "pipeline.enqueue", "err", err,
					"event", event.Type, "repository", event.Repository)
			}
		}
	}

	return nil
}

func (p *Pipeline) Run(ctx context.Context, event *events.Event) error {
	// Пропускаем события, для которых гарантированно нет действий, чтобы лишний раз не делать запросы к gitverse API
	if !p.manager.HasPotentialActions(event.Repository, event) {
		logger.Info(ctx, "no potential actions, skipping",
			"action", "pipeline.run", "event", event.Type, "repository", event.Repository)
		return nil
	}

	p.Enrich(ctx, event)
	return p.Dispatch(ctx, event)
}

func (p *Pipeline) Enrich(ctx context.Context, event *events.Event) {
	for _, enricher := range p.enrichers {
		if enricher.Skip(event) {
			logger.Info(ctx, "event enriching skipped",
				"action", "pipeline.run", "event", event.Type, "enricher", enricher.Name())
			continue
		}

		if err := enricher.Enrich(ctx, event); err != nil {
			logger.Error(ctx, "event enriching failed",
				"action", "pipeline.run", "err", err,
				"enricher", enricher.Name(), "event", event.Type, "repository", event.Repository)
		}
	}
}

func (p *Pipeline) Dispatch(ctx context.Context, event *events.Event) error {
	rules := p.manager.ActionsFor(event.Repository, event)
	if len(rules) == 0 {
		logger.Info(ctx, "no actions for event", "action", "pipeline.run", "event", event.Type, "repository", event.Repository)
		return nil
	}

	for _, rule := range rules {
		handler, ok := p.handlers[rule.Action]
		if !ok {
			logger.Error(ctx, "no handlers registered for action",
				"action", "pipeline.run", "handler", rule.Action, "event", event.Type, "repository", event.Repository)
			continue
		}

		logger.Info(ctx, "processing action for event",
			"action", "pipeline.run", "handler", handler.Name(), "event", event.Type)
		if !handler.Ready() {
			logger.Warn(ctx, "handler is not configured, skipping", "action", "pipeline.run", "handler", rule.Action)
			continue
		}

		if err := handler.Run(ctx, event, &rule); err != nil {
			logger.Error(ctx, "action failed",
				"action", "pipeline.run", "handler", rule.Action,
				"event", event.Type, "repository", event.Repository, "err", err)
		}
	}

	return nil
}

func (p *Pipeline) Close(ctx context.Context) error {
	closeErr := p.flushGroups(ctx)
	workersErr := p.waitWorkers(ctx)

	if closeErr != nil {
		return closeErr
	}

	return workersErr
}

func (p *Pipeline) flushGroups(ctx context.Context) (closeErr error) {
	p.closeOnce.Do(func() {
		for _, event := range p.groups.PopAllGroups(ctx) {
			logger.Debug(ctx, "enqueue flushed group event",
				"action", "pipeline.shutdown", "event", event.Type, "repository", event.Repository)

			if err := p.queue.EnqueueBlocking(ctx, event); err != nil {
				logger.Error(ctx, "failed to enqueue flushed group event",
					"action", "pipeline.shutdown", "event", event.Type, "repository", event.Repository, "err", err)
				closeErr = err
			}
		}

		p.queue.Close()
	})

	return
}
