package queue

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type MessagePublisher interface {
	PublishMessage(ctx context.Context, payload []byte) error
}

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
			log.Println("Kafka is ready")
			return nil
		}
		log.Println("Waiting for Kafka to be ready...")
		time.Sleep(2 * time.Second)
	}

	return nil
}

func (k *KafkaPublisher) PublishMessage(ctx context.Context, payload []byte) error {
	if writer == nil {
		log.Println("Kafka writer is nil; skipping message publish")
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
		log.Println("Kafka consumer config not set; skipping consumer start.")
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
				log.Printf("Kafka read error: %v\n", err)
				continue
			}
			handler(msg)
		}
	}()

	log.Println("Kafka consumer started")
	return nil
}
