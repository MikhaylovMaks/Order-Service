package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MikhaylovMaks/wb_techl0/internal/config"
	"github.com/MikhaylovMaks/wb_techl0/internal/handlers"
	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/MikhaylovMaks/wb_techl0/internal/repository/postgres"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
	"github.com/MikhaylovMaks/wb_techl0/pkg/database"
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
	db, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	sugar.Infow("connected to postgres", "host", cfg.Postgres.Host, "db", cfg.Postgres.DBName)
	defer db.Pool.Close()
	store := storage.NewMemoryStorage()

	// 2. Создаём репозиторий для работы с заказами
	repo := postgres.NewRepository(db.Pool)

	// 3. Создаём тестовый заказ
	testOrder := &models.Order{
		OrderUID:        "765",
		TrackNumber:     "TR12345",
		Entry:           "test_entry",
		Locale:          "ru",
		CustomerID:      "cust_1",
		DeliveryService: "DPD",
		ShardKey:        "shard_1",
		SmID:            1,
		DateCreated:     time.Now(),
		OofShard:        "shard_1",
		Delivery: models.Delivery{
			Name:    "John Doe",
			Phone:   "+71234567890",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Lenina 1",
			Region:  "Moscow",
			Email:   "john@example.com",
		},
		Payment: models.Payment{
			Transaction:  "trx_123",
			Currency:     "RUB",
			Provider:     "Sberbank",
			Amount:       1000,
			PaymentDT:    time.Now().Unix(),
			Bank:         "Sberbank",
			DeliveryCost: 50,
			GoodsTotal:   950,
			CustomFee:    0,
		},
		Items: []models.Items{
			{
				ChrtID: 1, TrackNumber: "TR12345", Price: 300, RID: "RID1", Name: "Product1",
				Sale: 0, Size: "M", TotalPrice: 300, NmID: 1, Brand: "BrandX", Status: 1,
			},
			{
				ChrtID: 2, TrackNumber: "TR12345", Price: 400, RID: "RID2", Name: "Product2",
				Sale: 0, Size: "L", TotalPrice: 400, NmID: 2, Brand: "BrandY", Status: 1,
			},
			{
				ChrtID: 3, TrackNumber: "TR12345", Price: 250, RID: "RID3", Name: "Product3",
				Sale: 0, Size: "S", TotalPrice: 250, NmID: 3, Brand: "BrandZ", Status: 1,
			},
		},
	}

	// 4. Сохраняем заказ в базу
	if err := repo.SaveOrder(ctx, testOrder); err != nil {
		log.Fatalf("failed to save order: %v", err)
	}

	// 5. Можно проверить получение заказа
	orderFromDB, err := repo.GetOrderByUID(ctx, "765")
	if err != nil {
		log.Fatalf("failed to get order: %v", err)
	}
	fmt.Printf("Loaded order: %+v\n", orderFromDB)

	h := handlers.NewHandler(store)
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/health", h.HealthCheck)
	http.HandleFunc("/order/", h.GetOrder)   // GET /order/{id}
	http.HandleFunc("/order", h.CreateOrder) // POST /order

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
