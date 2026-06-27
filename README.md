# URL Shortener

Starter setup for a simple URL shortener API in Go. This project keeps the structure small and readable while separating responsibilities into clear layers.

## Stack

- Go
- Chi router
- PostgreSQL with `pgxpool`
- Redis with `go-redis`
- `golang-migrate` for SQL migrations
- Air for live reload
- Docker Compose for local services

## Project Structure

```text
url-shortener/
  cmd/
    api/
      main.go
    worker/
      main.go
  internal/
    analytics/
      dto.go
      producer.go
      repository.go
      worker.go
    app/
      app.go
      dependencies.go
      routes.go
      server.go
    cache/
      redis.go
    config/
      config.go
    database/
      postgres.go
    health/
      handler.go
    request/
      json.go
    response/
      json.go
    shorturl/
      cache.go
      dto.go
      errors.go
      generator.go
      handler.go
      model.go
      repository.go
      service.go
    store/
      storage.go
  migrations/
    000001_create_short_urls_table.down.sql
    000001_create_short_urls_table.up.sql
    000002_create_url_clicks_table.down.sql
    000002_create_url_clicks_table.up.sql
  .air.toml
  .env.example
  docker-compose.yml
  Makefile
  README.md
  go.mod
```

## Prerequisites

- Go installed
- Docker Desktop or Docker Engine
- `air` installed: `go install github.com/air-verse/air@latest`
- `migrate` CLI installed: `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

## Environment Setup

Copy the example environment file:

```bash
cp .env.example .env
```

If you are on Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

`APP_ENV=local` enables loading `.env` automatically through `godotenv`.

If you create `.env` manually, include:

```env
APP_BASE_URL=http://localhost:8080
SHORT_CODE_LENGTH=7
SHORT_URL_CACHE_TTL=24h
SHORT_URL_NEGATIVE_CACHE_TTL=1m
CLICK_STREAM_NAME=url_clicks
CLICK_STREAM_GROUP=url_click_workers
CLICK_STREAM_CONSUMER=worker-1
CLICK_STREAM_BATCH_SIZE=50
CLICK_STREAM_BLOCK_TIME=5s
WORKER_METRICS_PORT=9091
RATE_LIMIT_ENABLED=true
RATE_LIMIT_CREATE_LIMIT=10
RATE_LIMIT_CREATE_WINDOW=1m
RATE_LIMIT_RESOLVE_LIMIT=100
RATE_LIMIT_RESOLVE_WINDOW=1m
RATE_LIMIT_FAIL_OPEN=true
```

## Start PostgreSQL and Redis for Local Go Development

```bash
docker compose up -d postgres redis
```

This starts:

- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`

## Docker Compose Full Stack

Build and start PostgreSQL, Redis, two API instances, the analytics worker, Nginx, and Prometheus:

```bash
docker compose up -d --build
```

Run database migrations from the Compose migration service:

```bash
docker compose run --rm migrate
```

The migration service uses the `tools` profile, so it runs only when explicitly requested.

Check service status:

```bash
docker compose ps
```

The API is available through Nginx at <http://localhost:8080/health>. Nginx uses round-robin load balancing between `api-1` and `api-2`.

Prometheus is available at <http://localhost:9090>. It scrapes both API instances and the worker through the Docker network. Worker metrics are also mapped to <http://localhost:9091/metrics>.

Run the included load test against Nginx:

```bash
py scripts/load_test.py
```

PostgreSQL, Redis, and the worker each remain a single instance. This setup tests horizontal scaling at the API layer only.

## Run Migrations

The `Makefile` loads `.env` when it exists, so `DATABASE_URL` is available to the migrate commands.

```bash
make migrate-up
```

To roll back:

```bash
make migrate-down
```

To create a new migration:

```bash
make migrate-create name=add_users_table
```

## Run the App

Run directly with Go:

```bash
make run
```

Run with Air for live reload:

```bash
make dev
```

## Run the Analytics Worker

The API records redirects by adding click events to a Redis Stream. The worker consumes those events and inserts them into PostgreSQL, so analytics failures never delay or fail a redirect.

Start the worker in a separate terminal:

```bash
make worker
```

## Test the Health Endpoint

Once the app is running:

```bash
curl http://localhost:8080/health
```

Expected success response:

```json
{
  "status": "ok",
  "postgres": "ok",
  "redis": "ok"
}
```

If PostgreSQL or Redis is unavailable, the endpoint returns HTTP `503` and marks the failing dependency as `down`.

## Create a Short URL

Send a `POST` request to `/api/urls`:

```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d "{\"url\":\"https://example.com/very/long/url\"}"
```

Expected response:

```json
{
  "code": "aB91xZk",
  "short_url": "http://localhost:8080/aB91xZk",
  "original_url": "https://example.com/very/long/url"
}
```

Validation rules:

- `url` is required
- invalid JSON returns `400`
- only `http` and `https` URLs are allowed

## Redirect to the Original URL

Open a short URL with `GET /{code}`:

```bash
curl -i http://localhost:8080/aB91xZk
```

Expected response:

```http
HTTP/1.1 302 Found
Location: https://example.com/very/long/url
```

Error responses:

- `404` if the code does not exist
- `410` if the code is inactive or expired

## Distributed Rate Limiting

The API uses an atomic Redis Lua token bucket keyed by normalized client IP. Because both API instances use the same Redis database, limits apply across the full deployment rather than per process.

- `POST /api/urls` uses `RATE_LIMIT_CREATE_LIMIT` and `RATE_LIMIT_CREATE_WINDOW`.
- `GET /{code}` uses `RATE_LIMIT_RESOLVE_LIMIT` and `RATE_LIMIT_RESOLVE_WINDOW`.
- `RATE_LIMIT_ENABLED` enables or disables both limiters.
- `RATE_LIMIT_FAIL_OPEN=true` allows requests when the Redis rate-limit operation fails. `false` returns HTTP `503`.

Defaults allow 10 creates and 100 resolves per minute. A rejected request returns:

```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 0
Retry-After: 6
Content-Type: application/json

{"error":"rate limit exceeded"}
```

Successful limited requests also include `X-RateLimit-Limit` and `X-RateLimit-Remaining`. Nginx overwrites `X-Forwarded-For` and `X-Real-IP` with the direct client address so callers cannot supply an arbitrary rate-limit key.

The load-test script supports focused scenarios:

```powershell
$env:SCENARIO="normal"; py scripts/load_test.py
$env:SCENARIO="create-limit"; $env:BURST_REQUESTS="50"; py scripts/load_test.py
$env:SCENARIO="resolve-limit"; $env:BURST_REQUESTS="200"; py scripts/load_test.py
```

The two limit scenarios should report `429` responses and confirm the expected rate-limit headers. To exercise failure behavior, stop Redis after the API has started and send a request: fail-open continues to the handler, while fail-closed returns `503`.

## Redis Redirect Cache

Short URL redirects use Redis as a small cache in front of PostgreSQL:

- Creating a short URL writes it to PostgreSQL first, then caches `shorturl:{code}` with the original URL.
- Redirects check `shorturl:{code}` before querying PostgreSQL. A cache miss loads the URL from PostgreSQL and caches it.
- Unknown codes are temporarily cached as `shorturl:notfound:{code}` to avoid repeated database lookups.
- `SHORT_URL_CACHE_TTL` controls original URL entries and defaults to `24h`.
- `SHORT_URL_NEGATIVE_CACHE_TTL` controls unknown-code entries and defaults to `1m`.

Redis errors never fail URL creation or redirects. PostgreSQL remains the source of truth.

## Click Analytics

Each successful redirect publishes a best-effort event to the `CLICK_STREAM_NAME` Redis Stream. The worker creates the configured consumer group, reads events in batches, writes them to `url_clicks`, and acknowledges only successfully inserted events.

To verify analytics:

1. Start PostgreSQL and Redis, then run `make migrate-up`.
2. Run the API with `make dev` and the worker with `make worker` in another terminal.
3. Create a short URL and request its redirect.
4. Query PostgreSQL:

```sql
SELECT code, ip_address, user_agent, referer, clicked_at
FROM url_clicks
ORDER BY id DESC;
```

## Observability

The API exposes Prometheus metrics at <http://localhost:8080/metrics>. The analytics worker exposes its own process metrics at <http://localhost:9091/metrics>.

To run Prometheus locally, start the included Compose service:

```bash
docker compose up -d
```

Prometheus is available at <http://localhost:9090> and scrapes `api-1:8080`, `api-2:8080`, and `worker:9091` through the Docker network.

During a load test, watch these metrics:

- `http_requests_total`
- `http_request_duration_seconds`
- `shorturl_cache_hits_total`
- `shorturl_cache_misses_total`
- `analytics_events_produced_total`
- `analytics_worker_events_inserted_total`
- `analytics_worker_insert_errors_total`
- `rate_limited_requests_total{type="create|resolve"}`
- `rate_limit_errors_total{type="create|resolve"}`

## Why This Structure Works Well for Learning

This setup keeps the app small enough to understand quickly:

- `cmd/api/main.go` only starts the app.
- `internal/app` handles bootstrap, routes, and the HTTP server lifecycle.
- `internal/store` registers repositories in one place.
- `internal/shorturl` contains the short URL feature.
- `internal/health` contains the health endpoint.
- `internal/request` and `internal/response` hold shared HTTP helpers.

The first feature already uses a simple flow:

`Handler -> Service -> Repository`

- Handler: decodes JSON, validates the request shape, and returns HTTP responses.
- Service: validates and normalizes the URL, generates the code, hashes the URL, and handles retry logic.
- Repository: inserts the record with raw SQL using `pgxpool`.

That gives you clean separation without introducing too much abstraction too early.
