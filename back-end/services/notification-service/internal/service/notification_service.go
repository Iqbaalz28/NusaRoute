package service

import (
	"fmt"
	"github.com/nusaroute/pkg/logger"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nusaroute/services/notification-service/internal/repository"
)

type NotificationService interface {
	SendNotification(ctx context.Context, userID, channel, title, message, orderID, awb string) error
	GetNotifications(ctx context.Context, userID string) ([]repository.NotificationLog, error)
	MarkAsRead(ctx context.Context, id string) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
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
	logger.Info(context.Background(), fmt.Sprintf("[Notification] 📨 [%s] %s: %s", channel, title, message))

	return s.repo.InsertLog(ctx, notif)
}

func (s *notificationService) GetNotifications(ctx context.Context, userID string) ([]repository.NotificationLog, error) {
	return s.repo.GetLogsByUserID(ctx, userID, 50)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id string) error {
	return s.repo.MarkAsRead(ctx, id)
}
