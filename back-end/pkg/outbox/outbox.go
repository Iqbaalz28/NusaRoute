package outbox

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/pkg/kafka"
)

// OutboxEvent represents a single event stored in the outbox table.
type OutboxEvent struct {
	ID          string     `db:"id"`
	Topic       string     `db:"topic"`
	Payload     []byte     `db:"payload"`
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	ProcessedAt *time.Time `db:"processed_at"`
}

// InsertEvent inserts a new event into the outbox table within a transaction.
func InsertEvent(ctx context.Context, tx *sqlx.Tx, topic string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	// Generate UUID (import "github.com/google/uuid" inside the file if needed, wait I should use simple crypto/rand or just add the import)
	// I'll just use a simple string generator or require id.
	// Let's use simple crypto/rand or sqlx placeholder.
	// Actually we can let Postgres generate UUID by using GEN_RANDOM_UUID() if supported, but let's just generate it.
	// Wait, I didn't import github.com/google/uuid in this file. I will just do a simple query returning UUID.
	
	query := `
		INSERT INTO outbox_events (id, topic, payload, status)
		VALUES (gen_random_uuid(), $1, $2, 'PENDING')
	`
	_, err = tx.ExecContext(ctx, query, topic, payloadBytes)
	return err
}

// Worker polls the outbox_events table and publishes to Kafka.
type Worker struct {
	db       *sqlx.DB
	producer *kafka.Producer
	interval time.Duration
}

func NewWorker(db *sqlx.DB, producer *kafka.Producer, interval time.Duration) *Worker {
	if interval == 0 {
		interval = 2 * time.Second
	}
	return &Worker{
		db:       db,
		producer: producer,
		interval: interval,
	}
}

// Start runs the worker in a background goroutine until the context is canceled.
func (w *Worker) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[Outbox] Worker stopping...")
				return
			case <-ticker.C:
				w.processEvents(ctx)
			}
		}
	}()
}

func (w *Worker) processEvents(ctx context.Context) {
	if w.producer == nil {
		return // Cannot publish if producer is nil (e.g. in tests)
	}

	tx, err := w.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("[Outbox] Failed to begin tx: %v", err)
		return
	}
	defer tx.Rollback() // Ignored if committed

	// Fetch up to 50 pending events
	var events []OutboxEvent
	err = tx.SelectContext(ctx, &events, `
		SELECT id, topic, payload, status, created_at, processed_at
		FROM outbox_events
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT 50
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		log.Printf("[Outbox] Error fetching events: %v", err)
		return
	}

	for _, evt := range events {
		// Try to publish
		var payloadMap map[string]interface{}
		if err := json.Unmarshal(evt.Payload, &payloadMap); err != nil {
			log.Printf("[Outbox] Error unmarshaling event %s: %v", evt.ID, err)
			tx.ExecContext(ctx, "UPDATE outbox_events SET status = 'FAILED', processed_at = $1 WHERE id = $2", time.Now(), evt.ID)
			continue
		}

		err := w.producer.Publish(ctx, evt.Topic, evt.ID, payloadMap)
		if err != nil {
			log.Printf("[Outbox] Failed to publish event %s to topic %s: %v", evt.ID, evt.Topic, err)
			// Will retry on next tick
			continue
		}

		// Mark as PROCESSED
		tx.ExecContext(ctx, "UPDATE outbox_events SET status = 'PROCESSED', processed_at = $1 WHERE id = $2", time.Now(), evt.ID)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[Outbox] Failed to commit tx: %v", err)
	}
}
