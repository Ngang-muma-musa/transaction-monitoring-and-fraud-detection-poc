package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"frauddetection/internal/domain"
)

var ErrPaymentNotFound = errors.New("transaction not found")

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(
	ctx context.Context,
	tx *domain.Transaction,
) error {
	query := `
        INSERT INTO transactions (id, user_id, amount, status, created_at)
        VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		tx.ID,
		tx.UserID,
		tx.Amount,
		tx.Status,
		tx.CreatedAt,
	)
	return err
}

func (r *TransactionRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Transaction, error) {
	query := `
		SELECT id, user_id, amount, status, created_at
		FROM transactions WHERE id = $1`

	var tx domain.Transaction
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.Amount,
		&tx.Status,
		&tx.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrPaymentNotFound
	}
	return &tx, err
}

func (r *TransactionRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status domain.TransactionStatus,
) error {
	query := `UPDATE transactions SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *TransactionRepository) GetSumSince(
	ctx context.Context,
	userID string,
	since time.Time,
) (float64, error) {
	query := `
        SELECT COALESCE(SUM(amount), 0.0) 
        FROM transactions 
        WHERE user_id = $1 
        AND created_at >= $2 
        AND status != 'DECLINED'`

	var sum float64
	err := r.db.QueryRowContext(ctx, query, userID, since).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("error calculating velocity sum: %w", err)
	}
	return sum, nil
}
