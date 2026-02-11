package worker

import (
	"context"
	"encoding/json"
	"frauddetection/internal/application"
	"frauddetection/internal/domain"
	"frauddetection/internal/domain/ports"
	"log"
	"time"
)

type Worker struct {
	ctx   context.Context
	queue ports.TransactionQueue
	app   *application.TransactionProcessor
}

func NewWorker(
	ctx context.Context,
	queue ports.TransactionQueue,
	app *application.TransactionProcessor,

) *Worker {
	return &Worker{
		ctx:   ctx,
		queue: queue,
		app:   app,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("Worker started...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down gracefully...")
			return

		default:
			job, err := w.queue.Reserve(ctx)
			if err != nil {
				log.Printf("Error reserving job: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			var tx domain.Transaction
			if err := json.Unmarshal(job.Payload, &tx); err != nil {
				log.Printf("failed to deserialize job: %v", err)
				w.queue.Bury(ctx, job.ID)
				continue
			}

			log.Printf("Processing payment ID: %s for user %s amount: %.2f %s",
				tx.ID, tx.UserID, tx.Amount, tx.Currency)
			w.app.Process(ctx, tx)

			w.queue.Delete(ctx, job.ID)
			log.Printf("Payment %s processed successfully", tx.ID)
		}
	}
}
