package embedding

import (
	"context"
	"log"
	"time"
)

// Worker uses NULL embeddings in the database as a durable pending queue.
// Signals only wake the worker; coalescing them cannot lose pending records.
type Worker struct {
	process  func(context.Context) error
	wake     chan struct{}
	interval time.Duration
}

func NewWorker(p *Processor) *Worker {
	return &Worker{process: p.ProcessPendingContext, wake: make(chan struct{}, 1), interval: time.Minute}
}

// ProcessDemand and ProcessArea satisfy the services' embedder interfaces
// without making the HTTP request wait for inference.
func (w *Worker) ProcessDemand(uint) error { return w.notify() }
func (w *Worker) ProcessArea(uint) error   { return w.notify() }

func (w *Worker) notify() error {
	select {
	case w.wake <- struct{}{}:
	default:
	}
	return nil
}

// Run processes sequentially to avoid overloading Ollama. It catches up after
// restarts and retries failures on the next pass, even without new registrations.
func (w *Worker) Run(ctx context.Context, logger *log.Logger) {
	for ctx.Err() == nil {
		if err := w.process(ctx); err != nil && ctx.Err() == nil && logger != nil {
			logger.Printf("embedding worker: %v", err)
		}
		timer := time.NewTimer(w.interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-w.wake:
			timer.Stop()
		case <-timer.C:
		}
	}
}
