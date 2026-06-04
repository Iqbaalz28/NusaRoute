package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/nusaroute/pkg/logger"

	"github.com/google/uuid"
	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/services/courier-service/pkg/grpc/pb"
	"github.com/nusaroute/services/dispatch-service/internal/model"
	"github.com/nusaroute/services/dispatch-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

type DispatchService interface {
	AutoAssign(ctx context.Context, orderID, awb string, pickupLat, pickupLng float64, pickupAddr string) error
	ManualAssign(ctx context.Context, req model.ManualAssignRequest) error
	ListAssignments(ctx context.Context, status string, page, perPage int) ([]model.Assignment, error)
	RunNoShowMonitor(ctx context.Context)
}

type dispatchService struct {
	repo          repository.DispatchRepository
	producer      *kafka.Producer
	courierClient pb.CourierServiceClient
	redis         *redis.Client
}

func NewDispatchService(repo repository.DispatchRepository, producer *kafka.Producer, courierClient pb.CourierServiceClient, r *redis.Client) DispatchService {
	return &dispatchService{repo: repo, producer: producer, courierClient: courierClient, redis: r}
}

// CourierInfo represents the courier data returned from Courier Service API.
type CourierInfo struct {
	ID       string  `json:"id"`
	FullName string  `json:"full_name"`
	Phone    string  `json:"phone"`
	Lat      float64 `json:"current_lat"`
	Lng      float64 `json:"current_lng"`
}

// AutoAssign implements a simplified VRP: find nearest available courier to pickup point.
func (s *dispatchService) AutoAssign(ctx context.Context, orderID, awb string, pickupLat, pickupLng float64, pickupAddr string) error {
	// Call Courier Service to get available couriers near pickup location
	couriers, err := s.fetchAvailableCouriers(pickupLat, pickupLng, 15.0) // 15km radius
	if err != nil || len(couriers) == 0 {
		logger.Info(context.Background(), fmt.Sprintf("[Dispatch] No available couriers near (%.4f, %.4f) for order %s", pickupLat, pickupLng, orderID))
		return fmt.Errorf("no available couriers found")
	}

	// VRP: Select nearest courier (greedy nearest-neighbor algorithm)
	bestCourier := couriers[0]
	bestDist := haversine(pickupLat, pickupLng, bestCourier.Lat, bestCourier.Lng)
	for _, c := range couriers[1:] {
		dist := haversine(pickupLat, pickupLng, c.Lat, c.Lng)
		if dist < bestDist {
			bestDist = dist
			bestCourier = c
		}
	}

	// Create assignment
	assignment := &model.Assignment{
		OrderID: orderID, AWB: awb,
		CourierID: bestCourier.ID, CourierName: bestCourier.FullName,
		PickupLat: pickupLat, PickupLng: pickupLng, PickupAddr: pickupAddr,
	}
	// Publish CourierAssigned event
	event := events.CourierAssignedEvent{
		BaseEvent: events.BaseEvent{
			EventID: uuid.New().String(), EventType: events.TopicCourierAssigned,
			Timestamp: time.Now(), Source: "dispatch-service", TraceID: logger.GetTraceID(ctx)},
		OrderID: orderID, AWB: awb,
		CourierID: bestCourier.ID, CourierName: bestCourier.FullName,
		CourierPhone:        bestCourier.Phone,
		EstimatedPickupTime: time.Now().Add(30 * time.Minute),
	}

	if err := s.repo.CreateAssignment(ctx, assignment, events.TopicCourierAssigned, event); err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	logger.Info(context.Background(), fmt.Sprintf("[Dispatch]  Assigned courier %s (%.1fkm away) to order %s", bestCourier.FullName, bestDist, orderID))
	return nil
}

func (s *dispatchService) ManualAssign(ctx context.Context, req model.ManualAssignRequest) error {
	assignment := &model.Assignment{
		OrderID: req.OrderID, CourierID: req.CourierID,
	}
	return s.repo.CreateAssignment(ctx, assignment, "", nil)
}

func (s *dispatchService) ListAssignments(ctx context.Context, status string, page, perPage int) ([]model.Assignment, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	return s.repo.ListAssignments(ctx, status, page, perPage)
}

// RunNoShowMonitor checks for couriers who haven't picked up within 2 hours and reassigns.
func (s *dispatchService) RunNoShowMonitor(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lock := database.NewRedisMutex(s.redis, "lock:cron:no_show_monitor")
			if !lock.Acquire(ctx, 9*time.Minute) {
				continue
			}

			noShows, err := s.repo.GetNoShowAssignments(ctx, 2*time.Hour)
			if err != nil {
				logger.Info(context.Background(), fmt.Sprintf("[Dispatch] NoShow check error: %v", err))
				lock.Release(ctx)
				continue
			}

			for _, a := range noShows {
				logger.Info(context.Background(), fmt.Sprintf("[Dispatch] ⚠️ Courier %s no-show for order %s, reassigning...", a.CourierID, a.OrderID))
				event := events.CourierReassignedEvent{
					BaseEvent: events.BaseEvent{
						EventID: uuid.New().String(), EventType: events.TopicCourierReassigned,
						Timestamp: time.Now(), Source: "dispatch-service", TraceID: logger.GetTraceID(ctx)},
					OrderID: a.OrderID, OldCourierID: a.CourierID,
					Reason: "No-show: courier did not pick up within 2 hours",
				}
				s.repo.UpdateStatus(ctx, a.ID, model.AssignmentStatusNoShow, events.TopicCourierReassigned, event)

				// Try to reassign to another courier
				err := s.AutoAssign(ctx, a.OrderID, a.AWB, a.PickupLat, a.PickupLng, a.PickupAddr)
				if err != nil {
					logger.Info(context.Background(), fmt.Sprintf("[Dispatch] Failed to reassign order %s: %v", a.OrderID, err))
					continue
				}
			}
			lock.Release(ctx)
		}
	}
}

func (s *dispatchService) fetchAvailableCouriers(lat, lng, radiusKm float64) ([]CourierInfo, error) {
	req := &pb.AvailableCouriersRequest{
		Latitude:  lat,
		Longitude: lng,
		RadiusKm:  radiusKm,
	}

	resp, err := s.courierClient.GetAvailableCouriers(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("gRPC error: %w", err)
	}

	var result []CourierInfo
	for _, c := range resp.Data {
		result = append(result, CourierInfo{
			ID:       c.Id,
			FullName: c.FullName,
			Phone:    c.Phone,
			Lat:      c.CurrentLat,
			Lng:      c.CurrentLng,
		})
	}

	return result, nil
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
