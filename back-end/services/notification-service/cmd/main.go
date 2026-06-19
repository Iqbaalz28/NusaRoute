package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/nusaroute/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/notification-service/internal/broker"
	"github.com/nusaroute/services/notification-service/internal/orders"
	"github.com/nusaroute/services/notification-service/internal/repository"
	"github.com/nusaroute/services/notification-service/internal/service"
)

func main() {
	logger.InitLogger("notification-service")
	logger.Info(context.Background(), fmt.Sprint(" Starting NusaRoute Notification Service..."))
	port := getEnv("PORT", "8009")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	mongoDB, err := database.ConnectMongo(database.MongoConfig{
		URI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName: getEnv("MONGO_DB", "nusaroute_notification"),
	})
	if err != nil { logger.Log.Fatal(fmt.Sprintf("MongoDB failed: %v", err)) }

	repo := repository.NewNotificationRepository(mongoDB)
	notifBroker := broker.New()
	orderResolver := orders.NewResolver(getEnv("ORDER_SERVICE_URL", "http://localhost:8004"))
	notifSvc := service.NewNotificationService(repo, notifBroker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// send resolves the recipient (order owner) from the AWB, then stores + pushes
	// the notification. Events carry an AWB but not the customer's user id.
	send := func(ctx context.Context, channel, title, message, orderID, awb string) error {
		userID, err := orderResolver.UserIDForAWB(ctx, awb)
		if err != nil {
			logger.Info(context.Background(), fmt.Sprintf("[Notification] could not resolve user for AWB %s: %v", awb, err))
		}
		return notifSvc.SendNotification(ctx, userID, channel, title, message, orderID, awb)
	}

	cg := kafka.NewConsumerGroup()

	cg.Subscribe(ctx, kafkaBrokers, events.TopicCourierAssigned, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.CourierAssignedEvent
		json.Unmarshal(value, &evt)
		return send(ctx, "PUSH", "Kurir Dalam Perjalanan 🛵",
			fmt.Sprintf("Kurir %s sedang menuju lokasi Anda untuk menjemput paket (AWB: %s)", evt.CourierName, evt.AWB),
			evt.OrderID, evt.AWB)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackagePickedUp, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackagePickedUpEvent
		json.Unmarshal(value, &evt)
		return send(ctx, "WHATSAPP", "Paket Dijemput ✅",
			fmt.Sprintf("Paket Anda (AWB: %s) telah dijemput oleh kurir dan sedang dalam perjalanan.", evt.AWB),
			evt.OrderID, evt.AWB)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackageDelivered, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackageDeliveredEvent
		json.Unmarshal(value, &evt)
		return send(ctx, "WHATSAPP", "Paket Terkirim 📦✅",
			fmt.Sprintf("Paket Anda (AWB: %s) telah diterima oleh %s. Terima kasih menggunakan NusaRoute!", evt.AWB, evt.ReceiverName),
			evt.OrderID, evt.AWB)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicDeliveryFailed, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.DeliveryFailedEvent
		json.Unmarshal(value, &evt)
		return send(ctx, "PUSH", "Pengiriman Gagal ⚠️",
			fmt.Sprintf("Pengiriman paket (AWB: %s) gagal: %s. Percobaan ke-%d dari %d.", evt.AWB, evt.Reason, evt.AttemptNum, evt.MaxAttempts),
			evt.OrderID, evt.AWB)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackageLost, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackageLostSuspectedEvent
		json.Unmarshal(value, &evt)
		return send(ctx, "EMAIL", "⚠️ Paket Diduga Hilang",
			fmt.Sprintf("Paket Anda (AWB: %s) tidak menunjukkan aktivitas selama %d jam. Tim kami sedang menginvestigasi.", evt.AWB, evt.HoursSinceUpdate),
			evt.OrderID, evt.AWB)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicResolutionCreated, "notification-service", func(ctx context.Context, key, value []byte) error {
		var evt events.ResolutionCreatedEvent
		json.Unmarshal(value, &evt)
		return send(ctx, "EMAIL", "Tiket Resolusi Dibuat 📋",
			fmt.Sprintf("Tiket #%s telah dibuat untuk paket AWB: %s. Tipe: %s. Tim kami akan segera menindaklanjuti.", evt.TicketID, evt.AWB, evt.Type),
			evt.OrderID, evt.AWB)
	})

	defer cg.CloseAll()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("GET /api/v1/notifications", func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" { userID = r.URL.Query().Get("user_id") }

		logs, err := notifSvc.GetNotifications(r.Context(), userID)
		if err != nil { response.InternalError(w, err.Error()); return }
		response.Success(w, "notifications retrieved", logs)
	})
	// Real-time stream — pushes the user's new notifications over SSE. Token is in
	// the query (EventSource can't set headers); the gateway validates it and
	// forwards X-User-ID, which HeaderAuth places in the context.
	mux.HandleFunc("GET /api/v1/notifications/stream", func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			response.BadRequest(w, "unauthenticated")
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := notifBroker.Subscribe(userID)
		defer notifBroker.Unsubscribe(userID, ch)
		fmt.Fprint(w, ": connected\n\n")
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			}
		}
	})
	mux.HandleFunc("PUT /api/v1/notifications/read-all", func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			response.BadRequest(w, "unauthenticated")
			return
		}
		if err := notifSvc.MarkAllAsRead(r.Context(), userID); err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "all marked as read", nil)
	})
	mux.HandleFunc("PUT /api/v1/notifications/{id}/read", func(w http.ResponseWriter, r *http.Request) {
		if err := notifSvc.MarkAsRead(r.Context(), r.PathValue("id")); err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "marked as read", nil)
	})
	mux.HandleFunc("GET /api/v1/notifications/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "notification-service is healthy", map[string]string{"status": "UP"})
	})

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.HeaderAuth(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	h = middleware.Metrics(h)

	logger.Info(context.Background(), fmt.Sprintf("✅ Notification Service running on port %s", port))
	if err := http.ListenAndServe(":"+port, h); err != nil { logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err)) }
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
