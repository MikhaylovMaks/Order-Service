.PHONY: build run clean migrate-up migrate-down up down logs rebuild test cover
CONFIG_PATH := ./config/config.yaml

PORT := 8081

BIN := service

run:
	CONFIG_PATH=$(CONFIG_PATH) go run ./cmd/service

build:
	go build -o $(BIN) ./cmd/service

clean:
	rm -f $(BIN)

migrate-up:
	docker run --rm \
	-v $(PWD)/migrations:/migrations \
	--network host \
	migrate/migrate \
	-path=/migrations/ -database postgres://orders_user:maksim19@localhost:5432/orders_db?sslmode=disable up

migrate-down:
	docker run --rm \
	-v $(PWD)/migrations:/migrations \
	--network host \
	migrate/migrate \
	-path=/migrations/ -database postgres://orders_user:maksim19@localhost:5432/orders_db?sslmode=disable down

# Docker compose helpers
up:
	docker compose up -d --build

down:
	docker compose down -v

logs:
	docker compose logs -f service

rebuild:
	docker compose build --no-cache service && docker compose up -d service

# Tests
test:
	go test -v ./...

cover:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
