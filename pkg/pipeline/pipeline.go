package pipeline

import (
	"context"
	"sync"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/dispatch"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/queue"
	"gitverse-notifier/pkg/queue/groups"
	"gitverse-notifier/pkg/settings"
)

type Pipeline struct {
	manager    *settings.SettingsManager
	enrichers  []events.Enricher
	dispatcher dispatch.EventDispatcher
	queue      queue.EventQueue
	groups    queue.EventGroupsManager

	wg        sync.WaitGroup
	closeOnce sync.Once
}

func BuildNewPipeline(
	manager *settings.SettingsManager,
	enrichers []events.Enricher,
	dispatcher dispatch.EventDispatcher,
) *Pipeline {
	cfg := config.GlobalConfig

	return &Pipeline{
		queue:      queue.NewMemoryQueue(cfg.EventQueueSize),
		groups:    groups.NewInMemoryEventGroupsManager(manager),
		enrichers:  enrichers,
		manager:    manager,
		dispatcher: dispatcher,
	}
}

// Enqueue ставит событие в очередь, если событие нужно группировать - пытается добавить в группу
func (p *Pipeline) Enqueue(ctx context.Context, event *events.Event) error {
	key, ok := p.groups.Add(ctx, event)
	if !ok {
		return p.queue.Enqueue(ctx, event)
	}

	if events, ok := p.groups.TryProcessGroup(ctx, key); ok {
		for _, event := range events {
			if err := p.queue.Enqueue(ctx, event); err != nil {
				logger.Error(ctx, "failed to enqueue group event",
					"event", event.Type, "repository", event.Repository,
					"err", err)
			}
		}
	}

	return nil
}

func (p *Pipeline) Run(ctx context.Context, event *events.Event) error {
	// Пропускаем события, для которых гарантированно нет действий, чтобы лишний раз не делать запросы к gitverse API
	if !p.dispatcher.Dispathable(event) {
		logger.Info(ctx, "no potential actions, skipping",
			"event", event.Type, "repository", event.Repository, "action", "pipeline_run")
		return nil
	}

	p.Enrich(ctx, event)
	p.dispatcher.Dispatch(ctx, event)

	return nil
}

func (p *Pipeline) Enrich(ctx context.Context, event *events.Event) {
	for _, enricher := range p.enrichers {
		if enricher.Skip(event) {
			logger.Info(ctx, "event enriching skipped",
				"action", "pipeline_run", "event", event.Type, "enricher", enricher.Name())
			continue
		}

		if err := enricher.Enrich(ctx, event); err != nil {
			logger.Error(ctx, "event enriching failed",
				"action", "pipeline_run",
				"enricher", enricher.Name(),
				"event", event.Type,
				"repository", event.Repository,
				"err", err)
		}
	}
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
				"event", event.Type, "repository", event.Repository)

			if err := p.queue.EnqueueBlocking(ctx, event); err != nil {
				logger.Error(ctx, "failed to enqueue flushed group event",
					"event", event.Type, "repository", event.Repository, "err", err)
				closeErr = err
			}
		}

		p.queue.Close()
	})

	return
}
