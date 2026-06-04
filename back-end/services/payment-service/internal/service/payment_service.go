package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nusaroute/pkg/logger"

	"github.com/google/uuid"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/services/payment-service/internal/model"
	"github.com/nusaroute/services/payment-service/internal/repository"
)

// PaymentService defines the interface for payment business logic.
type PaymentService interface {
	InitiatePayment(ctx context.Context, req model.InitiatePaymentRequest) (*model.Transaction, error)
	HandleWebhook(ctx context.Context, payload model.WebhookPayload) error
	GetPaymentStatus(ctx context.Context, orderID string) (*model.Transaction, error)
}

type paymentService struct {
	repo     repository.PaymentRepository
	producer *kafka.Producer
}

func NewPaymentService(repo repository.PaymentRepository, producer *kafka.Producer) PaymentService {
	return &paymentService{repo: repo, producer: producer}
}

func (s *paymentService) InitiatePayment(ctx context.Context, req model.InitiatePaymentRequest) (*model.Transaction, error) {
	if req.OrderID == "" || req.Amount <= 0 {
		return nil, errors.New("order_id and positive amount are required")
	}

	idempotencyKey := fmt.Sprintf("pay_%s_%s", req.OrderID, req.Method)

	// Check for existing transaction with same idempotency key (prevent duplicate)
	existing, _ := s.repo.GetByIdempotencyKey(ctx, idempotencyKey)
	if existing != nil {
		logger.Info(context.Background(), fmt.Sprintf("[Payment] Idempotent request detected for order %s, returning existing transaction", req.OrderID))
		return existing, nil
	}

	// Simulate payment gateway integration
	tx := &model.Transaction{
		OrderID:        req.OrderID,
		Amount:         req.Amount,
		Method:         req.Method,
		PaymentURL:     fmt.Sprintf("https://pay.nusaroute.id/checkout/%s", uuid.New().String()),
		ExternalID:     uuid.New().String(),
		IdempotencyKey: idempotencyKey,
	}

	if err := s.repo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return tx, nil
}

func (s *paymentService) HandleWebhook(ctx context.Context, payload model.WebhookPayload) error {
	// Idempotency check: if already processed, skip
	existing, _ := s.repo.GetByIdempotencyKey(ctx, payload.IdempotencyKey)
	if existing != nil && existing.Status != model.PaymentStatusPending {
		logger.Info(context.Background(), fmt.Sprintf("[Payment] Webhook already processed for key=%s, skipping", payload.IdempotencyKey))
		return nil // Idempotent - already processed
	}

	// Find the transaction
	tx, err := s.repo.GetByOrderID(ctx, payload.OrderID)
	if err != nil {
		return fmt.Errorf("transaction not found for order %s: %w", payload.OrderID, err)
	}

	switch payload.Status {
	case "PAID":
		event := events.PaymentSuccessEvent{
			BaseEvent: events.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: events.TopicPaymentSuccess,
				Timestamp: time.Now(),
				Source:    "payment-service", TraceID: logger.GetTraceID(ctx)},
			OrderID:       tx.OrderID,
			TransactionID: tx.ID,
			Amount:        tx.Amount,
			Method:        tx.Method,
		}

		if err := s.repo.MarkPaid(ctx, tx.ID, events.TopicPaymentSuccess, event); err != nil {
			return fmt.Errorf("failed to mark as paid and save outbox: %w", err)
		}

		logger.Info(context.Background(), fmt.Sprintf("[Payment]  Payment confirmed for order %s (amount: %.2f) - Event saved to outbox", tx.OrderID, tx.Amount))

	case "FAILED":
		event := events.PaymentFailedEvent{
			BaseEvent: events.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: events.TopicPaymentFailed,
				Timestamp: time.Now(),
				Source:    "payment-service", TraceID: logger.GetTraceID(ctx)},
			OrderID: tx.OrderID,
			Reason:  "Payment declined by gateway",
		}

		if err := s.repo.MarkFailed(ctx, tx.ID, events.TopicPaymentFailed, event); err != nil {
			return fmt.Errorf("failed to mark as failed and save outbox: %w", err)
		}
		logger.Info(context.Background(), fmt.Sprintf("[Payment] ❌ Payment failed for order %s - Event saved to outbox", tx.OrderID))

	default:
		return fmt.Errorf("unknown payment status: %s", payload.Status)
	}

	return nil
}

func (s *paymentService) GetPaymentStatus(ctx context.Context, orderID string) (*model.Transaction, error) {
	tx, err := s.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, errors.New("payment not found")
	}
	return tx, nil
}
