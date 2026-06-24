package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/logger"
	"github.com/nusaroute/pkg/routing"
	"github.com/nusaroute/services/order-service/internal/hubs"
	"github.com/nusaroute/services/order-service/internal/model"
	"github.com/nusaroute/services/order-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID string, req model.CreateOrderRequest) (*model.Order, error)
	GetOrder(ctx context.Context, id string) (*model.Order, error)
	GetOrderByAWB(ctx context.Context, awb string) (*model.Order, error)
	ListOrders(ctx context.Context, userID string, page, perPage int) ([]model.Order, int64, error)
	CancelOrder(ctx context.Context, orderID, userID string) error
	ListAllOrders(ctx context.Context, page, perPage int, search string) ([]model.Order, int64, error)
	AdminUpdateStatus(ctx context.Context, orderID, status string) error
	HandlePaymentSuccess(ctx context.Context, orderID string) error
	HandleDeliveryFailed(ctx context.Context, orderID string, attempts int) error
	HandlePackageDelivered(ctx context.Context, orderID string) error
	HandlePackagePickedUp(ctx context.Context, awb string) error
	HandleHubScan(ctx context.Context, awb, hubID, hubName, scanType string) error
	RunSLAMonitor(ctx context.Context)
	RunPaymentExpiryChecker(ctx context.Context)
	GetDashboardStats(ctx context.Context) (int64, float64, error)
	GetVolumeStats(ctx context.Context) ([]model.DailyVolume, error)
}

type orderService struct {
	repo     repository.OrderRepository
	producer *kafka.Producer
	redis    *redis.Client
	hubs     *hubs.Resolver
}

func NewOrderService(repo repository.OrderRepository, producer *kafka.Producer, r *redis.Client, hubResolver *hubs.Resolver) OrderService {
	return &orderService{repo: repo, producer: producer, redis: r, hubs: hubResolver}
}

// publish safely publishes a Kafka event, skipping if producer is nil (e.g. in tests).
func (s *orderService) publish(ctx context.Context, topic, key string, event interface{}) error {
	if s.producer == nil {
		log.Printf("[Order] Kafka producer is nil, skipping publish to %s", topic)
		return nil
	}
	return s.producer.Publish(ctx, topic, key, event)
}

func (s *orderService) CreateOrder(ctx context.Context, userID string, req model.CreateOrderRequest) (*model.Order, error) {
	if req.SenderName == "" || req.ReceiverName == "" {
		return nil, errors.New("sender and receiver info required")
	}

	// Decide routing: same-city + instant service is delivered directly by the
	// pickup courier; everything else is routed through a sortation hub.
	deliveryMode := routing.DecideRouteByCoords(req.ServiceType, req.SenderLat, req.SenderLng, req.ReceiverLat, req.ReceiverLng)

	// How the package reaches the origin hub: picked up by a courier (default) or
	// dropped off at the hub by the sender.
	pickupMode := req.PickupMode
	if pickupMode != "SELF_DROPOFF" {
		pickupMode = "COURIER"
	}

	order := &model.Order{
		UserID: userID, ServiceType: req.ServiceType, DeliveryMode: deliveryMode, PickupMode: pickupMode,
		SenderName: req.SenderName, SenderPhone: req.SenderPhone, SenderCity: req.SenderCity,
		SenderAddress: req.SenderAddress, SenderLat: req.SenderLat, SenderLng: req.SenderLng,
		ReceiverName: req.ReceiverName, ReceiverPhone: req.ReceiverPhone, ReceiverCity: req.ReceiverCity,
		ReceiverAddress: req.ReceiverAddress, ReceiverLat: req.ReceiverLat, ReceiverLng: req.ReceiverLng,
		ItemDescription: req.ItemDescription, WeightKg: req.WeightKg,
		LengthCm: req.LengthCm, WidthCm: req.WidthCm, HeightCm: req.HeightCm,
		IsInsured: req.IsInsured, InsuredValue: req.InsuredValue,
		ShippingCost: req.ShippingCost, InsuranceCost: req.InsuranceCost, TotalCost: req.TotalCost,
	}

	// Stamp the nearest sortation hubs to sender (origin) and receiver (dest).
	// Non-fatal: if the hub list is unavailable, the order is still created.
	if s.hubs != nil {
		if h, ok := s.hubs.Nearest(ctx, req.SenderLat, req.SenderLng); ok {
			order.OriginHubCode, order.OriginHubName = h.Code, h.Name
		}
		if h, ok := s.hubs.Nearest(ctx, req.ReceiverLat, req.ReceiverLng); ok {
			order.DestHubID, order.DestHubCode, order.DestHubName = h.ID, h.Code, h.Name
		}
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	log.Printf("[Order] Created order AWB=%s for user=%s", order.AWB, userID)
	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) GetOrderByAWB(ctx context.Context, awb string) (*model.Order, error) {
	return s.repo.GetByAWB(ctx, awb)
}

func (s *orderService) ListOrders(ctx context.Context, userID string, page, perPage int) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	return s.repo.GetByUserID(ctx, userID, page, perPage)
}

func (s *orderService) CancelOrder(ctx context.Context, orderID, userID string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return errors.New("order not found")
	}
	if order.UserID != userID {
		return errors.New("not authorized")
	}
	if order.Status != events.OrderStatusPendingPayment && order.Status != events.OrderStatusReadyForPickup {
		return errors.New("order cannot be cancelled in current status")
	}

	event := events.OrderCancelledEvent{
		BaseEvent: events.BaseEvent{
			EventID: uuid.New().String(), EventType: events.TopicOrderCancelled,
			Timestamp: time.Now(), Source: "order-service", TraceID: logger.GetTraceID(ctx)},
		OrderID: orderID, Reason: "Cancelled by user",
	}

	if err := s.repo.MarkCancelled(ctx, orderID, events.TopicOrderCancelled, event); err != nil {
		return err
	}

	return nil
}

// ListAllOrders returns all orders (admin view) with optional search.
func (s *orderService) ListAllOrders(ctx context.Context, page, perPage int, search string) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	return s.repo.ListAll(ctx, page, perPage, search)
}

// AdminUpdateStatus lets an admin override an order's status directly.
func (s *orderService) AdminUpdateStatus(ctx context.Context, orderID, status string) error {
	valid := map[string]bool{
		events.OrderStatusPendingPayment: true, events.OrderStatusReadyForPickup: true,
		events.OrderStatusPickedUp: true, events.OrderStatusInTransit: true,
		events.OrderStatusOutForDelivery: true, events.OrderStatusDelivered: true,
		events.OrderStatusDeliveryFailed: true, events.OrderStatusReturnToSender: true,
		events.OrderStatusCancelled: true, events.OrderStatusLostSuspected: true,
	}
	if !valid[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Setting an order READY_FOR_PICKUP must emit the same event the payment flow
	// does, otherwise dispatch-service never hears about it and no courier is
	// auto-assigned. Mirror HandlePaymentSuccess so the admin path is complete.
	if status == events.OrderStatusReadyForPickup {
		order, err := s.repo.GetByID(ctx, orderID)
		if err != nil {
			return err
		}
		event := s.buildReadyEvent(ctx, order)
		return s.repo.UpdateStatus(ctx, orderID, status, "Admin marked ready for pickup", "admin", events.TopicOrderReadyForPickup, event)
	}

	return s.repo.UpdateStatus(ctx, orderID, status, "Admin manual update", "admin", "", nil)
}

// buildReadyEvent assembles the (enriched) order.ready-for-pickup event, resolving
// the origin/destination hub coordinates the dispatch service needs to create the
// first-mile job (or the direct job).
func (s *orderService) buildReadyEvent(ctx context.Context, order *model.Order) events.OrderReadyForPickupEvent {
	evt := events.OrderReadyForPickupEvent{
		BaseEvent: events.BaseEvent{
			EventID: uuid.New().String(), EventType: events.TopicOrderReadyForPickup,
			Timestamp: time.Now(), Source: "order-service", TraceID: logger.GetTraceID(ctx)},
		OrderID: order.ID, AWB: order.AWB, SenderID: order.UserID,
		PickupLat: order.SenderLat, PickupLng: order.SenderLng, PickupAddr: order.SenderAddress,
		ReceiverAddr: order.ReceiverAddress, ReceiverLat: order.ReceiverLat, ReceiverLng: order.ReceiverLng,
		Weight: order.WeightKg, ServiceType: order.ServiceType,
		DeliveryMode: order.DeliveryMode, PickupMode: order.PickupMode,
		OriginHubName: order.OriginHubName, DestHubID: order.DestHubID, DestHubName: order.DestHubName,
	}
	if s.hubs != nil {
		if h, ok := s.hubs.Nearest(ctx, order.SenderLat, order.SenderLng); ok {
			evt.OriginHubName, evt.OriginHubLat, evt.OriginHubLng = h.Name, h.Lat, h.Lng
		}
		if h, ok := s.hubs.Nearest(ctx, order.ReceiverLat, order.ReceiverLng); ok {
			evt.DestHubID, evt.DestHubName, evt.DestHubLat, evt.DestHubLng = h.ID, h.Name, h.Lat, h.Lng
		}
	}
	return evt
}

func (s *orderService) HandlePaymentSuccess(ctx context.Context, orderID string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	event := s.buildReadyEvent(ctx, order)
	if err := s.repo.UpdateStatus(ctx, orderID, events.OrderStatusReadyForPickup, "Payment confirmed", "payment-service", events.TopicOrderReadyForPickup, event); err != nil {
		return err
	}

	return nil
}

func (s *orderService) HandleDeliveryFailed(ctx context.Context, orderID string, attempts int) error {
	if attempts >= events.MaxDeliveryAttempts {
		s.repo.UpdateStatus(ctx, orderID, events.OrderStatusReturnToSender, "Max delivery attempts reached", "system", "", nil)
		log.Printf("[Order] ⚠️ Order %s → RETURN_TO_SENDER after %d failed attempts", orderID, attempts)
	} else {
		s.repo.UpdateStatus(ctx, orderID, events.OrderStatusDeliveryFailed, fmt.Sprintf("Delivery attempt %d failed", attempts), "system", "", nil)
	}
	return s.repo.IncrementDeliveryAttempts(ctx, orderID)
}

func (s *orderService) HandlePackageDelivered(ctx context.Context, orderID string) error {
	return s.repo.MarkDelivered(ctx, orderID)
}

// HandlePackagePickedUp reacts to package.picked-up (first-mile / direct pickup):
// the courier has taken custody from the sender, so the order leaves
// READY_FOR_PICKUP for PICKED_UP. Keyed on AWB. Only advances from a pre-pickup
// state so it never regresses a package that is already further along.
func (s *orderService) HandlePackagePickedUp(ctx context.Context, awb string) error {
	order, err := s.repo.GetByAWB(ctx, awb)
	if err != nil {
		return nil // unknown AWB — ignore
	}
	if order.Status == events.OrderStatusReadyForPickup || order.Status == events.OrderStatusPendingPayment {
		return s.repo.UpdateStatus(ctx, order.ID, events.OrderStatusPickedUp, "Picked up by courier", "dispatch-service", "", nil)
	}
	return nil
}

// HandleHubScan reacts to package.scanned-at-hub and advances the order status as
// the package moves through the hub network (keyed on AWB — the console sends no
// order_id):
//   - any hub ARRIVED/DEPARTED → IN_TRANSIT (so an order in a hub no longer shows
//     READY_FOR_PICKUP / PICKED_UP)
//   - ARRIVED at the destination hub → also emit order.ready-for-last-mile so
//     dispatch opens the last-mile job
//   - DEPARTED from the destination hub (last-mile courier collected) → OUT_FOR_DELIVERY
func (s *orderService) HandleHubScan(ctx context.Context, awb, hubID, hubName, scanType string) error {
	order, err := s.repo.GetByAWB(ctx, awb)
	if err != nil {
		return nil // unknown AWB — ignore
	}
	// Never move a terminal/finished order backward.
	switch order.Status {
	case events.OrderStatusDelivered, events.OrderStatusCancelled,
		events.OrderStatusReturnToSender, events.OrderStatusLostSuspected:
		return nil
	}

	atDestHub := order.DeliveryMode != routing.RouteDirect && order.DestHubID != "" && hubID == order.DestHubID

	// Arrived at the destination hub → open the last-mile leg (idempotent in
	// dispatch via ExistsLeg) and mark IN_TRANSIT.
	if scanType == events.ScanTypeArrived && atDestHub && order.Status != events.OrderStatusOutForDelivery {
		evt := events.OrderReadyForLastMileEvent{
			BaseEvent: events.BaseEvent{
				EventID: uuid.New().String(), EventType: events.TopicOrderReadyForLastMile,
				Timestamp: time.Now(), Source: "order-service", TraceID: logger.GetTraceID(ctx)},
			OrderID: order.ID, AWB: order.AWB, HubID: hubID, HubName: hubName, PickupAddr: hubName,
			ReceiverAddr: order.ReceiverAddress, ReceiverLat: order.ReceiverLat, ReceiverLng: order.ReceiverLng,
		}
		if s.hubs != nil {
			if h, ok := s.hubs.Nearest(ctx, order.ReceiverLat, order.ReceiverLng); ok {
				evt.PickupLat, evt.PickupLng = h.Lat, h.Lng
				if evt.HubName == "" {
					evt.HubName, evt.PickupAddr = h.Name, h.Name
				}
			}
		}
		return s.repo.UpdateStatus(ctx, order.ID, events.OrderStatusInTransit, "Arrived at destination hub", "hub-service", events.TopicOrderReadyForLastMile, evt)
	}

	// Departed the destination hub with the last-mile courier → out for delivery.
	if scanType == events.ScanTypeDeparted && atDestHub {
		if order.Status != events.OrderStatusOutForDelivery {
			return s.repo.UpdateStatus(ctx, order.ID, events.OrderStatusOutForDelivery, "Out for delivery from "+hubName, "hub-service", "", nil)
		}
		return nil
	}

	// Any other hub scan (origin/intermediate arrive or depart) → in transit.
	if (scanType == events.ScanTypeArrived || scanType == events.ScanTypeDeparted) && order.Status != events.OrderStatusInTransit {
		return s.repo.UpdateStatus(ctx, order.ID, events.OrderStatusInTransit, "In transit via "+hubName, "hub-service", "", nil)
	}
	return nil
}

// RunSLAMonitor checks for packages stuck > 48 hours without status updates.
func (s *orderService) RunSLAMonitor(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lock := database.NewRedisMutex(s.redis, "lock:cron:sla_monitor")
			if !lock.Acquire(ctx, 5*time.Minute) {
				continue // Lock not acquired, another instance is processing
			}

			stuckOrders, err := s.repo.GetStuckOrders(ctx, 48*time.Hour)
			if err != nil {
				log.Printf("[SLA] Error: %v", err)
				lock.Release(ctx)
				continue
			}

			for _, order := range stuckOrders {
				log.Printf("[SLA] ⚠️ Package LOST suspected: AWB=%s, stuck since %v", order.AWB, order.UpdatedAt)
				s.repo.UpdateStatus(ctx, order.ID, events.OrderStatusLostSuspected, "No scan for 48+ hours", "sla-monitor", "", nil)

				event := events.PackageLostSuspectedEvent{
					BaseEvent: events.BaseEvent{
						EventID: uuid.New().String(), EventType: events.TopicPackageLost,
						Timestamp: time.Now(), Source: "order-service",
					},
					OrderID: order.ID, AWB: order.AWB, LastScanTime: order.UpdatedAt,
					HoursSinceUpdate: int(time.Since(order.UpdatedAt).Hours()),
				}
				s.publish(ctx, events.TopicPackageLost, order.ID, event)
			}
			lock.Release(ctx)
		}
	}
}

// RunPaymentExpiryChecker cancels orders pending payment > 24 hours (Saga Pattern).
func (s *orderService) RunPaymentExpiryChecker(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lock := database.NewRedisMutex(s.redis, "lock:cron:payment_expiry")
			if !lock.Acquire(ctx, 14*time.Minute) {
				continue // Lock not acquired
			}

			expired, err := s.repo.GetExpiredPendingOrders(ctx, 24*time.Hour)
			if err != nil {
				log.Printf("[Saga] Error: %v", err)
				lock.Release(ctx)
				continue
			}

			for _, order := range expired {
				log.Printf("[Saga] Auto-cancelling expired order AWB=%s (created %v)", order.AWB, order.CreatedAt)
				event := events.OrderCancelledEvent{
					BaseEvent: events.BaseEvent{
						EventID: uuid.New().String(), EventType: events.TopicOrderCancelled,
						Timestamp: time.Now(), Source: "order-service", TraceID: logger.GetTraceID(ctx)},
					OrderID: order.ID, Reason: "Payment timeout (24h expired)",
				}
				s.repo.MarkCancelled(ctx, order.ID, events.TopicOrderCancelled, event)
			}
			lock.Release(ctx)
		}
	}
}

func (s *orderService) GetDashboardStats(ctx context.Context) (int64, float64, error) {
	return s.repo.GetDashboardStats(ctx)
}

func (s *orderService) GetVolumeStats(ctx context.Context) ([]model.DailyVolume, error) {
	return s.repo.GetVolumeStats(ctx)
}
