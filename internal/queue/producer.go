package queue

import (
	"context"
	"fmt"
	"os"
	"portfolio-rebalancer/internal/logging"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct{}

func NewKafkaPublisher() *KafkaPublisher {
	return &KafkaPublisher{}
}

var writer *kafka.Writer

func InitKafka() error {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	topic := os.Getenv("KAFKA_TOPIC")

	if kafkaBroker == "" || topic == "" {
		return nil
	}

	writer = &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	for i := 0; i < 10; i++ {
		err := writer.WriteMessages(context.Background(), kafka.Message{
			Value: []byte("ping"),
		})
		if err == nil {
			logging.Infof("Kafka is ready")
			return nil
		}
		logging.Infof("waiting for Kafka to be ready")
		time.Sleep(2 * time.Second)
	}

	return nil
}

func (k *KafkaPublisher) PublishMessage(ctx context.Context, payload []byte) error {
	if writer == nil {
		logging.Warnf("Kafka writer is nil; skipping message publish")
		return fmt.Errorf("kafka writer not initialized")
	}

	msg := kafka.Message{
		Value: payload,
	}

	return writer.WriteMessages(ctx, msg)
}

func (k *KafkaPublisher) ConsumeMessage(ctx context.Context, handler func(kafka.Message)) error {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	topic := os.Getenv("KAFKA_TOPIC")

	if kafkaBroker == "" || topic == "" {
		logging.Warnf("Kafka consumer config not set; skipping consumer start")
		return nil
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaBroker},
		Topic:     topic,
		Partition: 0,
		MinBytes:  10e3,
		MaxBytes:  10e6,
	})

	reader.SetOffset(kafka.FirstOffset)

	go func() {
		defer reader.Close()
		for {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				logging.Errorf("Kafka read error: %v", err)
				continue
			}
			handler(msg)
		}
	}()

	logging.Infof("Kafka consumer started")
	return nil
}
