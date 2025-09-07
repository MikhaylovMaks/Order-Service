package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MikhaylovMaks/wb_techl0/internal/config"
	"github.com/MikhaylovMaks/wb_techl0/internal/handlers"
	"github.com/MikhaylovMaks/wb_techl0/internal/kafka"
	"github.com/MikhaylovMaks/wb_techl0/internal/repository/postgres"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
	"github.com/MikhaylovMaks/wb_techl0/pkg/database"
	"github.com/MikhaylovMaks/wb_techl0/pkg/logger"
	"go.uber.org/zap"
)

type App struct {
	cfg *config.Config
}

func New() *App {
	return &App{}
}

func (a *App) Run() error {
	// инициализация логгера
	log, err := logger.NewLogger()
	if err != nil {
		return err
	}
	defer log.Sync()
	log.Info("service starting...")

	// создаём контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// config
	cfg, err := config.NewConfig()
	if err != nil {
		return err
	}
	a.cfg = cfg

	// db
	db, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer db.Close()
	repo := postgres.NewRepository(db.Pool)

	// cache
	cache := storage.NewMemoryStorage()

	// прогреваем кэш
	if err := warmUpCache(ctx, log, repo, cache); err != nil {
		log.Fatalf("cache warm-up failed: %v", err)
	}

	// kafka
	consumer := kafka.NewConsumer(
		[]string{cfg.Kafka.Broker},
		cfg.Kafka.Topic,
		cfg.Kafka.GroupID,
		repo,
		cache,
		log,
	)
	producer := kafka.NewProducer([]string{cfg.Kafka.Broker}, cfg.Kafka.Topic, log)

	// http server
	server := handlers.NewServer(cfg.Server.Port, repo, cache, log)

	var wg sync.WaitGroup
	// запускаем consumer в отдельной горутин
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.Start(ctx)
	}()
	// запускаем producer в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		producer.Start(ctx)
	}()
	// запускаем HTTP-сервер в отдельной горутине
	wg.Add(1)
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)

	wg.Wait()
	log.Info("service stopped")
	return nil
}

// warmUpCache — предварительно загружает заказы из БД в кэш
func warmUpCache(ctx context.Context, log *zap.SugaredLogger, repo postgres.OrderRepository, cache storage.Cache) error {
	warmCtx, warmCancel := context.WithTimeout(ctx, 10*time.Second)
	uids, err := repo.GetAllOrderUIDs(warmCtx)
	warmCancel()
	if err != nil {
		log.Errorw("failed to warm cache: get uids", "err", err)
		return err
	}
	for _, uid := range uids {
		orderCtx, orderCancel := context.WithTimeout(ctx, 5*time.Second)
		order, err := repo.GetOrderByUID(orderCtx, uid)
		orderCancel()
		if err != nil {
			log.Errorw("failed to warm cache: get order", "order_uid", uid, "err", err)
			return err
		}
		cache.Set(uid, order)
	}
	log.Infow("cache warm-up completed", "count", len(uids))
	return nil
}
