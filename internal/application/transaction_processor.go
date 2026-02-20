package application

import (
	"context"
	"frauddetection/internal/domain"
	"frauddetection/internal/domain/ports"
	"log"
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

	log.Printf("Fraud analysis result for transaction %s: score=%.2f reasons=%v",
		tx.ID, result.RiskScore, result.Reason)

	if result.RiskScore > 0.6 {
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
