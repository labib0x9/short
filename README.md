# short — URL Shortener

A URL shortener service written in Go, built with a Domain-Driven Design (DDD) architecture. Uses raw `net/http` (no framework), PostgreSQL for persistence, Redis for caching and rate limiting, and RabbitMQ for async click analytics processing.

## Project Structure

```
cmd/                    → entrypoint & wiring
  short/                → server's main.go              
  migration/            → migration's main.go
config/                 → environment config
internal/
  domain/               → entities, repository interfaces, domain errors
    url/
    queue/              → queue interface & ClickEvent entity
    cache/              → cache interface
  app/                  → application services
    url/              
  infra/
    postgres/           → PostgreSQL repository implementations
    redis/              → Redis client
      cache/            → Cache repo implementation
      rate_limitter/    → Rate limiter repo implementation and Lua script
    rabbitmq/           → RabbitMQ connection, publisher, consumer
  transport/
    http/               → server, middleware manager, route handlers
  worker/               → async analytics consumer (RabbitMQ)
  utils/                → code generation, JSON helpers
  cron/                 → Cron cleaner 
migrations/             → SQL migration files
static/                 → Frontend code (claude generated)
test/                   → Test files (k6 load test)
```

## Tech Stack

| Concern | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP | `net/http` (stdlib only, no framework) |
| Database | PostgreSQL (`sqlx`, `lib/pq`) |
| Migrations | `golang-migrate` |
| Cache | Redis (`go-redis/v9`) |
| Message Queue | RabbitMQ (`amqp091-go`) |
| Validation | `go-playground/validator` |
| ID Generation | `google/uuid` |
| User-Agent Parsing | `mileusna/useragent` |

## Features

- **URL shortening** — FNV-64 URL hash XOR'd into a UUID, base64url-encoded and truncated to 8 characters 
- **Redirect** — `GET /:code` with Redis cache layer; expired URLs served from cache to avoid DB hits
- **Click analytics** — async pipeline via RabbitMQ; captures browser, OS, device type, referrer, and timestamp per click
- **Analytics endpoint** — `GET /:code/stat` returns aggregated browser/OS/device breakdown and total click count
- **Rate limiting** — token bucket implemented as an atomic Lua script in Redis
- **Dead-letter queue** — failed analytics messages are routed to `analytics.queue.dead` after exhausting retries
- **Graceful shutdown** — SIGINT/SIGTERM handling with configurable drain timeout
- **Cleaner** - Cron job to delete expired urls

## API

### Shorten a URL
```
POST /short
Content-Type: application/json

{
  "url": "https://example.com/some/long/path",
  "expire_at": "2025-12-31T00:00:00Z"   // optional
}
```
```json
{
  "msg": "success",
  "code": "aB3kZ9",
  "short_url": "http://localhost:3000/aB3kZ9",
  "expire_at": "2025-12-31T00:00:00Z"
}
```

### Redirect
```
GET /:code
→ 302 redirect to original URL
```

### Analytics
```
GET /:code/stat
```
```json
{
  "short": "aB3kZ9",
  "total_count": 142,
  "browser": { "Chrome": 98, "Firefox": 44 },
  "device":  { "Desktop": 120, "Mobile": 22 },
  "os":      { "Windows": 80, "macOS": 40, "Linux": 22 },
  "expire_at": null
}
```

## Click Analytics Pipeline

```mermaid
sequenceDiagram
    participant User
    participant API as URL Service
    participant MQ as RabbitMQ
    participant Worker as Analytics Worker
    participant DB as PostgreSQL

    User->>API: GET /:code
    API-->>User: 302 Redirect

    API->>MQ: Publish ClickEvent (async)

    MQ->>Worker: Consume ClickEvent

    Worker->>Worker: Parse User-Agent
    Note right of Worker: Detect browser,\ndevice, OS

    Worker->>DB: Insert click record
    Worker->>DB: Increment total clicks
    Worker->>DB: Update last_clicked_at

    alt Success
        DB-->>Worker: OK
    else Failure
        Worker->>MQ: Retry (max 2x)
        MQ->>Worker: Requeue event

        alt Retry limit exceeded
            Worker->>MQ: Dead Letter Queue
        end
    end
```

The worker runs with configurable concurrency (default 10) using a semaphore pattern, and uses manual acknowledgement (`autoAck: false`) so no clicks are lost on crash.

## Setup

### Prerequisites

- Go 1.26+
- PostgreSQL
- Redis
- RabbitMQ

### Environment

Copy `.env.example` to `.env` and fill in your values:

```env
ADDR=127.0.0.1
PORT=3000
PREFIX=http://localhost:3000/
SERVICE_NAME=short-api

PG_USER=shortuser
PG_PASSWORD=secret
PG_PORT=5432
PG_ADDRESS=localhost
PG_NAME=urlshortener
PG_SSLMODE=disable

PG_SUPERUSER=postgres
PG_SUPERDB=postgres

REDIS_ADDR=localhost:6379

RMQ_ADDR=localhost:5672
RMQ_USER=guest
RMQ_PASS=guest
```

### Run Migration
See migration section

### Run Server

```bash
go run ./cmd/short/main.go
```

The application will:
1. Connect to PostgreSQL, Redis and RabbitMQ
2. Declare the `analytics.queue` and `analytics.queue.dead` queues
3. Start the analytics worker goroutine
4. Start the HTTP server


## Migration

- Migrations are managed with `golang-migrate` and need to run manually via migrate cli.

```bash
# connects to super user and create user, database
go run ./cmd/migration/main.go -setup

# run the migration
go run ./cmd/migration/main.go -up

# rollback migration level-1
go run ./cmd/migration/main.go -down
```

## Performance / Load Testing

Load tested with [k6](https://k6.io) at constant arrival rates with 50-500 max VUs, 60s sustained load per run.

**Note:** I have tested uncommenting rate limiter on '../transport/server.go' file

**Run:** `k6 run -e RATE=<n> ./test/load.js`

| Target Rate | p95 Latency | Error Rate | Achieved Throughput |
|---|---|---|---|
| 20 req/s    | 54ms   | 0% | 20/s   |
| 100 req/s   | 20ms   | 0% | 100/s  |
| 500 req/s   | 979ms  | 0% | 487/s  |
| 1000 req/s  | 3.62s  | 0% | 403/s* |


**Findings:**
- Zero request failures across all load levels
- **Redis keyspace hygiene significantly affects tail latency under load.** 
  Clearing accumulated `short:*` keys between test runs nearly halved 
  p95 latency at 500 req/s (1.96s → 979ms) and improved throughput by 
  ~46% at 1000 req/s (276/s → 403/s)


## Future-Work
- Upgrade token bucket rate limiting to sliding window
- Dockerize the entire app
- Follow ACID Principle on database query
- Query optimization
