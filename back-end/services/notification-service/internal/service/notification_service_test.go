//go:build unit

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nusaroute/services/notification-service/internal/repository"
	"github.com/nusaroute/services/notification-service/internal/service"
)

type MockNotificationRepository struct {
	logs map[string]*repository.NotificationLog
}

func NewMockNotificationRepo() *MockNotificationRepository {
	return &MockNotificationRepository{logs: make(map[string]*repository.NotificationLog)}
}

func (m *MockNotificationRepository) InsertLog(ctx context.Context, log repository.NotificationLog) error {
	m.logs[log.ID] = &log
	return nil
}

func (m *MockNotificationRepository) GetLogsByUserID(ctx context.Context, userID string, limit int64) ([]repository.NotificationLog, error) {
	var result []repository.NotificationLog
	for _, l := range m.logs {
		if l.UserID == userID {
			result = append(result, *l)
		}
	}
	return result, nil
}

func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, id string) error {
	if l, ok := m.logs[id]; ok {
		l.IsRead = true
		return nil
	}
	return errors.New("not found")
}

func TestSendNotification_Success(t *testing.T) {
	repo := NewMockNotificationRepo()
	svc := service.NewNotificationService(repo)

	err := svc.SendNotification(context.Background(), "user-1", "EMAIL", "Test Title", "Test Message", "ord-1", "AWB123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logs, _ := repo.GetLogsByUserID(context.Background(), "user-1", 10)
	if len(logs) != 1 {
		t.Fatal("expected 1 notification log")
	}
	if logs[0].Channel != "EMAIL" {
		t.Errorf("expected EMAIL, got %s", logs[0].Channel)
	}
}

func TestMarkAsRead(t *testing.T) {
	repo := NewMockNotificationRepo()
	svc := service.NewNotificationService(repo)

	svc.SendNotification(context.Background(), "user-1", "PUSH", "Test", "Message", "ord-1", "AWB123")
	
	logs, _ := repo.GetLogsByUserID(context.Background(), "user-1", 10)
	id := logs[0].ID

	err := svc.MarkAsRead(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logsAfter, _ := repo.GetLogsByUserID(context.Background(), "user-1", 10)
	if !logsAfter[0].IsRead {
		t.Error("expected notification to be read")
	}
}
