package dispath

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
)

type Processor struct {
	dispatcher *Dispatcher
	enrichers  []events.Enricher
}

func New(dispatcher *Dispatcher, enrichers ...events.Enricher) *Processor {
	return &Processor{
		dispatcher: dispatcher,
		enrichers:  enrichers,
	}
}

func (p *Processor) Process(ctx context.Context, event events.Event) {
	if !p.dispatcher.PotentialyDispathable(event) {
		logger.Info(ctx, "no actions for event, skipping",
			"event", event.Type, "repository", event.Repository)
		return
	}

	for _, enricher := range p.enrichers {
		if enricher.Skip(&event) {
			logger.Info(ctx, "event enriching skipped", "event", event.Type, "enricher", enricher.Name())
			continue
		}

		if err := enricher.Enrich(ctx, &event); err != nil {
			logger.Error(ctx, "event enriching failed",
				"action", "event_processing",
				"enricher", enricher.Name(),
				"event", event.Type,
				"repository", event.Repository,
				"err", err)
		}
	}

	p.dispatcher.Dispatch(ctx, event)
}
