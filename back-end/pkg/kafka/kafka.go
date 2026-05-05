// Package kafka provides a thin wrapper around segmentio/kafka-go
// for producing and consuming events in NusaRoute's Event-Driven Architecture.
package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

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

// Consumer wraps kafka-go Reader for consuming events.
type Consumer struct {
	reader *kafkago.Reader
}

// NewConsumer creates a new Kafka consumer for a specific topic and group.
func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafkago.FirstOffset,
	})
	return &Consumer{reader: r}
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
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[Kafka Consumer] Error reading message: %v", err)
				continue
			}

			log.Printf("[Kafka Consumer] Received message from topic=%s partition=%d offset=%d",
				msg.Topic, msg.Partition, msg.Offset)

			if err := handler(ctx, msg.Key, msg.Value); err != nil {
				log.Printf("[Kafka Consumer] Error processing message from topic=%s: %v", msg.Topic, err)
				// In production, we would send this to a DLQ
			}
		}
	}
}

// Close closes the consumer.
func (c *Consumer) Close() error {
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
