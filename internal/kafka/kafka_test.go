package kafka

import (
	"testing"

	"github.com/MikhaylovMaks/wb_techl0/internal/repository/postgres"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
	"go.uber.org/zap"
)

func TestNewConsumer(t *testing.T) {
	log, _ := zap.NewDevelopment()
	cache := storage.NewMemoryStorage()
	var repo postgres.OrderRepository
	consumer := NewConsumer([]string{"localhost:9092"}, "topic", "group", repo, cache, log.Sugar())
	if consumer == nil || consumer.reader == nil {
		t.Fatal("expected non-nil consumer")
	}
}

func TestNewProducer(t *testing.T) {
	log, _ := zap.NewDevelopment()
	producer := NewProducer([]string{"localhost:9092"}, "topic", log.Sugar())
	if producer == nil || producer.writer == nil {
		t.Fatal("expected non-nil producer")
	}
}
