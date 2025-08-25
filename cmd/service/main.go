package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/MikhaylovMaks/wb_techl0/internal/config"
	"github.com/MikhaylovMaks/wb_techl0/internal/handlers"
	"github.com/MikhaylovMaks/wb_techl0/internal/repository/postgres"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
	"go.uber.org/zap"
)

func main() {
	//config
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	sugar.Infow("starting server")
	// db
	ctx := context.Background()
	db, err := postgres.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	sugar.Infow("connected to postgres", "host", cfg.Postgres.Host, "db", cfg.Postgres.DBName)
	defer db.Pool.Close()

	store := storage.NewMemoryStorage()

	h := handlers.NewHandler(store)

	http.HandleFunc("/health", h.HealthCheck)
	http.HandleFunc("/order/", h.GetOrder)   // GET /order/{id}
	http.HandleFunc("/order", h.CreateOrder) // POST /order

	addr := ":8081"
	fmt.Println("Server started at", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("server failed:", err)
	}
	fmt.Printf("Server is ready on port: %d\n", cfg.Server.Port)
	var greeting string
	err = db.Pool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v", err)
	}
	log.Println("DB says:", greeting)
}
