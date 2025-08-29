package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/MikhaylovMaks/wb_techl0/internal/repository/postgres"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	repo   postgres.OrderRepository
	cache  storage.Cache
}

func NewConsumer(brokers []string, topic, groupID string, repo postgres.OrderRepository, cache storage.Cache) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	return &Consumer{
		reader: r,
		repo:   repo,
		cache:  cache,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	defer c.reader.Close()

	log.Println("Kafka consumer started...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka consumer stopped")
			return
		default:
			m, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("error reading kafka message: %v", err)
				continue
			}

			var order models.Order
			if err := json.Unmarshal(m.Value, &order); err != nil {
				log.Printf("invalid json: %v", err)
				continue
			}

			if err := c.repo.SaveOrder(ctx, &order); err != nil {
				log.Printf("failed to save order: %v", err)
				continue
			}

			c.cache.Set(order.OrderUID, &order)
			log.Printf("order saved from kafka: %s", order.OrderUID)
		}
	}
}
