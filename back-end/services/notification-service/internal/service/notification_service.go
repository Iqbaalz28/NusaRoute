package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nusaroute/pkg/logger"
	"github.com/nusaroute/services/notification-service/internal/broker"
	"github.com/nusaroute/services/notification-service/internal/repository"
)

type NotificationService interface {
	SendNotification(ctx context.Context, userID, channel, title, message, orderID, awb string) error
	GetNotifications(ctx context.Context, userID string) ([]repository.NotificationLog, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
}

type notificationService struct {
	repo   repository.NotificationRepository
	broker *broker.Broker
}

// NewNotificationService takes an optional broker (may be nil in tests) used to
// push notifications to the recipient in real time over SSE.
func NewNotificationService(repo repository.NotificationRepository, b *broker.Broker) NotificationService {
	return &notificationService{repo: repo, broker: b}
}

func (s *notificationService) SendNotification(ctx context.Context, userID, channel, title, message, orderID, awb string) error {
	notif := repository.NotificationLog{
		ID:        uuid.New().String(),
		UserID:    userID,
		Channel:   channel,
		Title:     title,
		Message:   message,
		OrderID:   orderID,
		AWB:       awb,
		Status:    "SENT",
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	// In production: integrate with Twilio (WhatsApp), Firebase (Push), SMTP (Email)
	logger.Info(context.Background(), fmt.Sprintf("[Notification] 📨 [%s→%s] %s: %s", channel, userID, title, message))

	if err := s.repo.InsertLog(ctx, notif); err != nil {
		return err
	}

	// Push it live to the recipient's open tabs.
	if s.broker != nil && userID != "" {
		if data, err := json.Marshal(notif); err == nil {
			s.broker.Publish(userID, data)
		}
	}
	return nil
}

func (s *notificationService) GetNotifications(ctx context.Context, userID string) ([]repository.NotificationLog, error) {
	return s.repo.GetLogsByUserID(ctx, userID, 50)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id string) error {
	return s.repo.MarkAsRead(ctx, id)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}
