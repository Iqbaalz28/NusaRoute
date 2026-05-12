package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/services/order-service/internal/model"
	"github.com/nusaroute/services/order-service/internal/repository"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID string, req model.CreateOrderRequest) (*model.Order, error)
	GetOrder(ctx context.Context, id string) (*model.Order, error)
	ListOrders(ctx context.Context, userID string, page, perPage int) ([]model.Order, int64, error)
	CancelOrder(ctx context.Context, orderID, userID string) error
	HandlePaymentSuccess(ctx context.Context, orderID string) error
	HandleDeliveryFailed(ctx context.Context, orderID string, attempts int) error
	HandlePackageDelivered(ctx context.Context, orderID string) error
	RunSLAMonitor(ctx context.Context)
	RunPaymentExpiryChecker(ctx context.Context)
}

type orderService struct {
	repo     repository.OrderRepository
	producer *kafka.Producer
}

func NewOrderService(repo repository.OrderRepository, producer *kafka.Producer) OrderService {
	return &orderService{repo: repo, producer: producer}
}

func (s *orderService) publish(ctx context.Context, topic, key string, event interface{}) error {
	if s.producer == nil { return nil }
	return s.producer.Publish(ctx, topic, key, event)
}

func (s *orderService) CreateOrder(ctx context.Context, userID string, req model.CreateOrderRequest) (*model.Order, error) {
	if req.SenderName == "" || req.ReceiverName == "" {
		return nil, errors.New("sender and receiver info required")
	}

	order := &model.Order{
		UserID: userID, ServiceType: req.ServiceType,
		SenderName: req.SenderName, SenderPhone: req.SenderPhone,
		SenderAddress: req.SenderAddress, SenderLat: req.SenderLat, SenderLng: req.SenderLng,
		ReceiverName: req.ReceiverName, ReceiverPhone: req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress, ReceiverLat: req.ReceiverLat, ReceiverLng: req.ReceiverLng,
		ItemDescription: req.ItemDescription, WeightKg: req.WeightKg,
		ShippingCost: req.ShippingCost, InsuranceCost: req.InsuranceCost, TotalCost: req.TotalCost,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) ListOrders(ctx context.Context, userID string, page, perPage int) ([]model.Order, int64, error) {
	if page < 1 { page = 1 }
	if perPage < 1 { perPage = 10 }
	return s.repo.GetByUserID(ctx, userID, page, perPage)
}

func (s *orderService) CancelOrder(ctx context.Context, orderID, userID string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil { return errors.New("order not found") }
	if order.UserID != userID { return errors.New("not authorized") }
	
	if err := s.repo.MarkCancelled(ctx, orderID); err != nil { return err }

	event := events.OrderCancelledEvent{
		BaseEvent: events.BaseEvent{
			EventID: uuid.New().String(), EventType: events.TopicOrderCancelled,
			Timestamp: time.Now(), Source: "order-service",
		},
		OrderID: orderID, Reason: "Cancelled by user",
	}
	s.publish(ctx, events.TopicOrderCancelled, orderID, event)
	return nil
}

func (s *orderService) HandlePaymentSuccess(ctx context.Context, orderID string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil { return err }

	if err := s.repo.UpdateStatus(ctx, orderID, events.OrderStatusReadyForPickup, "Payment confirmed", "payment-service"); err != nil {
		return err
	}

	event := events.OrderReadyForPickupEvent{
		BaseEvent: events.BaseEvent{
			EventID: uuid.New().String(), EventType: events.TopicOrderReadyForPickup,
			Timestamp: time.Now(), Source: "order-service",
		},
		OrderID: orderID, AWB: order.AWB, SenderID: order.UserID,
		PickupLat: order.SenderLat, PickupLng: order.SenderLng, PickupAddr: order.SenderAddress,
		ReceiverAddr: order.ReceiverAddress, ReceiverLat: order.ReceiverLat, ReceiverLng: order.ReceiverLng,
		Weight: order.WeightKg, ServiceType: order.ServiceType,
	}
	return s.publish(ctx, events.TopicOrderReadyForPickup, orderID, event)
}

func (s *orderService) HandleDeliveryFailed(ctx context.Context, orderID string, attempts int) error {
	if attempts >= events.MaxDeliveryAttempts {
		s.repo.UpdateStatus(ctx, orderID, events.OrderStatusReturnToSender, "Max delivery attempts reached", "system")
	} else {
		s.repo.UpdateStatus(ctx, orderID, events.OrderStatusDeliveryFailed, fmt.Sprintf("Delivery attempt %d failed", attempts), "system")
	}
	return s.repo.IncrementDeliveryAttempts(ctx, orderID)
}

func (s *orderService) HandlePackageDelivered(ctx context.Context, orderID string) error {
	return s.repo.MarkDelivered(ctx, orderID)
}

func (s *orderService) RunSLAMonitor(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done(): return
		case <-ticker.C:
			stuckOrders, err := s.repo.GetStuckOrders(ctx, 48*time.Hour)
			if err != nil { continue }
			for _, o := range stuckOrders {
				s.repo.UpdateStatus(ctx, o.ID, events.OrderStatusLostSuspected, "No scan for 48+ hours", "sla-monitor")
			}
		}
	}
}

func (s *orderService) RunPaymentExpiryChecker(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done(): return
		case <-ticker.C:
			expired, err := s.repo.GetExpiredPendingOrders(ctx, 24*time.Hour)
			if err != nil { continue }
			for _, o := range expired {
				s.repo.MarkCancelled(ctx, o.ID)
			}
		}
	}
}
