package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
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

var (
	mongoDB    *mongo.Database
	collection *mongo.Collection
)

func main() {
	log.Println("🚀 Starting NusaRoute Notification Service...")
	port := getEnv("PORT", "8009")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	var err error
	mongoDB, err = database.ConnectMongo(database.MongoConfig{
		URI:    getEnv("MONGO_URI", "mongodb://nusaroute:nusaroute_secret@localhost:27017"),
		DBName: getEnv("MONGO_DB", "nusaroute_notification"),
	})
	if err != nil { log.Fatalf("MongoDB failed: %v", err) }

	collection = mongoDB.Collection("notification_logs")
	collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cg := kafka.NewConsumerGroup()

	// Subscribe to all notification-worthy events
	cg.Subscribe(ctx, kafkaBrokers, events.TopicCourierAssigned, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.CourierAssignedEvent
		json.Unmarshal(value, &evt)
		return sendNotification(ctx, "", "PUSH", "Kurir Dalam Perjalanan 🛵",
			fmt.Sprintf("Kurir %s sedang menuju lokasi Anda untuk menjemput paket (AWB: %s)", evt.CourierName, evt.AWB),
			evt.OrderID, evt.AWB)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackagePickedUp, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackagePickedUpEvent
		json.Unmarshal(value, &evt)
		return sendNotification(ctx, "", "WHATSAPP", "Paket Dijemput ✅",
			fmt.Sprintf("Paket Anda (AWB: %s) telah dijemput oleh kurir dan sedang dalam perjalanan.", evt.AWB),
			evt.OrderID, evt.AWB)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackageDelivered, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackageDeliveredEvent
		json.Unmarshal(value, &evt)
		return sendNotification(ctx, "", "WHATSAPP", "Paket Terkirim 📦✅",
			fmt.Sprintf("Paket Anda (AWB: %s) telah diterima oleh %s. Terima kasih menggunakan NusaRoute!", evt.AWB, evt.ReceiverName),
			evt.OrderID, evt.AWB)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicDeliveryFailed, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.DeliveryFailedEvent
		json.Unmarshal(value, &evt)
		return sendNotification(ctx, "", "PUSH", "Pengiriman Gagal ⚠️",
			fmt.Sprintf("Pengiriman paket (AWB: %s) gagal: %s. Percobaan ke-%d dari %d.", evt.AWB, evt.Reason, evt.AttemptNum, evt.MaxAttempts),
			evt.OrderID, evt.AWB)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackageLost, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackageLostSuspectedEvent
		json.Unmarshal(value, &evt)
		return sendNotification(ctx, "", "EMAIL", "⚠️ Paket Diduga Hilang",
			fmt.Sprintf("Paket Anda (AWB: %s) tidak menunjukkan aktivitas selama %d jam. Tim kami sedang menginvestigasi.", evt.AWB, evt.HoursSinceUpdate),
			evt.OrderID, evt.AWB)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicResolutionCreated, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.ResolutionCreatedEvent
		json.Unmarshal(value, &evt)
		return sendNotification(ctx, "", "EMAIL", "Tiket Resolusi Dibuat 📋",
			fmt.Sprintf("Tiket #%s telah dibuat untuk paket AWB: %s. Tipe: %s. Tim kami akan segera menindaklanjuti.", evt.TicketID, evt.AWB, evt.Type),
			evt.OrderID, evt.AWB)
	})

	defer cg.CloseAll()

	// HTTP endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/notifications", func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" { userID = r.URL.Query().Get("user_id") }

		opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(50)
		cursor, err := collection.Find(r.Context(), bson.M{"user_id": userID}, opts)
		if err != nil { response.InternalError(w, err.Error()); return }
		var logs []NotificationLog
		cursor.All(r.Context(), &logs)
		response.Success(w, "notifications retrieved", logs)
	})
	mux.HandleFunc("PUT /api/v1/notifications/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-2]
		collection.UpdateOne(r.Context(), bson.M{"_id": id}, bson.M{"$set": bson.M{"is_read": true}})
		response.Success(w, "marked as read", nil)
	})
	mux.HandleFunc("GET /api/v1/notifications/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "notification-service is healthy", map[string]string{"status": "UP"})
	})

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	log.Printf("✅ Notification Service running on port %s", port)
	if err := http.ListenAndServe(":"+port, h); err != nil { log.Fatalf("Server failed: %v", err) }
}

func sendNotification(ctx context.Context, userID, channel, title, message, orderID, awb string) error {
	notif := NotificationLog{
		ID: uuid.New().String(), UserID: userID, Channel: channel,
		Title: title, Message: message, OrderID: orderID, AWB: awb,
		Status: "SENT", IsRead: false, CreatedAt: time.Now(),
	}

	// In production: integrate with Twilio (WhatsApp), Firebase (Push), SMTP (Email)
	// For now, log and save to MongoDB
	log.Printf("[Notification] 📨 [%s] %s: %s", channel, title, message)

	_, err := collection.InsertOne(ctx, notif)
	return err
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
