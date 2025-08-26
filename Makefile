CONFIG_PATH := ./config/config.yaml

PORT := 8081

BIN := service

run:
	CONFIG_PATH=$(CONFIG_PATH) go run ./cmd/service

build:
	CONFIG_PATH=$(CONFIG_PATH) go build -o $(BIN) ./cmd/service

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
