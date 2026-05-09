package service

import (
	"context"
	"errors"

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

func (s *orderService) CreateOrder(ctx context.Context, userID string, req model.CreateOrderRequest) (*model.Order, error) {
	// TODO: Implement create order (Tubes Tahap Dua)
	return nil, errors.New("method CreateOrder not implemented")
}

func (s *orderService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	// TODO: Implement get order (Tubes Tahap Dua)
	return nil, errors.New("method GetOrder not implemented")
}

func (s *orderService) ListOrders(ctx context.Context, userID string, page, perPage int) ([]model.Order, int64, error) {
	return nil, 0, errors.New("method ListOrders not implemented")
}

func (s *orderService) CancelOrder(ctx context.Context, orderID, userID string) error {
	// TODO: Implement cancel order (Tubes Tahap Dua)
	return errors.New("method CancelOrder not implemented")
}

func (s *orderService) HandlePaymentSuccess(ctx context.Context, orderID string) error {
	return errors.New("method HandlePaymentSuccess not implemented")
}

func (s *orderService) HandleDeliveryFailed(ctx context.Context, orderID string, attempts int) error {
	return errors.New("method HandleDeliveryFailed not implemented")
}

func (s *orderService) HandlePackageDelivered(ctx context.Context, orderID string) error {
	return errors.New("method HandlePackageDelivered not implemented")
}

func (s *orderService) RunSLAMonitor(ctx context.Context) {
}

func (s *orderService) RunPaymentExpiryChecker(ctx context.Context) {
}
