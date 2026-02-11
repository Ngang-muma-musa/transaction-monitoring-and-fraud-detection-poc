package ports

import (
	"context"
	"frauddetection/internal/domain"
)

type FraudDetector interface {
	Analyze(
		ctx context.Context,
		tx domain.Transaction,
	) (*domain.FraudAlert, error)
}
