package fraudengine

import (
	"frauddetection/internal/domain"
	"math/rand"
)

// MLClient simulates a connection to a model server (like SageMaker or Vertex AI)
type MLClient struct{}

func NewMLClient() MLClient {
	return MLClient{}
}

func (m *MLClient) GetInference(tx domain.Transaction) float64 {

	var risk float64

	// High-risk hours
	hour := tx.CreatedAt.Hour()
	if hour >= 2 && hour <= 5 {
		risk += 0.25
	}

	// Odd amounts
	if tx.Amount == 100.00 || tx.Amount == 500.00 {
		risk += 0.15
	}

	// model noise
	noise := rand.Float64() * 0.2

	return risk + noise
}
