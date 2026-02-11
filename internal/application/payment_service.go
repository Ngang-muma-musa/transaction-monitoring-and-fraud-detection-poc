package application

import (
	"context"
	"encoding/json"
	"errors"
	"frauddetection/internal/domain"
	"frauddetection/internal/domain/ports"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

type PaymentServiceApp interface {
	CreateAndQueuePayment(
		ctx context.Context,
		userID string,
		amount float64,
		currency string,
	) (*domain.Transaction, error)

	GetPaymentByID(
		ctx context.Context,
		paymentID string,
	) (*domain.Transaction, error)
}

type PaymentService struct {
	repo        ports.TransactionRepository
	queue       ports.TransactionQueue
	rateLimiter ports.RateLimiterPort
}

func NewPaymentService(
	repo ports.TransactionRepository,
	queue ports.TransactionQueue,
	rateLimiter ports.RateLimiterPort,
) *PaymentService {
	return &PaymentService{
		repo:        repo,
		queue:       queue,
		rateLimiter: rateLimiter,
	}
}

// CreateAndQueuePayment handles:
// - rate limit check
// - create payment record
// - enqueue payment job
func (s *PaymentService) CreateAndQueuePayment(
	ctx context.Context,
	userID string,
	amount float64,
	currency string,
) (*domain.Transaction, error) {
	allowed, err := s.rateLimiter.Allow(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrRateLimitExceeded
	}

	payment := &domain.Transaction{
		ID:        gofakeit.UUID(),
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	p, err := json.Marshal(payment)
	if err != nil {
		return nil, err
	}

	job := domain.Job{
		ID:      payment.ID,
		Payload: p,
	}

	if err := s.queue.Publish(ctx, job); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateStatus(ctx, payment.ID, "QUEUED"); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *PaymentService) GetPaymentByID(
	ctx context.Context,
	paymentID string,
) (*domain.Transaction, error) {
	return s.repo.FindByID(ctx, paymentID)
}
