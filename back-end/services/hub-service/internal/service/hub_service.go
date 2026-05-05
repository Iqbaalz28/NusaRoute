package service

import (
	"context"
	"errors"
	"log"
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
	ListHubs(ctx context.Context) ([]model.Hub, error)
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
	if err := s.repo.CreateScanLog(ctx, scan); err != nil {
		return nil, err
	}

	// Publish event to Kafka — this is the "Event Factory"
	event := events.PackageScannedAtHubEvent{
		BaseEvent: events.BaseEvent{
			EventID: uuid.New().String(), EventType: events.TopicPackageScannedHub,
			Timestamp: time.Now(), Source: "hub-service",
		},
		OrderID: req.OrderID, AWB: req.AWB,
		HubID: req.HubID, HubName: hub.Name,
		ScanType: req.ScanType, OperatorID: req.OperatorID,
	}
	s.producer.Publish(ctx, events.TopicPackageScannedHub, req.AWB, event)

	log.Printf("[Hub] 📦 Scan %s: AWB=%s at hub=%s by operator=%s", req.ScanType, req.AWB, hub.Name, req.OperatorID)
	return scan, nil
}

func (s *hubService) GetScanHistory(ctx context.Context, awb string) ([]model.ScanLog, error) {
	return s.repo.GetScansByAWB(ctx, awb)
}

func (s *hubService) GetManifest(ctx context.Context, hubID string) ([]model.ScanLog, error) {
	return s.repo.GetManifest(ctx, hubID, time.Now())
}

func (s *hubService) ListHubs(ctx context.Context) ([]model.Hub, error) {
	return s.repo.ListHubs(ctx)
}
