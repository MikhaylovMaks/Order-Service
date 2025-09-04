# Order Service

A demo microservice written in Go for processing and displaying orders.
The service consumes messages from **Kafka**, validates and stores them in **PostgreSQL**, **caches** the latest orders in memory, and exposes them via an **HTTP API** and a simple **web interface**.

---

## 📌 Features

- Connects to Kafka and processes messages in real time.
- Stores valid order data in PostgreSQL using transactions.
- Caches recent orders in memory for fast repeated queries.
- Preloads the cache from the database on startup.
- HTTP API:
  - `GET /orders/{order_uid}` — returns order details as JSON.
- Web interface:
  - Minimal HTML page to enter an order_id and view order data via the API.

---

## Tech Stack

- **Go** — main language.
- **Kafka** — message broker.
- **PostgreSQL** — database.
- **Gorilla Mux** — HTTP router.
- **Zap** — structured logging.
- **Docker & Docker Compose** — local infrastructure.

---

## Architecture Overview

[Kafka Topic] --> [Consumer] --> [Parser/Validator] --> [PostgreSQL]
| ^
v |
[In-memory Cache] ---+
|
[HTTP API]
|
[Web UI (/web)]

- Consumer subscribes to a Kafka topic with orders.
- Parser/Validator processes incoming JSON, discarding/logging invalid messages.
- Repository stores the order model in PostgreSQL atomically.
- Cache keeps recent orders in memory (map) and is reloaded from DB on startup.
- HTTP API retrieves orders by order_uid (from cache first, DB fallback).
- Web UI — static page that queries the API.

## Repository Structure

- cmd/service/ — service entrypoint (main).
- config/ — configuration files / environment defaults.
- internal/ — domain logic (consumer, producer, cache, repository, http-handlers).
- pkg/ — shared packages (logger, postgres).
- migrations/ — SQL migrations for PostgreSQL.
- web/ — static frontend (HTML).
- compose.yaml — Docker Compose configuration for local infra..
- Dockerfile, .dockerignore — containerization.
