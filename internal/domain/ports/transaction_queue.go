package ports

import (
	"context"
	"frauddetection/internal/domain"
)

type TransactionQueue interface {
	Publish(ctx context.Context, tx domain.Job) error
	Reserve(ctx context.Context) (*domain.Job, error)
	// Bury puts a job into a "buried" state where it will not be processed
	// until it is manually "kicked" back into the ready queue.
	Bury(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}
