//go:build unit

package service_test

import (
	"context"
	"testing"

	"github.com/nusaroute/services/hub-service/internal/model"
	"github.com/nusaroute/services/hub-service/internal/service"
	"github.com/nusaroute/pkg/events"
	"time"
	"errors"
)

type MockHubRepository struct {
	hubs  map[string]*model.Hub
	scans []model.ScanLog
}

func NewMockHubRepo() *MockHubRepository {
	return &MockHubRepository{
		hubs: map[string]*model.Hub{
			"hub-jkt-1": {ID: "hub-jkt-1", Name: "Hub Jakarta Pusat", Code: "JKT-01", City: "Jakarta", IsActive: true},
			"hub-bdg-1": {ID: "hub-bdg-1", Name: "Hub Bandung", Code: "BDG-01", City: "Bandung", IsActive: true},
		},
		scans: make([]model.ScanLog, 0),
	}
}

func (m *MockHubRepository) GetHubByID(ctx context.Context, id string) (*model.Hub, error) {
	h, ok := m.hubs[id]
	if !ok { return nil, errors.New("not found") }
	return h, nil
}

func (m *MockHubRepository) ListHubs(ctx context.Context) ([]model.Hub, error) {
	var result []model.Hub
	for _, h := range m.hubs { result = append(result, *h) }
	return result, nil
}

func (m *MockHubRepository) CreateScanLog(ctx context.Context, scan *model.ScanLog, outboxTopic string, outboxPayload interface{}) error {
	scan.ID = "scan-test-id"
	scan.ScannedAt = time.Now()
	m.scans = append(m.scans, *scan)
	return nil
}

func (m *MockHubRepository) GetScansByAWB(ctx context.Context, awb string) ([]model.ScanLog, error) {
	var result []model.ScanLog
	for _, s := range m.scans {
		if s.AWB == awb { result = append(result, s) }
	}
	return result, nil
}

func (m *MockHubRepository) GetManifest(ctx context.Context, hubID string, date time.Time) ([]model.ScanLog, error) {
	var result []model.ScanLog
	for _, s := range m.scans {
		if s.HubID == hubID { result = append(result, s) }
	}
	return result, nil
}

func TestScan_Inbound_Success(t *testing.T) {
	repo := NewMockHubRepo()
	svc := service.NewHubService(repo, nil)

	scan, err := svc.Scan(context.Background(), model.ScanRequest{
		AWB: "NR0000000001", OrderID: "order-1", HubID: "hub-jkt-1",
		ScanType: events.ScanTypeArrived, OperatorID: "operator-1",
	})

	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if scan.AWB != "NR0000000001" { t.Errorf("wrong AWB: %s", scan.AWB) }
	if scan.ScanType != "ARRIVED" { t.Errorf("wrong scan type: %s", scan.ScanType) }
}

func TestScan_InvalidScanType(t *testing.T) {
	repo := NewMockHubRepo()
	svc := service.NewHubService(repo, nil)

	_, err := svc.Scan(context.Background(), model.ScanRequest{
		AWB: "NR0000000001", HubID: "hub-jkt-1", ScanType: "INVALID",
	})
	if err == nil { t.Error("expected error for invalid scan type") }
}

func TestScan_HubNotFound(t *testing.T) {
	repo := NewMockHubRepo()
	svc := service.NewHubService(repo, nil)

	_, err := svc.Scan(context.Background(), model.ScanRequest{
		AWB: "NR0000000001", HubID: "nonexistent", ScanType: events.ScanTypeArrived,
	})
	if err == nil { t.Error("expected error for non-existent hub") }
}

func TestScan_MissingFields(t *testing.T) {
	repo := NewMockHubRepo()
	svc := service.NewHubService(repo, nil)

	_, err := svc.Scan(context.Background(), model.ScanRequest{})
	if err == nil { t.Error("expected error for missing fields") }
}

func TestGetScanHistory(t *testing.T) {
	repo := NewMockHubRepo()
	svc := service.NewHubService(repo, nil)

	// Create scans
	svc.Scan(context.Background(), model.ScanRequest{
		AWB: "NR1111111111", OrderID: "order-1", HubID: "hub-jkt-1",
		ScanType: events.ScanTypeArrived, OperatorID: "op-1",
	})
	svc.Scan(context.Background(), model.ScanRequest{
		AWB: "NR1111111111", OrderID: "order-1", HubID: "hub-jkt-1",
		ScanType: events.ScanTypeSorted, OperatorID: "op-1",
	})

	history, err := svc.GetScanHistory(context.Background(), "NR1111111111")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(history) != 2 { t.Errorf("expected 2 scans, got %d", len(history)) }
}

func TestListHubs(t *testing.T) {
	repo := NewMockHubRepo()
	svc := service.NewHubService(repo, nil)

	hubs, err := svc.ListHubs(context.Background())
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(hubs) != 2 { t.Errorf("expected 2 hubs, got %d", len(hubs)) }
}
