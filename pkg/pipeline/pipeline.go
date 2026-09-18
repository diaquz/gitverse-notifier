package pipeline

import (
	"context"
	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/dispath"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/queue"
	"gitverse-notifier/pkg/settings"
)

type Pipeline struct {
	manager   *settings.SettingsManager
	enrichers []events.Enricher
	dispather *dispath.Dispatcher
	queue     queue.EventQueue
	batches   queue.BatchesManager
}

func BuildNewPipeline(
	manager *settings.SettingsManager,
	enrichers []events.Enricher,
	dispather *dispath.Dispatcher,
) *Pipeline {
	cfg := config.GlobalConfig

	return &Pipeline{
		queue:     queue.NewMemoryQueue(cfg.EventQueueSize),
		batches:   queue.NewInMemoryBatchesManager(manager),
		enrichers: enrichers,
		manager:   manager,
		dispather: dispather,
	}
}

// Enqueue ставит событие в очередь, если событие нужно групировать - пытается добавить в группу
func (p *Pipeline) Enqueue(ctx context.Context, event *events.Event) error {
	key, ok := p.batches.Add(event)
	if !ok {
		return p.queue.Enqueue(ctx, event)
	}

	if events, ok := p.batches.TryProcessBucket(key); ok {
		for _, event := range events {
			p.queue.Enqueue(ctx, event)
			// TODO: Log error
		}
	}

	return nil
}

func (p *Pipeline) Run(ctx context.Context, event *events.Event) error {
	setting := p.manager.SettingsByRepository(event.Repository)
	// Пропускаем события, для которых гарантированно нет событий, чтобы лишни раз не делать запросы к gitverse API
	if !setting.HasPotentialActions(*event) {
		logger.Info(ctx, "no potential actions, skipping",
			"event", event.Type, "repository", event.Repository, "action", "pipeline_run")
		return nil
	}

	p.Enrich(ctx, event)
	p.dispather.Dispatch(ctx, *event)

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
