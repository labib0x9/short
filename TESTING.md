# Testing Guide for `short` URL Shortener

This guide covers all testing approaches for the **short** project: setup, unit tests, integration tests, and end-to-end testing.

---

## 📋 Table of Contents

1. [Quick Start](#quick-start)
2. [Unit Tests](#unit-tests)
3. [Integration Tests](#integration-tests)
4. [End-to-End (E2E) Tests](#end-to-end-tests)
5. [Docker-Compose Setup](#docker-compose-setup)
6. [Manual Testing with cURL](#manual-testing-with-curl)
7. [Load Testing](#load-testing)
8. [CI/CD Testing](#cicd-testing)

---

## Quick Start

### Prerequisites

```bash
# Install dependencies
go mod download

# Verify Go version (1.22+)
go version
```

### Run All Unit Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Test

```bash
# Run only worker tests
go test -v ./internal/worker

# Run only one test function
go test -v -run TestHandle_Success ./internal/worker
```

---

## Unit Tests

### Existing Unit Tests

The project already has comprehensive unit tests for the analytics worker:

```bash
go test -v ./internal/worker
```

**Test file**: `internal/worker/analytics_worker_test.go`

**Test coverage**:
- ✅ `TestHandle_Success` — Message processed successfully
- ✅ `TestHandle_InvalidJSON` — Malformed JSON handled correctly
- ✅ `TestHandle_ServiceError_Retries` — Failed messages retried
- ✅ `TestHandle_ServiceError_MaxRetriesExceeded` — Dead-letter queue behavior
- ✅ `TestRun_ContextCancel` — Graceful shutdown on context cancel
- ✅ `TestRun_ChannelClosed` — Handles broker disconnection

### Running with Coverage

```bash
go test -v -coverprofile=worker.out ./internal/worker
go tool cover -html=worker.out -o worker_coverage.html
open worker_coverage.html
```

---

## Integration Tests

Integration tests require running services (PostgreSQL, Redis, RabbitMQ).

### Setup Test Environment

```bash
# Start services using Docker Compose
docker-compose -f docker-compose.test.yml up -d

# Wait for services to be ready
sleep 10

# Run integration tests
go test -v -tags=integration ./...

# Stop services
docker-compose -f docker-compose.test.yml down
```

---

## End-to-End Tests

E2E tests validate the entire flow: HTTP → Service → DB → Cache → RabbitMQ → Worker.

### Complete E2E Testing Flow

```bash
# 1. Start all services
docker-compose -f docker-compose.test.yml up -d

# 2. Run migrations
go run ./cmd/migration/main.go -setup
go run ./cmd/migration/main.go -up

# 3. Start the server (in another terminal)
go run ./cmd/short/main.go

# 4. Run manual tests (see Manual Testing section below)
```

---

## Docker-Compose Setup

### Create `docker-compose.test.yml`

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: short
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: short
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U short"]
      interval: 5s
      timeout: 5s
      retries: 5
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
    volumes:
      - redis_data:/data

  rabbitmq:
    image: rabbitmq:3.12-management-alpine
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
    ports:
      - "5672:5672"
      - "15672:15672"
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq

volumes:
  postgres_data:
  redis_data:
  rabbitmq_data:

networks:
  default:
    name: short-test-network
```

### Start Services

```bash
docker-compose -f docker-compose.test.yml up -d

# Wait for health checks
docker-compose -f docker-compose.test.yml exec postgres pg_isready -U short

# Check all services are healthy
docker-compose -f docker-compose.test.yml ps

# View logs
docker-compose -f docker-compose.test.yml logs -f

# Stop services
docker-compose -f docker-compose.test.yml down -v
```

---

## Manual Testing with cURL

### 1. Setup Environment

```bash
# Copy example config
cp .env.example .env

# Start services
docker-compose -f docker-compose.test.yml up -d

# Run migrations
go run ./cmd/migration/main.go -setup
go run ./cmd/migration/main.go -up

# Start the server (in another terminal)
go run ./cmd/short/main.go
```

### 2. Test Shorten Endpoint

```bash
# Create short URL
curl -X POST http://localhost:3000/short \
  -H "Content-Type: application/json" \
  -d '{"url":"https://www.example.com/very/long/path","expire_at":"2025-12-31T23:59:59Z"}'

# Expected Response (201 Created):
# {
#   "msg": "success",
#   "code": "aB3kZ9",
#   "short_url": "http://localhost:3000/aB3kZ9",
#   "expire_at": "2025-12-31T23:59:59Z"
# }
```

### 3. Test Redirect (Triggers Analytics)

```bash
# Redirect to original URL
curl -L http://localhost:3000/aB3kZ9

# You should be redirected to: https://www.example.com/very/long/path
# Check server logs for: "Message Published" (analytics event sent to RabbitMQ)
```

### 4. Test Analytics Endpoint

```bash
# Get analytics (wait 2-3 seconds for worker to process)
sleep 3
curl http://localhost:3000/aB3kZ9/stat

# Expected Response (200 OK):
# {
#   "short": "aB3kZ9",
#   "total_count": 1,
#   "browser": { "Chrome": 1 },
#   "device": { "Desktop": 1 },
#   "os": { "Linux": 1 },
#   "expire_at": "2025-12-31T23:59:59Z"
# }
```

### 5. Test Error Cases

```bash
# Invalid URL format
curl -X POST http://localhost:3000/short \
  -H "Content-Type: application/json" \
  -d '{"url":"not-a-url"}'
# Expected: 400 Bad Request

# Expired URL
curl -X POST http://localhost:3000/short \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","expire_at":"2020-01-01T00:00:00Z"}'
# Expected: 410 Gone

# Non-existent code
curl http://localhost:3000/nonexistent
# Expected: 404 Not Found

# Missing code parameter
curl http://localhost:3000//stat
# Expected: 400 Bad Request
```

### 6. Test Rate Limiter

```bash
# Send 50+ requests rapidly (rate limit is 5 req/s, capacity 10)
for i in {1..60}; do
  curl http://localhost:3000/aB3kZ9
done

# You should see:
# - First 10 succeed immediately
# - Rest get 429 Too Many Requests with Retry-After header
```

### 7. Monitor RabbitMQ Queue

```bash
# Check analytics queue length
curl -u guest:guest http://localhost:15672/api/queues/%2F/analytics.queue

# Check dead-letter queue length
curl -u guest:guest http://localhost:15672/api/queues/%2F/analytics.queue.dead

# Access RabbitMQ Management UI
# Browser: http://localhost:15672 (username: guest, password: guest)
```

### 8. Monitor Redis Cache

```bash
# Connect to Redis
redis-cli

# Check cached short URL
GET "short:aB3kZ9"

# Check expired URL cache
GET "expire:aB3kZ9"

# View all keys
KEYS *

# Flush all (⚠️ only in test environment)
FLUSHALL
```

### 9. Check PostgreSQL Data

```bash
# Connect to PostgreSQL
psql -h localhost -U short -d short

# List all shortened URLs
SELECT id, url, short, total, created_at, expire_at FROM urls;

# List all clicks for a URL
SELECT * FROM clicks WHERE url_id = '<uuid>' ORDER BY clicked_at DESC;

# Count clicks by browser
SELECT browser, COUNT(*) FROM clicks GROUP BY browser;

# Exit
\q
```

---

## Load Testing

### Using `vegeta` (Recommended)

Install:
```bash
go install github.com/tsenart/vegeta@latest
```

Create `load_test.txt`:
```
GET http://localhost:3000/aB3kZ9
```

Run load test:
```bash
# 100 requests per second for 30 seconds
vegeta attack -duration=30s -rate=100 -targets=load_test.txt | vegeta report

# Generate latency histogram
vegeta attack -duration=30s -rate=100 -targets=load_test.txt | vegeta dump | vegeta report --percentiles=50,95,99

# Export as JSON
vegeta attack -duration=30s -rate=100 -targets=load_test.txt | vegeta report -type=json > results.json
```

### Using `ab` (Apache Bench)

```bash
# 1000 requests, 100 concurrent
ab -n 1000 -c 100 http://localhost:3000/aB3kZ9/

# Output shows:
# - Requests per second
# - Time per request
# - Transfer rate
# - Connection times (min, mean, median, max)
```

### Using `wrk` (Fast HTTP Benchmarking)

Install:
```bash
git clone https://github.com/wg/wrk.git
cd wrk && make
```

Run test:
```bash
# 4 threads, 100 connections, 30 seconds
./wrk -t4 -c100 -d30s http://localhost:3000/aB3kZ9

# Output:
# Running 30s test @ http://localhost:3000/aB3kZ9
#   4 threads and 100 connections
#   Thread Stats   Avg      Stdev     Max   +/- Stdev
#     Latency     2.34ms   1.23ms  45.2ms   86.45%
#     Req/Sec   10.23k    1.34k   12.4k    72.56%
#   Latency Distribution
#      50%    2.10ms
#      75%    2.95ms
#      90%    4.20ms
#      99%   12.30ms
#   1222876 requests in 30.00s, 145.3MB read
#   Requests/sec: 40762.53
#   Transfer/sec: 4.84MB
```

---

## CI/CD Testing

Add to `.github/workflows/test.yml`:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_USER: short
          POSTGRES_PASSWORD: secret
          POSTGRES_DB: short
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

      rabbitmq:
        image: rabbitmq:3.12-alpine
        options: >-
          --health-cmd "rabbitmq-diagnostics ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5672:5672

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.26'
      - run: go test -v -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out
      - uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

---

## Test Checklist

- [ ] Unit tests pass: `go test ./...`
- [ ] Worker tests pass: `go test -v ./internal/worker`
- [ ] Coverage > 70%: `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
- [ ] Integration tests pass (with Docker): `docker-compose -f docker-compose.test.yml up -d && go test -v ./... && docker-compose -f docker-compose.test.yml down`
- [ ] E2E tests pass: Test all endpoints manually with cURL
- [ ] Shorten endpoint creates URLs correctly
- [ ] Redirect endpoint triggers analytics
- [ ] Analytics endpoint returns correct data
- [ ] Error cases handled correctly (400, 404, 410, 429)
- [ ] Rate limiter enforces 5 req/s limit
- [ ] Cache reduces database queries
- [ ] Worker processes analytics messages
- [ ] Dead-letter queue captures failed messages
- [ ] Load test: Verify latency < 10ms at 1000 req/s

---

## Troubleshooting

### Tests Fail with "Connection Refused"

```bash
# Ensure services are running
docker-compose -f docker-compose.test.yml ps

# If not running, start them
docker-compose -f docker-compose.test.yml up -d

# Check logs
docker-compose -f docker-compose.test.yml logs postgres
```

### Analytics Worker Not Processing Messages

```bash
# Check RabbitMQ queue length
curl -u guest:guest http://localhost:15672/api/queues/%2F/analytics.queue

# Check worker logs
go run ./cmd/short/main.go  # Should see "Analytics worker started"

# Manually check messages
redis-cli  # Then check if any messages stuck
```

### Cache Not Working

```bash
# Check Redis connection
redis-cli ping  # Should return PONG

# Check cache key
redis-cli get "short:aB3kZ9"

# View all cache keys
redis-cli KEYS "short:*"

# Flush all (⚠️ only in test environment)
redis-cli FLUSHALL
```

### Rate Limiter Failing

```bash
# Test Lua script
redis-cli SCRIPT EXISTS <sha>

# View rate limiter keys
redis-cli KEYS "rate*"

# Check current rate limit state
redis-cli HGETALL "rate:127.0.0.1"
```

### Migration Failures

```bash
# Check migration status
psql -h localhost -U short -d short -c "SELECT version, dirty FROM schema_migrations;"

# Rollback to previous version
go run ./cmd/migration/main.go -down

# Reapply migrations
go run ./cmd/migration/main.go -up
```

---

## Performance Benchmarks (Target Metrics)

| Metric | Target | Actual |
|--------|--------|--------|
| Redirect latency (p50) | < 2ms | |
| Redirect latency (p99) | < 10ms | |
| Throughput (req/s) | > 10k | |
| Cache hit rate | > 80% | |
| Analytics processing latency | < 100ms | |
| Worker error rate | < 0.1% | |

---

## Summary

| Type | Command | Time | Coverage |
|------|---------|------|----------|
| **Unit** | `go test ./...` | ~5s | High |
| **Integration** | `docker-compose up && go test -v ./...` | ~30s | Medium |
| **E2E** | Manual cURL tests | ~10 min | Complete flow |
| **Load** | `wrk -t4 -c100 -d30s` | ~30s | Performance |

---

**Start with unit tests, then layer in integration/E2E as needed for complete confidence.**
