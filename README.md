# Order Service

A demo microservice in Go for processing and displaying orders.
The service consumes messages from **Kafka**, stores them in **PostgreSQL**, caches recent orders in memory, and exposes data via an **HTTP API** and a simple **web interface**.

---

## 📌 Features

- Connects to Kafka and processes messages in real time.
- Validates and stores order data in PostgreSQL.
- In-memory caching of recent orders for fast access.
- Cache warm-up from database on service startup.
- HTTP API:
  - `GET /orders/{order_uid}` — returns order details as JSON.
- Web interface:
  - Minimal HTML page to search for an order by ID and display results.

---

## 🛠️ Tech Stack

- **Go** — main language.
- **Kafka** — message broker.
- **PostgreSQL** — database.
- **Gorilla Mux** — HTTP router.
- **Zap** — structured logging.
- **Docker & Docker Compose** — local environment setup.
- **Makefile** — automation.

---

## 🏗️ Architecture
