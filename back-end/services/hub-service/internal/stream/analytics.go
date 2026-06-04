package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/logger"
)

type HubAnalyticsStream struct {
	producer      *kafka.Producer
	brokers       []string
	windowSize    time.Duration
	hubScansCount map[string]int
	mu            sync.Mutex
}

func NewHubAnalyticsStream(brokers []string, producer *kafka.Producer, windowSize time.Duration) *HubAnalyticsStream {
	return &HubAnalyticsStream{
		producer:      producer,
		brokers:       brokers,
		windowSize:    windowSize,
		hubScansCount: make(map[string]int),
	}
}

// Start initiates the Stateful Stream Processing topology
// It reads from package.scanned-at-hub and emits aggregated metrics every window period.
func (s *HubAnalyticsStream) Start(ctx context.Context) {
	logger.Info(context.Background(), fmt.Sprintf(" Starting Stateful Stream Processor (Tumbling Window: %v)", s.windowSize))

	consumer := kafka.NewConsumerGroup()
	defer consumer.CloseAll()

	// 1. Start the Tumbling Window Emitter (Background goroutine)
	go s.runWindowEmitter(ctx)

	// 2. Consume and aggregate events into state (Stateful Processing)
	consumer.Subscribe(ctx, s.brokers, events.TopicPackageScannedHub, "hub-analytics-group",
		func(cCtx context.Context, key, value []byte) error {
			var evt events.PackageScannedAtHubEvent
			if err := json.Unmarshal(value, &evt); err != nil {
				return err
			}

			// Update Stateful In-Memory Aggregation
			s.mu.Lock()
			s.hubScansCount[evt.HubID]++
			s.mu.Unlock()

			return nil
		})

	// Block until context is canceled
	<-ctx.Done()
	logger.Info(context.Background(), "Stopping Stateful Stream Processor")
}

func (s *HubAnalyticsStream) runWindowEmitter(ctx context.Context) {
	ticker := time.NewTicker(s.windowSize)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.emitWindow()
		}
	}
}

func (s *HubAnalyticsStream) emitWindow() {
	s.mu.Lock()
	// Snapshot current state and clear it (Tumbling Window semantics)
	snapshot := s.hubScansCount
	s.hubScansCount = make(map[string]int)
	s.mu.Unlock()

	if len(snapshot) == 0 {
		return // Nothing to emit
	}

	for hubID, count := range snapshot {
		payload := map[string]interface{}{
			"hub_id":          hubID,
			"packages_scanned": count,
			"window_time":     time.Now().Format(time.RFC3339),
		}
		
		val, _ := json.Marshal(payload)
		err := s.producer.Publish(context.Background(), "hub.analytics.realtime", hubID, val)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("Failed to emit stream analytics for Hub %s: %v", hubID, err))
		} else {
			logger.Info(context.Background(), fmt.Sprintf("[Stream Analytics] Emitted window for Hub %s: %d scans", hubID, count))
		}
	}
}
