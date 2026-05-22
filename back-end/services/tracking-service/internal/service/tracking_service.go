package service

import (
	"github.com/nusaroute/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nusaroute/services/tracking-service/internal/model"
	"github.com/nusaroute/services/tracking-service/internal/repository"
)

type TrackingService interface {
	RecordEvent(ctx context.Context, event model.TrackingEvent) error
	GetTimeline(ctx context.Context, awb string) (*model.TrackingTimeline, error)
	UpdateGPS(ctx context.Context, gps model.CourierGPS) error
	GetLiveGPS(ctx context.Context, awb string) (*model.CourierGPS, error)
}

type trackingService struct {
	repo repository.TrackingRepository
}

func NewTrackingService(repo repository.TrackingRepository) TrackingService {
	return &trackingService{repo: repo}
}

func (s *trackingService) RecordEvent(ctx context.Context, event model.TrackingEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	if err := s.repo.InsertEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to record tracking event: %w", err)
	}

	logger.Info(context.Background(), fmt.Sprintf("[Tracking] 📍 Recorded: AWB=%s status=%s location=%s", event.AWB, event.Status, event.Location))
	return nil
}

func (s *trackingService) GetTimeline(ctx context.Context, awb string) (*model.TrackingTimeline, error) {
	events, err := s.repo.GetEventsByAWB(ctx, awb)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracking events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no tracking events found for AWB %s", awb)
	}

	currentStatus := events[len(events)-1].Status
	orderID := events[0].OrderID

	timeline := &model.TrackingTimeline{
		AWB:     awb,
		OrderID: orderID,
		Status:  currentStatus,
		Events:  events,
	}

	gps, err := s.GetLiveGPS(ctx, awb)
	if err == nil && gps != nil {
		timeline.LiveGPS = gps
	}

	return timeline, nil
}

func (s *trackingService) UpdateGPS(ctx context.Context, gps model.CourierGPS) error {
	gps.Timestamp = time.Now().Unix()
	return s.repo.SetGPS(ctx, gps)
}

func (s *trackingService) GetLiveGPS(ctx context.Context, awb string) (*model.CourierGPS, error) {
	return s.repo.GetGPS(ctx, awb)
}
