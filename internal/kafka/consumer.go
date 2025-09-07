package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/MikhaylovMaks/wb_techl0/internal/repository/postgres"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// структура Kafka-консюмера
type Consumer struct {
	reader *kafka.Reader
	repo   postgres.OrderRepository
	cache  storage.Cache
	log    *zap.SugaredLogger
	v      *validator.Validate
}

// конструктор Kafka Consumer
func NewConsumer(brokers []string, topic, groupID string, repo postgres.OrderRepository, cache storage.Cache, log *zap.SugaredLogger) *Consumer {
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
		log:    log,
		v:      validator.New(),
	}
}

// запускает обработку сообщений из Kafka
func (c *Consumer) Start(ctx context.Context) {
	defer c.reader.Close()
	c.log.Info("Kafka consumer started")
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				c.log.Info("Kafka consumer stopped")
				return
			}
			c.log.Errorw("error fetching message", "err", err)
			continue
		}

		var order models.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			c.log.Warnw("invalid json", "err", err, "raw", string(m.Value))
			_ = c.reader.CommitMessages(ctx, m)
			continue
		} // Валидация структуры заказа
		if err := c.v.Struct(order); err != nil {
			c.log.Warnw("validation failed", "err", err, "order_uid", order.OrderUID)
			_ = c.reader.CommitMessages(ctx, m)
			continue
		}

		if err := c.repo.SaveOrder(ctx, &order); err != nil {
			// simple retry with fixed backoff
			var saved bool
			for attempt := 1; attempt <= 3; attempt++ {
				if err := c.repo.SaveOrder(ctx, &order); err == nil {
					saved = true
					break
				}
				c.log.Warnw("retry save order", "attempt", attempt, "order_uid", order.OrderUID)
				time.Sleep(500 * time.Millisecond)
			}
			if !saved {
				c.log.Errorw("failed to save order after retries", "order_uid", order.OrderUID)
				continue
			}
		}

		c.cache.Set(order.OrderUID, &order)
		c.log.Infow("order saved", "order_uid", order.OrderUID)

		_ = c.reader.CommitMessages(ctx, m)
	}
}
