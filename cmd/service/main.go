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
	"github.com/MikhaylovMaks/wb_techl0/pkg/database"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	sugar.Infow("starting server")
	// db
	ctx := context.Background()
	db, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	sugar.Infow("connected to postgres", "host", cfg.Postgres.Host, "db", cfg.Postgres.DBName)
	defer db.Pool.Close()
	store := storage.NewMemoryStorage()

	repo := postgres.NewRepository(db.Pool)

	orderFromDB, err := repo.GetOrderByUID(ctx, "1")
	if err != nil {
		log.Fatalf("failed to get order: %v", err)
	}
	fmt.Printf("Loaded order: %+v\n", orderFromDB)

	h := handlers.NewHandler(store)
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/health", h.HealthCheck)
	http.HandleFunc("/order/", h.GetOrder)
	http.HandleFunc("/order", h.CreateOrder)

	addr := ":8081"
	var greeting string
	err = db.Pool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v", err)
	}
	log.Println("DB says:", greeting)
	fmt.Println("Server started at", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("server failed:", err)
	}
	fmt.Printf("Server is ready on port: %d\n", cfg.Server.Port)
}
