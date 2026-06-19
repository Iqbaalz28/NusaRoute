package service

import (
	"fmt"
	"github.com/nusaroute/pkg/logger"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/services/hub-service/internal/model"
	"github.com/nusaroute/services/hub-service/internal/repository"
)

type HubService interface {
	Scan(ctx context.Context, req model.ScanRequest) (*model.ScanLog, error)
	GetScanHistory(ctx context.Context, awb string) ([]model.ScanLog, error)
	GetManifest(ctx context.Context, hubID string) ([]model.ScanLog, error)
	GetCurrentInventory(ctx context.Context, hubID string) ([]model.ScanLog, error)
	ListHubs(ctx context.Context) ([]model.Hub, error)
	ListAllHubs(ctx context.Context) ([]model.Hub, error)
	CreateHub(ctx context.Context, req model.HubUpsertRequest) (*model.Hub, error)
	UpdateHub(ctx context.Context, id string, req model.HubUpsertRequest) (*model.Hub, error)
	GetDashboardStats(ctx context.Context) (int64, int64, error)
}

type hubService struct {
	repo     repository.HubRepository
	producer *kafka.Producer
}

func NewHubService(repo repository.HubRepository, producer *kafka.Producer) HubService {
	return &hubService{repo: repo, producer: producer}
}

func (s *hubService) Scan(ctx context.Context, req model.ScanRequest) (*model.ScanLog, error) {
	if req.AWB == "" || req.HubID == "" || req.ScanType == "" {
		return nil, errors.New("awb, hub_id, and scan_type are required")
	}

	// Validate scan type
	if req.ScanType != events.ScanTypeArrived && req.ScanType != events.ScanTypeSorted && req.ScanType != events.ScanTypeDeparted {
		return nil, errors.New("invalid scan_type: must be ARRIVED, SORTED, or DEPARTED")
	}

	hub, err := s.repo.GetHubByID(ctx, req.HubID)
	if err != nil { return nil, errors.New("hub not found") }

	scan := &model.ScanLog{
		AWB: req.AWB, OrderID: req.OrderID, HubID: req.HubID,
		ScanType: req.ScanType, OperatorID: req.OperatorID, Note: req.Note,
	}
	// Publish event to Kafka — this is the "Event Factory"
	event := events.PackageScannedAtHubEvent{
		BaseEvent: events.BaseEvent{
			EventID: uuid.New().String(), EventType: events.TopicPackageScannedHub,
			Timestamp: time.Now(), Source: "hub-service", TraceID: logger.GetTraceID(ctx)},
		OrderID: req.OrderID, AWB: req.AWB,
		HubID: req.HubID, HubName: hub.Name,
		ScanType: req.ScanType, OperatorID: req.OperatorID,
	}

	if err := s.repo.CreateScanLog(ctx, scan, events.TopicPackageScannedHub, event); err != nil {
		return nil, err
	}

	logger.Info(context.Background(), fmt.Sprintf("[Hub] 📦 Scan %s: AWB=%s at hub=%s by operator=%s", req.ScanType, req.AWB, hub.Name, req.OperatorID))
	return scan, nil
}

func (s *hubService) GetScanHistory(ctx context.Context, awb string) ([]model.ScanLog, error) {
	return s.repo.GetScansByAWB(ctx, awb)
}

func (s *hubService) GetManifest(ctx context.Context, hubID string) ([]model.ScanLog, error) {
	return s.repo.GetManifest(ctx, hubID, time.Now())
}

func (s *hubService) GetCurrentInventory(ctx context.Context, hubID string) ([]model.ScanLog, error) {
	return s.repo.GetCurrentInventory(ctx, hubID)
}

func (s *hubService) ListHubs(ctx context.Context) ([]model.Hub, error) {
	return s.repo.ListHubs(ctx)
}

func (s *hubService) ListAllHubs(ctx context.Context) ([]model.Hub, error) {
	return s.repo.ListAllHubs(ctx)
}

func (s *hubService) CreateHub(ctx context.Context, req model.HubUpsertRequest) (*model.Hub, error) {
	if req.Name == "" || req.Code == "" || req.City == "" {
		return nil, errors.New("name, code, and city are required")
	}
	hub := &model.Hub{
		Name: req.Name, Code: req.Code, City: req.City, Province: req.Province,
		Lat: req.Lat, Lng: req.Lng, Type: req.Type, IsActive: true,
	}
	if hub.Type == "" {
		hub.Type = "SORTATION"
	}
	if req.IsActive != nil {
		hub.IsActive = *req.IsActive
	}
	if err := s.repo.CreateHub(ctx, hub); err != nil {
		return nil, err
	}
	logger.Info(context.Background(), fmt.Sprintf("[Hub] ➕ Created hub %s (%s)", hub.Name, hub.Code))
	return hub, nil
}

func (s *hubService) UpdateHub(ctx context.Context, id string, req model.HubUpsertRequest) (*model.Hub, error) {
	hub, err := s.repo.GetHubByID(ctx, id)
	if err != nil {
		return nil, errors.New("hub not found")
	}
	if req.Name != "" {
		hub.Name = req.Name
	}
	if req.Code != "" {
		hub.Code = req.Code
	}
	if req.City != "" {
		hub.City = req.City
	}
	if req.Province != "" {
		hub.Province = req.Province
	}
	if req.Type != "" {
		hub.Type = req.Type
	}
	hub.Lat, hub.Lng = req.Lat, req.Lng
	if req.IsActive != nil {
		hub.IsActive = *req.IsActive
	}
	if err := s.repo.UpdateHub(ctx, hub); err != nil {
		return nil, err
	}
	logger.Info(context.Background(), fmt.Sprintf("[Hub] ✏️ Updated hub %s (%s) active=%v", hub.Name, hub.Code, hub.IsActive))
	return hub, nil
}

func (s *hubService) GetDashboardStats(ctx context.Context) (int64, int64, error) {
	return s.repo.GetDashboardStats(ctx)
}
