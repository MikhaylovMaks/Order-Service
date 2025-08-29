package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	topic  string
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
		topic: topic,
	}
}

func (p *Producer) Start(ctx context.Context) {
	log.Println("Kafka producer started...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka producer stopped")
			return
		case <-ticker.C:
			order := models.GenerateFakeOrder()
			data, _ := json.Marshal(order)
			err := p.writer.WriteMessages(ctx,
				kafka.Message{Value: data},
			)
			if err != nil {
				log.Printf("producer error: %v", err)
			} else {
				log.Printf("order sent to kafka: %s", order.OrderUID)
			}
		}
	}
}
