package domain

type FraudAlert struct {
	TransactionID string  `json:"transaction_id"`
	RiskScore     float64 `json:"risk_score"`
	Reason        string  `json:"reason"`
}
