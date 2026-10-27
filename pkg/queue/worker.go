package queue

import (
	"context"
	"stitchforce-server/pkg/logger"
	"sync"
)

type WorkerPool struct {
	queue   Queue
	handler Handler
	count   int
	logger  *logger.Logger
}

func NewWorkerPool(queue Queue, handler Handler, count int, logger *logger.Logger) *WorkerPool {
	return &WorkerPool{queue: queue, handler: handler, count: count, logger: logger}
}

func (p *WorkerPool) Start(ctx context.Context) {
	jobs, err := p.queue.Consume(ctx)
	if err != nil {
		p.logger.Error("queue consumer failed to start", "error", err)
		return
	}

	var wg sync.WaitGroup
	for i := 0; i < p.count; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}
					if err := p.handler.Handle(ctx, job); err != nil {
						p.logger.Error("background job failed",
							"worker_id", id,
							"job_type", job.Type(),
							"error", err)
					}
				}
			}
		}(i + 1)
	}

	go func() {
		wg.Wait()
		p.logger.Info("worker pool stopped")
	}()
}
