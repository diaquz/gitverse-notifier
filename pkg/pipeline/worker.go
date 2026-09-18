package pipeline

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
)

type EventProcessor interface {
	Process(ctx context.Context, event events.Event)
}

func (p *Pipeline) StartWorkers(ctx context.Context, workersCount int) {
	for i := 0; i < workersCount; i++ {
		go p.startWorker(ctx, i)
	}
}

func (p *Pipeline) startWorker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-p.queue.Events():
			if !ok {
				return
			}
			logger.Debug(ctx, "queue worker processing event",
				"worker", id,
				"event", event.Type,
				"repository", event.Repository)

			p.Run(ctx, event)
		}
	}
}
