### Building and running your application

When you're ready, start your application by running:
`docker compose up --build`.

Your application will be available at http://localhost:8081.

### Deploying your application to the cloud

First, build your image, e.g.: `docker build -t myapp .`.
If your cloud uses a different CPU architecture than your development
machine (e.g., you are on a Mac M1 and your cloud provider is amd64),
you'll want to build the image for that platform, e.g.:
`docker build --platform=linux/amd64 -t myapp .`.

Then, push it to your registry, e.g. `docker push myregistry.com/myapp`.

Consult Docker's [getting started](https://docs.docker.com/go/get-started-sharing/)
docs for more detail on building and pushing.

### References

- [Docker's Go guide](https://docs.docker.com/language/golang/)






Простой сервис заказов: принимает сообщения из Kafka, валидирует и сохраняет заказы в Postgres, кэширует их в памяти и отдаёт по HTTP.

## Содержание
- Быстрый старт
- Конфигурация
- Запуск локально (без Docker)
- Архитектура и основные компоненты
- HTTP API
- Kafka
- База данных и миграции
- Кэш: прогрев, инвалидация
- Валидация и обработка ошибок
- Грейсфул-шатдаун
- Тесты
- Траблшутинг

## Быстрый старт
Требования: Docker + Docker Compose.

```bash
# из корня проекта
docker compose up -d --build
# HTTP сервер: http://localhost:8081
# Kafka UI (Kafdrop): http://localhost:9000
# Postgres: localhost:5432
```

Проверка:
- Откройте `http://localhost:8081` (статическая страница из `web/`).
- Эндпоинт: `GET http://localhost:8081/orders/{order_uid}`.
- Продюсер публикует тестовые заказы в Kafka каждые 5 секунд.

Остановка:
```bash
docker compose down -v
```

## Конфигурация
Файл по умолчанию: `config/config.yaml` (монтируется в контейнер сервисов). Можно переопределять переменными окружения.

Параметры:
- server.port — порт HTTP сервера
- postgres.host, port, user, password, dbname — доступ к БД
- kafka.broker, topic, group_id — подключение к Kafka

ENV (пример из `compose.yaml`):
- CONFIG_PATH=/config/config.yaml
- POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB
- KAFKA_BROKER, KAFKA_TOPIC, KAFKA_GROUP_ID

## Запуск локально (без Docker)
Требования: Go 1.23+, Postgres, Kafka.

```bash
export CONFIG_PATH=$(pwd)/config/config.yaml
# убедитесь, что Postgres и Kafka доступны как в конфиге

go run ./cmd/service
```

Тесты:
```bash
go test -v ./...
```

## Архитектура и основные компоненты
- `cmd/service/main.go` — точка входа, инициализация, прогрев кэша, запуск HTTP/Kafka воркеров, graceful shutdown.
- `internal/handlers` — HTTP сервер (Gorilla Mux), `GET /orders/{order_uid}`: сначала кэш, при промахе — БД, затем в кэш.
- `internal/kafka` — producer (тестовые данные), consumer (чтение, валидация, сохранение, кэширование, commit).
- `internal/repository/postgres` — репозиторий на pgx с транзакциями и выборками.
- `internal/storage` — in-memory кэш с mutex-ами и методами `Get/Set/Invalidate/InvalidateAll`.
- `internal/models` — структуры данных заказа.
- `pkg/database` — подключение к Postgres (ping).
- `pkg/logger` — zap-сахарный логгер.

Поток данных:
1) Producer публикует заказ в Kafka (JSON).
2) Consumer читает, валидирует, сохраняет в БД транзакционно, кладёт в кэш, подтверждает сообщение (commit).
3) HTTP `GET /orders/{order_uid}` отдаёт из кэша; при промахе — читает из БД и кэширует.
4) При старте выполняется прогрев кэша из БД (при ошибке старт прерывается).

## HTTP API
- `GET /orders/{order_uid}`
  - 200: JSON заказа
  - 404: заказ не найден
  - 400: некорректный запрос
  - 500: внутренняя ошибка

Пример ответа (усечён):
```json
{
  "order_uid": "...",
  "track_number": "...",
  "delivery": { "name": "...", "phone": "..." },
  "payment": { "transaction": "...", "amount": 123 },
  "items": [ { "chrt_id": 1, "name": "..." } ],
  "date_created": "2024-01-01T10:00:00Z"
}
```

## Kafka
- Используется `segmentio/kafka-go`.
- Конфигурируется через `kafka.*` в конфиге.
- Producer: отправляет фейковые заказы для демонстрации.
- Consumer: обрабатывает поток, валидирует и сохраняет с ретраями.

## База данных и миграции
- Postgres, схема создаётся миграцией из `migrations/0001_init.up.sql`.
- При запуске через compose, миграции применяются автоматически (`/docker-entrypoint-initdb.d`).

Таблицы: `delivery`, `payment`, `orders`, `items` (FK на `delivery` и `payment`, `items` привязаны к `orders.order_uid`).

## Кэш: прогрев, инвалидация
- Прогрев: при старте сервис загружает все `order_uid` и каждый заказ в кэш. При ошибке старт прерывается.
- Методы:
  - `Get(uid) (*Order, bool)`
  - `Set(uid, *Order)`
  - `Invalidate(uid)` — удалить элемент
  - `InvalidateAll()` — очистить кэш

## Валидация и обработка ошибок
- Валидация входящих сообщений: `go-playground/validator` на структуре заказа.
- Ошибки на слое БД оборачиваются `fmt.Errorf("...: %w", err)` и логируются на уровне consumer/handlers.
- Consumer: простой retry на сохранение в БД (3 попытки, 500ms между попытками), после успеха выполняется commit.

## Грейсфул-шатдаун
- HTTP сервер останавливается через `Shutdown(ctx)`.
- Consumer/Producer завершаются по `ctx.Done()`.

## Тесты
- Базовые юнит-тесты можно запускать `go test -v ./...`.
- Интеграционные тесты репозитория можно добавить при необходимости (с тестовой БД и миграциями).

## Траблшутинг
- "connection refused" к Postgres/Kafka: проверьте, что контейнеры живы (`docker compose ps`), и хосты/порты соответствуют `config.yaml`.
- Kafka topic отсутствует: включите авто‑создание в брокере или создайте топик вручную.
- Ошибки валидации: сообщения логируются и подтверждаются (не зацикливаются);
- ST1005 (error strings capitalized): все строки ошибок начинаются с маленькой буквы.