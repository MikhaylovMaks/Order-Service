package kafka

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/MikhaylovMaks/wb_techl0/internal/repository/postgres"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	reader *kafka.Reader
	repo   postgres.OrderRepository
	cache  storage.Cache
	log    *zap.SugaredLogger
}

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
	}
}

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
		if err := json.Unmarshal(m.Value, &order); err != nil || order.OrderUID == "" {
			c.log.Warnw("invalid message", "err", err, "raw", string(m.Value))
			_ = c.reader.CommitMessages(ctx, m)
			continue
		}

		if err := c.repo.SaveOrder(ctx, &order); err != nil {
			c.log.Errorw("failed to save order", "err", err, "order_uid", order.OrderUID)
			continue
		}

		c.cache.Set(order.OrderUID, &order)
		c.log.Infow("order saved", "order_uid", order.OrderUID)

		_ = c.reader.CommitMessages(ctx, m)
	}
}
