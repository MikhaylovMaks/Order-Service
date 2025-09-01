package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/MikhaylovMaks/wb_techl0/internal/faker"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	topic  string
	log    *zap.SugaredLogger
}

func NewProducer(brokers []string, topic string, log *zap.SugaredLogger) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
		},
		topic: topic,
		log:   log,
	}
}

func (p *Producer) Start(ctx context.Context) {
	log.Println("Kafka producer started...")
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		if err := p.writer.Close(); err != nil {
			p.log.Errorw("failed to close kafka writer", "err", err)
		}
	}()
	p.log.Info("Kafka producer started")

	for {
		select {
		case <-ctx.Done():
			p.log.Info("Kafka producer stopped")
			return
		case <-ticker.C:
			order := faker.GenerateFakeOrder()
			data, err := json.Marshal(order)
			if err != nil {
				p.log.Errorw("failed to marshal order", "err", err)
				continue
			}

			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = p.writer.WriteMessages(writeCtx, kafka.Message{Value: data})
			cancel()

			if err != nil {
				p.log.Errorw("failed to write message", "err", err, "order_uid", order.OrderUID)
				continue
			}

			p.log.Infow("order sent",
				"order_uid", order.OrderUID,
				"topic", p.topic,
			)
		}
	}
}
