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
			expired := p.batches.PopExpired(time.Now())
			for _, event := range expired {
				if err := p.queue.Enqueue(ctx, event); err != nil {
					logger.Error(ctx, "failed to enqueue expired batch event",
						"event", event.Type,
						"repository", event.Repository,
						"err", err)
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

			procCtx := ctx
			if event.RequestId != "" {
				procCtx = logger.With(ctx, "request-id", event.RequestId)
			}

			logger.Debug(procCtx, "queue worker processing event",
				"worker", id,
				"event", event.Type,
				"repository", event.Repository)

			if err := p.Run(procCtx, event); err != nil {
				logger.Error(procCtx, "pipeline run failed",
					"worker", id,
					"event", event.Type,
					"repository", event.Repository,
					"err", err)
			}
		}
	}
}
