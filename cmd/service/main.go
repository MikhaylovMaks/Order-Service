package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/MikhaylovMaks/wb_techl0/internal/config"
	"github.com/MikhaylovMaks/wb_techl0/internal/handlers"
	"github.com/MikhaylovMaks/wb_techl0/internal/kafka"
	"github.com/MikhaylovMaks/wb_techl0/internal/repository/postgres"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
	"github.com/MikhaylovMaks/wb_techl0/pkg/database"
	"github.com/MikhaylovMaks/wb_techl0/pkg/logger"
)

func main() {
	log, err := logger.NewLogger()
	if err != nil {
		fmt.Printf("failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()
	log.Info("service starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := postgres.NewRepository(db.Pool)

	cache := storage.NewMemoryStorage()

	consumer := kafka.NewConsumer(
		[]string{cfg.Kafka.Broker},
		cfg.Kafka.Topic,
		cfg.Kafka.GroupID,
		repo,
		cache,
		log,
	)

	producer := kafka.NewProducer([]string{cfg.Kafka.Broker}, cfg.Kafka.Topic, log)
	var wg sync.WaitGroup
	wg.Add(3)
	server := handlers.NewServer(cfg.Server.Port, repo, cache, log)
	http.Handle("/", http.FileServer(http.Dir("./web")))
	go func() {
		defer wg.Done()
		consumer.Start(ctx)
	}()

	go func() {
		defer wg.Done()
		producer.Start(ctx)
	}()

	go func() {
		defer wg.Done()
		if err := server.Start(); err != nil {
			log.Errorw("http server stopped", "err", err)
			cancel()
		}
	}()
	<-sigs
	log.Info("shutdown signal received")
	cancel()

	wg.Wait()
	log.Info("service stopped")
}
