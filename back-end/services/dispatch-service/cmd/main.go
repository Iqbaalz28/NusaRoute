package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nusaroute/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/outbox"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/dispatch-service/internal/broker"
	"github.com/nusaroute/services/dispatch-service/internal/model"
	"github.com/nusaroute/services/courier-service/pkg/grpc/pb"
	"github.com/nusaroute/services/dispatch-service/internal/repository"
	"github.com/nusaroute/services/dispatch-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger.InitLogger("dispatch-service")
	logger.Info(context.Background(), fmt.Sprint(" Starting NusaRoute Dispatch Service..."))
	port := getEnv("PORT", "8006")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", ""),
		DBName: getEnv("DB_NAME", "nusaroute_dispatch"),
	})
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("DB failed: %v", err))
	}
	defer db.Close()

	producer := kafka.NewProducer(kafkaBrokers)
	defer producer.Close()

	// Connect to Courier gRPC Service
	grpcTarget := getEnv("COURIER_GRPC_TARGET", "localhost:9005")
	logger.Info(context.Background(), fmt.Sprintf("Connecting to Courier gRPC at %s", grpcTarget))
	grpcConn, err := grpc.Dial(grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("Failed to connect to courier gRPC: %v", err))
	}
	defer grpcConn.Close()
	courierClient := pb.NewCourierServiceClient(grpcConn)

	redisClient, err := database.ConnectRedis(database.RedisConfig{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
	})
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("Redis unavailable, distributed lock disabled: %v", err))
	}

	jobBroker := broker.New()
	dispatchRepo := repository.NewDispatchRepository(db)
	dispatchSvc := service.NewDispatchService(dispatchRepo, producer, courierClient, redisClient, jobBroker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// Outbox Worker — drains assignment/courier lifecycle events to Kafka
	outbox.NewWorker(db, producer, 2*time.Second).Start(ctx)

	// Background: monitor courier no-shows
	wg.Add(1)
	go func() {
		defer wg.Done()
		dispatchSvc.RunNoShowMonitor(ctx)
	}()

	// Kafka: subscribe to order.ready-for-pickup → auto-assign courier
	consumerGroup := kafka.NewConsumerGroup()
	consumerGroup.Subscribe(ctx, kafkaBrokers, events.TopicOrderReadyForPickup, "dispatch-service",
		func(ctx context.Context, key, value []byte) error {
			var evt events.OrderReadyForPickupEvent
			if err := json.Unmarshal(value, &evt); err != nil {
				return err
			}
			logger.Info(context.Background(), fmt.Sprintf("[Dispatch] Order ready for pickup: %s", evt.OrderID))
			// Take-job model: open the first leg (first-mile or direct) for couriers.
			return dispatchSvc.CreateJobFromReady(ctx, evt)
		})

	// Kafka: order arrived at destination hub → open the last-mile job
	consumerGroup.Subscribe(ctx, kafkaBrokers, events.TopicOrderReadyForLastMile, "dispatch-service",
		func(ctx context.Context, key, value []byte) error {
			var evt events.OrderReadyForLastMileEvent
			if err := json.Unmarshal(value, &evt); err != nil {
				return err
			}
			logger.Info(context.Background(), fmt.Sprintf("[Dispatch] Order ready for last-mile: %s", evt.OrderID))
			return dispatchSvc.CreateLastMileJob(ctx, evt)
		})

	// Kafka: hub operator scanned a package IN → close the courier's first-mile leg.
	// The receiving party (the hub) performs the handover scan, not the courier.
	consumerGroup.Subscribe(ctx, kafkaBrokers, events.TopicPackageScannedHub, "dispatch-service",
		func(ctx context.Context, key, value []byte) error {
			var evt events.PackageScannedAtHubEvent
			if err := json.Unmarshal(value, &evt); err != nil {
				return err
			}
			if evt.ScanType != events.ScanTypeArrived {
				return nil // only an inbound (ARRIVED) scan is a handover into the hub
			}
			return dispatchSvc.HandleHubArrival(ctx, evt.AWB, evt.HubName)
		})
	defer consumerGroup.CloseAll()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("GET /api/v1/dispatch/assignments", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		assignments, err := dispatchSvc.ListAssignments(r.Context(), status, 1, 50)
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "assignments retrieved", assignments)
	})
	// Job board — OPEN jobs available for couriers to claim.
	mux.HandleFunc("GET /api/v1/dispatch/jobs", func(w http.ResponseWriter, r *http.Request) {
		jobs, err := dispatchSvc.ListOpenJobs(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "open jobs retrieved", jobs)
	})
	// SSE stream — pushes new OPEN jobs to connected couriers in real time.
	mux.HandleFunc("GET /api/v1/dispatch/jobs/stream", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := jobBroker.Subscribe()
		defer jobBroker.Unsubscribe(ch)
		fmt.Fprint(w, ": connected\n\n") // open the stream
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
	// My tasks — assignments claimed by the authenticated courier.
	mux.HandleFunc("GET /api/v1/dispatch/my", func(w http.ResponseWriter, r *http.Request) {
		courierID := middleware.GetUserID(r.Context())
		if courierID == "" {
			response.BadRequest(w, "unauthenticated")
			return
		}
		list, err := dispatchSvc.ListMyAssignments(r.Context(), courierID)
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "my assignments retrieved", list)
	})
	// Claim an OPEN job (first-come-first-served). Courier identity from JWT.
	mux.HandleFunc("POST /api/v1/dispatch/claim", func(w http.ResponseWriter, r *http.Request) {
		var req model.ClaimRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid body")
			return
		}
		if req.OrderID == "" {
			response.BadRequest(w, "order_id is required")
			return
		}
		courierID := middleware.GetUserID(r.Context())
		if courierID == "" {
			response.BadRequest(w, "unauthenticated")
			return
		}
		a, err := dispatchSvc.ClaimJob(r.Context(), req, courierID)
		if errors.Is(err, service.ErrJobTaken) {
			http.Error(w, `{"success":false,"error":"job sudah diambil kurir lain"}`, http.StatusConflict)
			return
		}
		if err != nil {
			response.BadRequest(w, err.Error())
			return
		}
		response.Success(w, "job claimed", a)
	})
	mux.HandleFunc("POST /api/v1/dispatch/pickup", func(w http.ResponseWriter, r *http.Request) {
		var req model.PickupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid body")
			return
		}
		if req.OrderID == "" {
			response.BadRequest(w, "order_id is required")
			return
		}
		if err := dispatchSvc.PickupPackage(r.Context(), req); err != nil {
			response.BadRequest(w, err.Error())
			return
		}
		response.Success(w, "package picked up", map[string]string{"order_id": req.OrderID, "status": "PICKED_UP"})
	})
	mux.HandleFunc("POST /api/v1/dispatch/deliver", func(w http.ResponseWriter, r *http.Request) {
		var req model.DeliverRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid body")
			return
		}
		if req.OrderID == "" {
			response.BadRequest(w, "order_id is required")
			return
		}
		if err := dispatchSvc.DeliverPackage(r.Context(), req); err != nil {
			response.BadRequest(w, err.Error())
			return
		}
		response.Success(w, "package delivered", map[string]string{"order_id": req.OrderID, "status": "DELIVERED"})
	})
	mux.HandleFunc("GET /api/v1/dispatch/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "dispatch-service is healthy", map[string]string{"status": "UP"})
	})

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.HeaderAuth(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	h = middleware.Metrics(h)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info(context.Background(), fmt.Sprint("Shutting down dispatch-service gracefully..."))
		cancel() // notify all background workers to stop

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Info(context.Background(), fmt.Sprintf("Server shutdown error: %v", err))
		}
	}()

	logger.Info(context.Background(), fmt.Sprintf(" Dispatch Service running on port %s", port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err))
	}

	wg.Wait()
	logger.Info(context.Background(), fmt.Sprint("All dispatch workers stopped. Goodbye!"))
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
