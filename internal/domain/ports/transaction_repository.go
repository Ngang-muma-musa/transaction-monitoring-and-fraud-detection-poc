package ports

import (
	"context"
	"frauddetection/internal/domain"
	"time"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *domain.Transaction) error
	FindByID(ctx context.Context, id string) (*domain.Transaction, error)
	FindAll(ctx context.Context) ([]domain.Transaction, error)
	UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus) error
	GetSumSince(ctx context.Context, userID string, since time.Time) (float64, error)
}
