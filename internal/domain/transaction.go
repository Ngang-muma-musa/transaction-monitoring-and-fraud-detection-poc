package domain

import "time"

type TransactionStatus string

const (
	StatusPending  TransactionStatus = "PENDING"
	StatusQueued   TransactionStatus = "QUEUED"
	StatusApproved TransactionStatus = "APPROVED"
	StatusDeclined TransactionStatus = "DECLINED"
	StatusFlagged  TransactionStatus = "FLAGGED"
)

type Transaction struct {
	ID        string            `json:"id"         db:"id"`
	UserID    string            `json:"user_id"    db:"user_id"`
	Amount    float64           `json:"amount"     db:"amount"`
	Currency  string            `json:"currency"   db:"currency"`
	Status    TransactionStatus `json:"status"     db:"status"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
}
