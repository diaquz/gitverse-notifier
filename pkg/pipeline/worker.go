package pipeline

import (
	"context"
	"time"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/logger"
)

const batchExpireCheckInterval = time.Second

func (p *Pipeline) StartWorkers(ctx context.Context) {
	go p.startBatchesChecker(ctx)

	for i := 0; i < config.GlobalConfig.EventWorkerCount; i++ {
		logger.Debug(ctx, "starting queue worker", "worker", i)
		go p.startWorker(ctx, i)
	}
}

func (p *Pipeline) startBatchesChecker(ctx context.Context) {
	ticker := time.NewTicker(batchExpireCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expired := p.batches.PopExpired(ctx, time.Now())

			for _, event := range expired {
				logger.Debug(ctx, "enqueue expired event", "event", event.Type, "repository", event.Repository)
				if err := p.queue.Enqueue(ctx, event); err != nil {
					logger.Error(ctx, "failed to enqueue expired batch event",
						"event", event.Type, "repository", event.Repository, "err", err)
				}
			}
		}
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

			// Все логи будут содержать request-id, полученный из gitverse delivery заголовка
			eventCtx := ctx
			if event.RequestId != "" {
				eventCtx = logger.With(ctx, "request-id", event.RequestId)
			}

			logger.Debug(eventCtx, "queue worker processing event",
				"worker", id, "event", event.Type, "repository", event.Repository)

			if err := p.Run(eventCtx, event); err != nil {
				logger.Error(eventCtx, "pipeline run failed",
					"worker", id, "event", event.Type, "repository", event.Repository, "err", err)
			}
		}
	}
}
