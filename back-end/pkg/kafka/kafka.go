// Package kafka provides a thin wrapper around segmentio/kafka-go
// for producing and consuming events in NusaRoute's Event-Driven Architecture.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

// ensureTopic creates the topic if it does not already exist (idempotent).
// Without this, a consumer that starts before any producer has created the
// topic attaches to nothing and silently never receives messages.
func ensureTopic(brokers []string, topic string) {
	if len(brokers) == 0 || topic == "" {
		return
	}
	conn, err := kafkago.Dial("tcp", brokers[0])
	if err != nil {
		log.Printf("[Kafka] ensureTopic dial failed for %s: %v", topic, err)
		return
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Printf("[Kafka] ensureTopic controller lookup failed for %s: %v", topic, err)
		return
	}
	ctrlConn, err := kafkago.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Printf("[Kafka] ensureTopic controller dial failed for %s: %v", topic, err)
		return
	}
	defer ctrlConn.Close()

	if err := ctrlConn.CreateTopics(kafkago.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}); err != nil {
		log.Printf("[Kafka] ensureTopic create failed for %s: %v", topic, err)
		return
	}
	log.Printf("[Kafka] ensureTopic OK: %s", topic)
}

// DLQTopicName is the Dead Letter Queue topic for failed events.
const DLQTopicName = "nusaroute.dlq"

// Producer wraps kafka-go Writer for publishing events.
type Producer struct {
	writer *kafkago.Writer
}

// NewProducer creates a new Kafka producer.
func NewProducer(brokers []string) *Producer {
	w := &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Balancer:     &kafkago.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafkago.RequireOne,
	}
	return &Producer{writer: w}
}

// Publish serializes the event to JSON and sends it to the specified topic.
func (p *Producer) Publish(ctx context.Context, topic string, key string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafkago.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
		Time:  time.Now(),
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("[Kafka Producer] Failed to publish to %s: %v", topic, err)
		return err
	}

	log.Printf("[Kafka Producer] Published event to topic=%s key=%s", topic, key)
	return nil
}

// Close closes the producer.
func (p *Producer) Close() error {
	return p.writer.Close()
}

// DLQEvent wraps a failed message with error context for the Dead Letter Queue.
type DLQEvent struct {
	OriginalTopic string `json:"original_topic"`
	OriginalKey   string `json:"original_key"`
	Payload       string `json:"payload"`
	ErrorReason   string `json:"error_reason"`
	FailedAt      string `json:"failed_at"`
}

// Consumer wraps kafka-go Reader for consuming events.
type Consumer struct {
	reader    *kafkago.Reader
	dlqWriter *kafkago.Writer // optional: nil disables DLQ publishing
}

// NewConsumer creates a new Kafka consumer for a specific topic and group.
func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	// Make sure the topic exists so the reader actually attaches to it even when
	// this consumer starts before any producer has published to the topic.
	ensureTopic(brokers, topic)

	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafkago.FirstOffset,
	})

	// DLQ writer — always initialized so failed events are never silently dropped.
	dlq := &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Topic:        DLQTopicName,
		Balancer:     &kafkago.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafkago.RequireOne,
	}
	return &Consumer{reader: r, dlqWriter: dlq}
}

// publishToDLQ sends a failed message to the Dead Letter Queue topic.
func (c *Consumer) publishToDLQ(ctx context.Context, msg kafkago.Message, handlerErr error) {
	if c.dlqWriter == nil {
		log.Printf("[Kafka DLQ] No DLQ writer configured, dropping failed event from topic=%s", msg.Topic)
		return
	}

	evt := DLQEvent{
		OriginalTopic: msg.Topic,
		OriginalKey:   string(msg.Key),
		Payload:       string(msg.Value),
		ErrorReason:   handlerErr.Error(),
		FailedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("[Kafka DLQ] Failed to marshal DLQ event: %v", err)
		return
	}

	dlqMsg := kafkago.Message{
		Key:   []byte(fmt.Sprintf("dlq-%s-%s", msg.Topic, string(msg.Key))),
		Value: data,
		Time:  time.Now(),
	}

	if err := c.dlqWriter.WriteMessages(ctx, dlqMsg); err != nil {
		log.Printf("[Kafka DLQ] Failed to publish to DLQ from topic=%s: %v", msg.Topic, err)
	} else {
		log.Printf("[Kafka DLQ]  Sent failed event to DLQ | original_topic=%s key=%s reason=%s",
			msg.Topic, string(msg.Key), handlerErr.Error())
	}
}

// MessageHandler is a function that processes a consumed Kafka message.
type MessageHandler func(ctx context.Context, key []byte, value []byte) error

// Consume starts consuming messages and calls the handler for each message.
// It blocks until the context is cancelled.
func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("[Kafka Consumer] Context cancelled, stopping consumer for topic=%s", c.reader.Config().Topic)
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[Kafka Consumer] Error fetching message: %v", err)
				continue
			}

			log.Printf("[Kafka Consumer] Received message from topic=%s partition=%d offset=%d",
				msg.Topic, msg.Partition, msg.Offset)

			var handlerErr error
			maxRetries := 3
			backoff := 100 * time.Millisecond

			for i := 0; i <= maxRetries; i++ {
				handlerErr = handler(ctx, msg.Key, msg.Value)
				if handlerErr == nil {
					break // success
				}
				if i < maxRetries {
					log.Printf("[Kafka Consumer] Error processing msg (attempt %d/%d), retrying in %v: %v", i+1, maxRetries, backoff, handlerErr)
					time.Sleep(backoff)
					backoff *= 2
				}
			}

			if handlerErr != nil {
				log.Printf("[Kafka Consumer] Max retries reached for topic=%s: %v", msg.Topic, handlerErr)
				// Send failed event to Dead Letter Queue for later inspection/replay.
				go c.publishToDLQ(ctx, msg, handlerErr)
			}
			
			// Commit offset ONLY after processing succeeds (or max retries reached and sent to DLQ)
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("[Kafka Consumer] Failed to commit message topic=%s offset=%d: %v", msg.Topic, msg.Offset, err)
			}
		}
	}
}

// Close closes the consumer and the DLQ writer.
func (c *Consumer) Close() error {
	if c.dlqWriter != nil {
		if err := c.dlqWriter.Close(); err != nil {
			log.Printf("[Kafka DLQ] Error closing DLQ writer: %v", err)
		}
	}
	return c.reader.Close()
}

// ConsumerGroup manages multiple consumers for different topics.
type ConsumerGroup struct {
	consumers []*Consumer
}

// NewConsumerGroup creates a new consumer group manager.
func NewConsumerGroup() *ConsumerGroup {
	return &ConsumerGroup{
		consumers: make([]*Consumer, 0),
	}
}

// Subscribe creates a consumer for a topic and starts consuming.
func (cg *ConsumerGroup) Subscribe(ctx context.Context, brokers []string, topic string, groupID string, handler MessageHandler) {
	consumer := NewConsumer(brokers, topic, groupID)
	cg.consumers = append(cg.consumers, consumer)
	go consumer.Consume(ctx, handler)
}

// CloseAll closes all consumers in the group.
func (cg *ConsumerGroup) CloseAll() {
	for _, c := range cg.consumers {
		if err := c.Close(); err != nil {
			log.Printf("[Kafka ConsumerGroup] Error closing consumer: %v", err)
		}
	}
}
