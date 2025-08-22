package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MikhaylovMaks/wb_techl0/internal/handlers"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
)

func main() {

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
}
