//go:build unit

package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nusaroute/services/tracking-service/internal/model"
	"github.com/nusaroute/services/tracking-service/internal/service"
)

type MockTrackingRepository struct {
	events map[string][]model.TrackingEvent
	gps    map[string]*model.CourierGPS
}

func NewMockTrackingRepo() *MockTrackingRepository {
	return &MockTrackingRepository{
		events: make(map[string][]model.TrackingEvent),
		gps:    make(map[string]*model.CourierGPS),
	}
}

func (m *MockTrackingRepository) InsertEvent(ctx context.Context, event model.TrackingEvent) error {
	m.events[event.AWB] = append(m.events[event.AWB], event)
	return nil
}

func (m *MockTrackingRepository) GetEventsByAWB(ctx context.Context, awb string) ([]model.TrackingEvent, error) {
	if evts, ok := m.events[awb]; ok {
		return evts, nil
	}
	return nil, nil
}

func (m *MockTrackingRepository) SetGPS(ctx context.Context, gps model.CourierGPS) error {
	m.gps[gps.AWB] = &gps
	return nil
}

func (m *MockTrackingRepository) GetGPS(ctx context.Context, awb string) (*model.CourierGPS, error) {
	if gps, ok := m.gps[awb]; ok {
		return gps, nil
	}
	return nil, errors.New("not found")
}

func TestRecordEvent_Success(t *testing.T) {
	repo := NewMockTrackingRepo()
	svc := service.NewTrackingService(repo)

	err := svc.RecordEvent(context.Background(), model.TrackingEvent{
		AWB: "AWB123", OrderID: "ORD123", Status: "IN_TRANSIT",
		Location: "Jakarta",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events, _ := repo.GetEventsByAWB(context.Background(), "AWB123")
	if len(events) != 1 {
		t.Fatal("expected 1 event recorded")
	}
	if events[0].ID == "" {
		t.Error("expected ID to be generated")
	}
	if events[0].Timestamp.IsZero() {
		t.Error("expected timestamp to be generated")
	}
}

func TestGetTimeline_Success(t *testing.T) {
	repo := NewMockTrackingRepo()
	svc := service.NewTrackingService(repo)

	svc.RecordEvent(context.Background(), model.TrackingEvent{
		AWB: "AWB123", OrderID: "ORD123", Status: "PICKED_UP", Timestamp: time.Now().Add(-2 * time.Hour),
	})
	svc.RecordEvent(context.Background(), model.TrackingEvent{
		AWB: "AWB123", OrderID: "ORD123", Status: "IN_TRANSIT", Timestamp: time.Now().Add(-1 * time.Hour),
	})
	svc.UpdateGPS(context.Background(), model.CourierGPS{AWB: "AWB123", CourierID: "C1", Lat: -6.1, Lng: 106.8})

	timeline, err := svc.GetTimeline(context.Background(), "AWB123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if timeline.Status != "IN_TRANSIT" {
		t.Errorf("expected IN_TRANSIT, got %s", timeline.Status)
	}
	if len(timeline.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(timeline.Events))
	}
	if timeline.LiveGPS == nil {
		t.Error("expected LiveGPS attached")
	} else if timeline.LiveGPS.Lat != -6.1 {
		t.Error("wrong GPS lat")
	}
}

func TestGetTimeline_NotFound(t *testing.T) {
	repo := NewMockTrackingRepo()
	svc := service.NewTrackingService(repo)

	_, err := svc.GetTimeline(context.Background(), "AWB999")
	if err == nil {
		t.Error("expected error for non-existent AWB")
	}
}
