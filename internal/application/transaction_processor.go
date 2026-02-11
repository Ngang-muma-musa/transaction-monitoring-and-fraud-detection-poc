package application

import (
	"context"
	"frauddetection/internal/domain"
	"frauddetection/internal/domain/ports"
	"time"
)

type TransactionProcessor struct {
	repo  ports.TransactionRepository
	fraud ports.FraudDetector
}

func NewTransactionProcessor(
	repo ports.TransactionRepository,
	fraud ports.FraudDetector,
) *TransactionProcessor {
	return &TransactionProcessor{repo: repo, fraud: fraud}
}

func (p *TransactionProcessor) Process(
	ctx context.Context,
	tx domain.Transaction,
) error {
	result, err := p.fraud.Analyze(ctx, tx)
	if err != nil {
		return err
	}

	if result.RiskScore > 0.8 {
		return p.repo.UpdateStatus(
			ctx,
			tx.ID,
			domain.StatusFlagged,
		)
	}

	// Simulate payment processing
	time.Sleep(3 * time.Second)

	return p.repo.UpdateStatus(
		ctx,
		tx.ID,
		domain.StatusApproved,
	)
}
