# Order Service

A demo microservice written in Go for processing and displaying orders.
The service consumes messages from **Kafka**, validates and stores them in **PostgreSQL**, and exposes order data through a **cached** in-memory layer, and exposes them via an **HTTP API** and a simple **web interface**.

---

## Features

- Connects to Kafka (segmentio/kafka-go) and processes messages in real time.
- Stores valid order data in PostgreSQL using transactions.
- In-memory cache with warm-up on startup and invalidation support.
- HTTP API:
  - `GET /orders/{order_uid}` — returns order details as JSON.
- Web interface:
  - Static HTML UI for querying orders by ID.

---

## Tech Stack

Go • Kafka • PostgreSQL • Docker Compose • Gorilla Mux • Zap Logger

---

## Architecture Overview

- Consumer subscribes to a Kafka topic with orders.
- Parser/Validator processes incoming JSON, discarding/logging invalid messages.
- Repository stores the order model in PostgreSQL atomically.
- Cache keeps recent orders in memory (map) and is reloaded from DB on startup.
- HTTP API retrieves orders by order_uid (from cache first, DB fallback).
- Web UI — static page that queries the API.

## Repository Structure

- cmd/service/ — service entrypoint (main).
- config/ — configuration files / environment defaults.
- internal/ — domain logic (consumer, producer, cache, repository, http-handlers, models).
- pkg/ — shared packages (logger, postgres).
- migrations/ — SQL migrations for PostgreSQL.
- web/ — static frontend (HTML).
- compose.yaml — Docker Compose configuration for local infra.
- Dockerfile, .dockerignore — containerization.

## Quick start (recommended)

Using Docker Compose (local dev):

```bash
git clone https://github.com/MikhaylovMaks/Order-Service.git
cd Order-Service
docker compose up -d --build
# HTTP server: http://localhost:8081
# Kafka UI (Kafdrop): http://localhost:9000
# Postgres: localhost:5432
```

- The compose stack uses compose.yaml from the repo root to bring up Kafka, PostgreSQL and the service.
- Service typically binds to :8081. See README.Docker.md for Docker-specific instructions.

## Verification

- Open http://localhost:8081 — static page served from the web/ directory.
- API endpoint: GET http://localhost:8081/orders/{order_uid}.
- A producer publishes test orders into Kafka every 5 seconds (see compose.yaml).

# Stopping the stack

```bash
docker compose down -v
```

## Run locally (without Docker)

````## Run locally (without Docker)

Requirements: Go 1.23+, PostgreSQL, Kafka.

```bash
export CONFIG_PATH=$(pwd)/config/config.yaml
# ensure Postgres & Kafka are running and match config

go run ./cmd/service
````

## Configuration

Default config file: `config/config.yaml` (mounted into the service container).
Environment variables can override config values.

# Environment variables (example from compose.yaml)

- CONFIG_PATH=/config/config.yaml
- POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB
- KAFKA_BROKER, KAFKA_TOPIC, KAFKA_GROUP_ID

# HTTP API

`GET /orders/{order_uid}`

- 200 — JSON with order details
- 404 — order not found
- 400 — invalid request
- 500 — internal server error

## Author

Developed by **Maksim Mikhaylov**
Task assignment: Wildberries
📧 Email: [maksskamm19@bk.ru](mailto:maksskamm19@bk.ru)
💻 GitHub: [MikhaylovMaks](https://github.com/MikhaylovMaks)
