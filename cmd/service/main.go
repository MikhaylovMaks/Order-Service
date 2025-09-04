package main

import (
	"log"

	"github.com/MikhaylovMaks/wb_techl0/internal/app"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		log.Fatalf("service stopped with error: %v", err)
	}
}
