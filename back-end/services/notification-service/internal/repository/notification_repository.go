package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NotificationLog represents a sent notification record in MongoDB.
type NotificationLog struct {
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"user_id" bson:"user_id"`
	Channel   string    `json:"channel" bson:"channel"` // EMAIL, WHATSAPP, PUSH
	Title     string    `json:"title" bson:"title"`
	Message   string    `json:"message" bson:"message"`
	OrderID   string    `json:"order_id,omitempty" bson:"order_id,omitempty"`
	AWB       string    `json:"awb,omitempty" bson:"awb,omitempty"`
	Status    string    `json:"status" bson:"status"` // SENT, FAILED
	IsRead    bool      `json:"is_read" bson:"is_read"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type NotificationRepository interface {
	InsertLog(ctx context.Context, log NotificationLog) error
	GetLogsByUserID(ctx context.Context, userID string, limit int64) ([]NotificationLog, error)
	MarkAsRead(ctx context.Context, id string) error
}

type notificationRepository struct {
	db *mongo.Database
}

func NewNotificationRepository(db *mongo.Database) NotificationRepository {
	collection := db.Collection("notification_logs")
	collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
	})
	return &notificationRepository{db: db}
}

func (r *notificationRepository) InsertLog(ctx context.Context, log NotificationLog) error {
	_, err := r.db.Collection("notification_logs").InsertOne(ctx, log)
	return err
}

func (r *notificationRepository) GetLogsByUserID(ctx context.Context, userID string, limit int64) ([]NotificationLog, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit)
	cursor, err := r.db.Collection("notification_logs").Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []NotificationLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id string) error {
	_, err := r.db.Collection("notification_logs").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"is_read": true}})
	return err
}
