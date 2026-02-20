package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"frauddetection/internal/domain"
	"frauddetection/internal/domain/ports"
	"log"
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

	GetAllPayments(
		ctx context.Context,
	) ([]domain.Transaction, error)
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
	log.Printf(
		"Creating payment for user %s amount: %.2f %s",
		userID, amount, currency,
	)
	allowed, err := s.rateLimiter.Allow(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		log.Printf("Rate limit exceeded for user %s", userID)
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
		log.Printf("Error creating payment record: %v", err)
		return nil, fmt.Errorf("Error inserting into db %s", err.Error())
	}

	p, err := json.Marshal(payment)
	if err != nil {
		log.Printf("Error serializing payment for queue: %v", err)
		return nil, err
	}

	job := domain.Job{
		ID:      payment.ID,
		Payload: p,
	}

	if err := s.queue.Publish(ctx, job); err != nil {
		log.Printf("Error publishing payment to queue: %v", err)
		return nil, fmt.Errorf("Error publishing to queue %s", err.Error())
	}

	if err := s.repo.UpdateStatus(ctx, payment.ID, domain.StatusQueued); err != nil {
		log.Printf("Error updating payment status to QUEUED: %v", err)
		return nil, fmt.Errorf("Error updating payment status to QUEUED: %s", err.Error())
	}

	return payment, nil
}

func (s *PaymentService) GetPaymentByID(
	ctx context.Context,
	paymentID string,
) (*domain.Transaction, error) {
	log.Printf("Fetching payment with ID: %s", paymentID)
	return s.repo.FindByID(ctx, paymentID)
}

func (s *PaymentService) GetAllPayments(
	ctx context.Context,
) ([]domain.Transaction, error) {
	log.Printf("Fetching all payments")
	return s.repo.FindAll(ctx)
}
