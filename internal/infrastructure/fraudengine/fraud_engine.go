package fraudengine

import (
	"context"
	"encoding/json"
	"frauddetection/internal/domain"
	"frauddetection/internal/domain/ports"
	"log"
	"math"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type FraudEngine struct {
	repo     ports.TransactionRepository
	mlClient MLClient
	redis    *redis.Client
}

func NewFraudEngine(
	repo ports.TransactionRepository,
	mlClient MLClient,
	redis *redis.Client,
) *FraudEngine {
	return &FraudEngine{
		repo:     repo,
		mlClient: mlClient,
		redis:    redis,
	}
}

func (e *FraudEngine) Analyze(
	ctx context.Context,
	tx domain.Transaction,
) (*domain.FraudAlert, error) {
	var totalScore float64
	var reasons []string

	// LAYER 1: RULES (Static)
	// Watchlists or Sanctions Lists are usually checked here
	if tx.Currency != "XAF" && tx.Amount > 6000 {
		totalScore += 0.5
		reasons = append(reasons, "high_value_foreign_currency")
	}

	// LAYER 2: BEHAVIORAL (Velocity/History)
	lastHourTotal, _ := e.repo.GetSumSince(
		ctx,
		tx.UserID,
		time.Now().Add(-1*time.Hour),
	)

	if lastHourTotal > 1000 {
		totalScore += 0.3
		reasons = append(reasons, "velocity_exceeded")
	}

	//  LAYER 3: PROBABILISTIC (ML/AI)
	aiProbability := e.mlClient.GetInference(tx)
	if aiProbability > 0.4 {
		totalScore += (aiProbability * 0.4)
		reasons = append(reasons, "ml_anomaly_detected")
	}

	alert := domain.FraudAlert{
		RiskScore: math.Min(totalScore, 1.0),
		Reason:    strings.Join(reasons, ", "),
	}

	// Convert the alert to JSON
	payload, _ := json.Marshal(alert)

	err := e.redis.Publish(ctx, "fraud_events", payload).Err()
	if err != nil {
		log.Printf("failed to publish fraud alert: %v", err)
	}

	// TODO: Update audit logs with fraud analysis results

	return &alert, nil
}
