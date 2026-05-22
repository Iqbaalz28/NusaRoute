package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nusaroute/services/tracking-service/internal/model"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collectionName = "tracking_events"
	gpsTTL         = 30 * time.Second
)

type TrackingRepository interface {
	InsertEvent(ctx context.Context, event model.TrackingEvent) error
	GetEventsByAWB(ctx context.Context, awb string) ([]model.TrackingEvent, error)
	SetGPS(ctx context.Context, gps model.CourierGPS) error
	GetGPS(ctx context.Context, awb string) (*model.CourierGPS, error)
}

type trackingRepository struct {
	db    *mongo.Database
	redis *redis.Client
}

func NewTrackingRepository(db *mongo.Database, rdb *redis.Client) TrackingRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := db.Collection(collectionName)
	collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "awb", Value: 1}}},
		{Keys: bson.D{{Key: "order_id", Value: 1}}},
		{Keys: bson.D{{Key: "timestamp", Value: 1}}},
	})

	return &trackingRepository{db: db, redis: rdb}
}

func (r *trackingRepository) InsertEvent(ctx context.Context, event model.TrackingEvent) error {
	collection := r.db.Collection(collectionName)
	_, err := collection.InsertOne(ctx, event)
	return err
}

func (r *trackingRepository) GetEventsByAWB(ctx context.Context, awb string) ([]model.TrackingEvent, error) {
	collection := r.db.Collection(collectionName)
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	cursor, err := collection.Find(ctx, bson.M{"awb": awb}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []model.TrackingEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *trackingRepository) SetGPS(ctx context.Context, gps model.CourierGPS) error {
	data, err := json.Marshal(gps)
	if err != nil { return err }

	key := fmt.Sprintf("gps:%s", gps.AWB)
	return r.redis.Set(ctx, key, string(data), gpsTTL).Err()
}

func (r *trackingRepository) GetGPS(ctx context.Context, awb string) (*model.CourierGPS, error) {
	key := fmt.Sprintf("gps:%s", awb)
	data, err := r.redis.Get(ctx, key).Result()
	if err != nil { return nil, err }

	var gps model.CourierGPS
	if err := json.Unmarshal([]byte(data), &gps); err != nil { return nil, err }
	return &gps, nil
}
