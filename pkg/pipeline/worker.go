package pipeline

import (
	"context"
	"time"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/logger"
)

const batchExpireCheckInterval = time.Second

func (p *Pipeline) StartWorkers(ctx context.Context) {
	p.wg.Add(1)
	go p.startBatchesChecker(ctx)

	for i := 0; i < config.GlobalConfig.EventWorkerCount; i++ {
		logger.Debug(ctx, "starting queue worker", "worker", i)
		p.wg.Add(1)
		go p.startWorker(ctx, i)
	}
}

func (p *Pipeline) startBatchesChecker(ctx context.Context) {
	defer p.wg.Done()

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
	defer p.wg.Done()
	baseCtx := context.WithoutCancel(ctx)

	for event := range p.queue.Events() {
		start := time.Now()
		eventCtx := baseCtx
		if event.RequestId != "" {
			eventCtx = logger.With(baseCtx, "request-id", event.RequestId)
		}

		if err := p.Run(eventCtx, event); err != nil {
			logger.Error(eventCtx, "queue worker failed to process event",
				"time", time.Since(start).Milliseconds(),
				"worker", id, "event", event.Type, "repository", event.Repository, "err", err)
			continue
		}

		logger.Info(eventCtx, "queue worker processed event",
			"time", time.Since(start).Milliseconds(),
			"worker", id, "event", event.Type, "repository", event.Repository)
	}
}
