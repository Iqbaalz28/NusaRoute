package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/nusaroute/services/tracking-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collectionName = "tracking_events"
	gpsTTL         = 30 * time.Second
)

type TrackingService interface {
	RecordEvent(ctx context.Context, event model.TrackingEvent) error
	GetTimeline(ctx context.Context, awb string) (*model.TrackingTimeline, error)
	UpdateGPS(ctx context.Context, gps model.CourierGPS) error
	GetLiveGPS(ctx context.Context, awb string) (*model.CourierGPS, error)
}

type trackingService struct {
	db    *mongo.Database
	redis *redis.Client
}

func NewTrackingService(db *mongo.Database, redis *redis.Client) TrackingService {
	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := db.Collection(collectionName)
	collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "awb", Value: 1}}},
		{Keys: bson.D{{Key: "order_id", Value: 1}}},
		{Keys: bson.D{{Key: "timestamp", Value: 1}}},
	})

	return &trackingService{db: db, redis: redis}
}

func (s *trackingService) RecordEvent(ctx context.Context, event model.TrackingEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	collection := s.db.Collection(collectionName)
	_, err := collection.InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to record tracking event: %w", err)
	}

	log.Printf("[Tracking] 📍 Recorded: AWB=%s status=%s location=%s", event.AWB, event.Status, event.Location)
	return nil
}

func (s *trackingService) GetTimeline(ctx context.Context, awb string) (*model.TrackingTimeline, error) {
	collection := s.db.Collection(collectionName)

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	cursor, err := collection.Find(ctx, bson.M{"awb": awb}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracking events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []model.TrackingEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no tracking events found for AWB %s", awb)
	}

	// Get current status (last event)
	currentStatus := events[len(events)-1].Status
	orderID := events[0].OrderID

	timeline := &model.TrackingTimeline{
		AWB:     awb,
		OrderID: orderID,
		Status:  currentStatus,
		Events:  events,
	}

	// Attach live GPS if available
	gps, err := s.GetLiveGPS(ctx, awb)
	if err == nil && gps != nil {
		timeline.LiveGPS = gps
	}

	return timeline, nil
}

func (s *trackingService) UpdateGPS(ctx context.Context, gps model.CourierGPS) error {
	gps.Timestamp = time.Now().Unix()
	data, err := json.Marshal(gps)
	if err != nil { return err }

	key := fmt.Sprintf("gps:%s", gps.AWB)
	return s.redis.Set(ctx, key, string(data), gpsTTL).Err()
}

func (s *trackingService) GetLiveGPS(ctx context.Context, awb string) (*model.CourierGPS, error) {
	key := fmt.Sprintf("gps:%s", awb)
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil { return nil, err }

	var gps model.CourierGPS
	if err := json.Unmarshal([]byte(data), &gps); err != nil { return nil, err }
	return &gps, nil
}
