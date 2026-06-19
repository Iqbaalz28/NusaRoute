package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/nusaroute/pkg/logger"

	"github.com/google/uuid"
	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/services/courier-service/pkg/grpc/pb"
	"github.com/nusaroute/services/dispatch-service/internal/broker"
	"github.com/nusaroute/services/dispatch-service/internal/model"
	"github.com/nusaroute/services/dispatch-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

type DispatchService interface {
	AutoAssign(ctx context.Context, orderID, awb string, pickupLat, pickupLng float64, pickupAddr string) error
	CreateJobFromReady(ctx context.Context, evt events.OrderReadyForPickupEvent) error
	CreateLastMileJob(ctx context.Context, evt events.OrderReadyForLastMileEvent) error
	ListOpenJobs(ctx context.Context) ([]model.Assignment, error)
	ListMyAssignments(ctx context.Context, courierID string) ([]model.Assignment, error)
	ClaimJob(ctx context.Context, req model.ClaimRequest, courierID string) (*model.Assignment, error)
	ManualAssign(ctx context.Context, req model.ManualAssignRequest) error
	PickupPackage(ctx context.Context, req model.PickupRequest) error
	DeliverPackage(ctx context.Context, req model.DeliverRequest) error
	HandleHubArrival(ctx context.Context, awb, hubName string) error
	ListAssignments(ctx context.Context, status string, page, perPage int) ([]model.Assignment, error)
	RunNoShowMonitor(ctx context.Context)
}

type dispatchService struct {
	repo          repository.DispatchRepository
	producer      *kafka.Producer
	courierClient pb.CourierServiceClient
	redis         *redis.Client
	broker        *broker.Broker
}

func NewDispatchService(repo repository.DispatchRepository, producer *kafka.Producer, courierClient pb.CourierServiceClient, r *redis.Client, b *broker.Broker) DispatchService {
	return &dispatchService{repo: repo, producer: producer, courierClient: courierClient, redis: r, broker: b}
}

// ErrJobTaken is re-exported so the HTTP layer can map it to 409 Conflict.
var ErrJobTaken = repository.ErrJobTaken

// createOpenLeg persists an unclaimed leg and pushes it to connected couriers via
// the SSE broker so they can race to claim it (first-come-first-served).
func (s *dispatchService) createOpenLeg(ctx context.Context, job *model.Assignment) error {
	if err := s.repo.CreateOpenJob(ctx, job); err != nil {
		return fmt.Errorf("failed to create open leg: %w", err)
	}
	s.broadcastJob(job)
	logger.Info(context.Background(), fmt.Sprintf("[Dispatch] 📢 OPEN %s job for order %s (AWB=%s)", job.Leg, job.OrderID, job.AWB))
	return nil
}

// CreateJobFromReady opens the first leg when an order is ready:
//   - DIRECT (same-city instant): one courier sender → receiver.
//   - VIA_HUB + courier pickup: first-mile job sender → origin hub.
//   - VIA_HUB + self-dropoff: no job (sender brings it; origin hub scans it in).
func (s *dispatchService) CreateJobFromReady(ctx context.Context, evt events.OrderReadyForPickupEvent) error {
	if evt.DeliveryMode == "DIRECT" {
		return s.createOpenLeg(ctx, &model.Assignment{
			OrderID: evt.OrderID, AWB: evt.AWB, Leg: model.LegDirect,
			PickupLat: evt.PickupLat, PickupLng: evt.PickupLng, PickupAddr: evt.PickupAddr,
			DropoffLat: evt.ReceiverLat, DropoffLng: evt.ReceiverLng, DropoffAddr: evt.ReceiverAddr,
		})
	}
	if evt.PickupMode == "SELF_DROPOFF" {
		logger.Info(context.Background(), fmt.Sprintf("[Dispatch] Order %s is self-dropoff; awaiting origin-hub scan (no first-mile job)", evt.OrderID))
		return nil
	}
	return s.createOpenLeg(ctx, &model.Assignment{
		OrderID: evt.OrderID, AWB: evt.AWB, Leg: model.LegFirstMile,
		PickupLat: evt.PickupLat, PickupLng: evt.PickupLng, PickupAddr: evt.PickupAddr,
		DropoffLat: evt.OriginHubLat, DropoffLng: evt.OriginHubLng, DropoffAddr: evt.OriginHubName,
		HubName: evt.OriginHubName,
	})
}

// CreateLastMileJob opens the last-mile leg (dest hub → receiver) when a package
// arrives at its destination hub. Idempotent: skips if the leg already exists.
func (s *dispatchService) CreateLastMileJob(ctx context.Context, evt events.OrderReadyForLastMileEvent) error {
	if exists, _ := s.repo.ExistsLeg(ctx, evt.OrderID, model.LegLastMile); exists {
		return nil
	}
	return s.createOpenLeg(ctx, &model.Assignment{
		OrderID: evt.OrderID, AWB: evt.AWB, Leg: model.LegLastMile,
		PickupLat: evt.PickupLat, PickupLng: evt.PickupLng, PickupAddr: evt.HubName,
		DropoffLat: evt.ReceiverLat, DropoffLng: evt.ReceiverLng, DropoffAddr: evt.ReceiverAddr,
		HubID: evt.HubID, HubName: evt.HubName,
	})
}

func (s *dispatchService) broadcastJob(job *model.Assignment) {
	if s.broker == nil {
		return
	}
	if data, err := json.Marshal(job); err == nil {
		s.broker.Broadcast(data)
	}
}

func (s *dispatchService) ListOpenJobs(ctx context.Context) ([]model.Assignment, error) {
	return s.repo.ListOpenJobs(ctx)
}

func (s *dispatchService) ListMyAssignments(ctx context.Context, courierID string) ([]model.Assignment, error) {
	return s.repo.ListByCourier(ctx, courierID)
}

// ClaimJob lets a courier atomically take an OPEN job. On success it emits
// CourierAssigned so the customer is notified the courier is on the way.
func (s *dispatchService) ClaimJob(ctx context.Context, req model.ClaimRequest, courierID string) (*model.Assignment, error) {
	buildEvent := func(a *model.Assignment) interface{} {
		return events.CourierAssignedEvent{
			BaseEvent: events.BaseEvent{
				EventID: uuid.New().String(), EventType: events.TopicCourierAssigned,
				Timestamp: time.Now(), Source: "dispatch-service", TraceID: logger.GetTraceID(ctx)},
			OrderID: a.OrderID, AWB: a.AWB,
			CourierID: a.CourierID, CourierName: a.CourierName,
			EstimatedPickupTime: time.Now().Add(30 * time.Minute),
		}
	}
	a, err := s.repo.ClaimJob(ctx, req.OrderID, courierID, req.CourierName, events.TopicCourierAssigned, buildEvent)
	if err != nil {
		return nil, err
	}
	logger.Info(context.Background(), fmt.Sprintf("[Dispatch] ✋ Courier %s (%s) claimed order %s", req.CourierName, courierID, req.OrderID))
	return a, nil
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

// PickupPackage marks the courier's assignment as picked up and publishes
// package.picked-up so the tracking timeline records the PICKED_UP stage.
func (s *dispatchService) PickupPackage(ctx context.Context, req model.PickupRequest) error {
	a, err := s.repo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return fmt.Errorf("no assignment found for order %s (courier not assigned yet): %w", req.OrderID, err)
	}

	// The courier must scan the sender's digital AWB; it has to match the job.
	if !awbMatches(req.AWB, a.AWB) {
		return fmt.Errorf("scanned AWB (%s) does not match this assignment (%s)", req.AWB, a.AWB)
	}

	lat, lng := req.Lat, req.Lng
	if lat == 0 && lng == 0 {
		lat, lng = a.PickupLat, a.PickupLng
	}

	var topic string
	var event interface{}
	if a.Leg == model.LegLastMile {
		// Last-mile pickup = collecting from the destination hub for delivery.
		topic = events.TopicPackageScannedHub
		event = events.PackageScannedAtHubEvent{
			BaseEvent: events.BaseEvent{
				EventID: uuid.New().String(), EventType: events.TopicPackageScannedHub,
				Timestamp: time.Now(), Source: "dispatch-service", TraceID: logger.GetTraceID(ctx)},
			OrderID: a.OrderID, AWB: a.AWB, HubID: a.HubID, HubName: a.HubName,
			ScanType: events.ScanTypeDeparted, OperatorID: a.CourierName,
		}
	} else {
		// First-mile / direct pickup = taking custody from the sender.
		topic = events.TopicPackagePickedUp
		event = events.PackagePickedUpEvent{
			BaseEvent: events.BaseEvent{
				EventID: uuid.New().String(), EventType: events.TopicPackagePickedUp,
				Timestamp: time.Now(), Source: "dispatch-service", TraceID: logger.GetTraceID(ctx)},
			OrderID: a.OrderID, AWB: a.AWB, CourierID: a.CourierID,
			PickedAt: time.Now(), Lat: lat, Lng: lng,
		}
	}

	if err := s.repo.MarkPickedUp(ctx, a.ID, topic, event); err != nil {
		return fmt.Errorf("failed to mark picked up: %w", err)
	}

	logger.Info(context.Background(), fmt.Sprintf("[Dispatch] 📦 Courier %s picked up %s leg of order %s (AWB=%s)", a.CourierName, a.Leg, a.OrderID, a.AWB))
	return nil
}

// DeliverPackage marks the assignment complete and publishes package.delivered
// so tracking records DELIVERED and the order is finalized.
func (s *dispatchService) DeliverPackage(ctx context.Context, req model.DeliverRequest) error {
	a, err := s.repo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return fmt.Errorf("no assignment found for order %s: %w", req.OrderID, err)
	}

	// The courier must scan the package AWB at the doorstep to confirm delivery.
	if !awbMatches(req.AWB, a.AWB) {
		return fmt.Errorf("scanned AWB (%s) does not match this assignment (%s)", req.AWB, a.AWB)
	}

	// First-mile handover to the origin hub is recorded by the HUB OPERATOR's
	// inbound scan (the receiving party scans), not by the courier — so there is no
	// courier "deliver" step for the first leg. The leg is closed via HandleHubArrival.
	if a.Leg == model.LegFirstMile {
		return fmt.Errorf("paket leg-pertama diserahkan ke hub asal (operator hub yang scan masuk); kurir tidak melakukan scan antar")
	}

	// Direct / last-mile dropoff = final delivery to the receiver.
	event := events.PackageDeliveredEvent{
		BaseEvent: events.BaseEvent{
			EventID: uuid.New().String(), EventType: events.TopicPackageDelivered,
			Timestamp: time.Now(), Source: "dispatch-service", TraceID: logger.GetTraceID(ctx)},
		OrderID: a.OrderID, AWB: a.AWB, CourierID: a.CourierID,
		ReceiverName: req.ReceiverName, Lat: req.Lat, Lng: req.Lng, GeofenceValid: true,
	}

	if err := s.repo.MarkCompleted(ctx, a.ID, events.TopicPackageDelivered, event); err != nil {
		return fmt.Errorf("failed to mark completed: %w", err)
	}

	logger.Info(context.Background(), fmt.Sprintf("[Dispatch] ✅ Courier %s completed %s leg of order %s (AWB=%s)", a.CourierName, a.Leg, a.OrderID, a.AWB))
	return nil
}

// HandleHubArrival reacts to package.scanned-at-hub (ARRIVED). When a hub operator
// scans a package IN, the hub takes custody — which completes the courier's
// first-mile leg (sender → origin hub). The operator performs this scan; the
// courier does not scan at the hub. No-op when there is no active first-mile leg
// (self-dropoff orders, or the destination-hub arrival of a package already past
// its first leg).
func (s *dispatchService) HandleHubArrival(ctx context.Context, awb, hubName string) error {
	done, err := s.repo.CompleteActiveLeg(ctx, awb, model.LegFirstMile)
	if err != nil {
		return fmt.Errorf("failed to complete first-mile leg for AWB %s: %w", awb, err)
	}
	if done {
		logger.Info(context.Background(), fmt.Sprintf("[Dispatch] 🏢 First-mile handover complete for AWB %s at %s", awb, hubName))
	}
	return nil
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
				logger.Info(context.Background(), fmt.Sprintf("[Dispatch] ⚠️ Courier %s no-show for order %s, reopening job...", a.CourierID, a.OrderID))
				event := events.CourierReassignedEvent{
					BaseEvent: events.BaseEvent{
						EventID: uuid.New().String(), EventType: events.TopicCourierReassigned,
						Timestamp: time.Now(), Source: "dispatch-service", TraceID: logger.GetTraceID(ctx)},
					OrderID: a.OrderID, OldCourierID: a.CourierID,
					Reason: "No-show: courier did not pick up within 2 hours",
				}
				// Put the job back on the board so another courier can claim it.
				if err := s.repo.UpdateStatus(ctx, a.ID, model.AssignmentStatusOpen, events.TopicCourierReassigned, event); err != nil {
					logger.Info(context.Background(), fmt.Sprintf("[Dispatch] Failed to reopen order %s: %v", a.OrderID, err))
					continue
				}
				s.broadcastJob(&a)
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

// awbMatches compares a scanned AWB to the expected one, ignoring case and
// surrounding whitespace (QR scanners sometimes append trailing characters).
func awbMatches(scanned, expected string) bool {
	return strings.EqualFold(strings.TrimSpace(scanned), strings.TrimSpace(expected))
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
