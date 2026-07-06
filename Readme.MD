# URL Shortener (Go)

A backend URL shortener built to learn how background job processing works in
practice — specifically, how a producer (API) and a consumer (worker) can be
built as two independent processes coordinating through a queue, instead of
doing everything in one request/response cycle.

This was built as a learning project, not a production service. Scope was
kept deliberately small so the core concept — async job processing — stayed
the focus, instead of getting lost in auth, rate limiting, or deployment
concerns.

## What it does

1. `POST /shorten` — submit a long URL, get back a short code.
2. `GET /{code}` — visiting the short code redirects to the original URL.
3. Every redirect also fires a "click" event into a Redis queue —
   asynchronously, without slowing down the redirect itself.
4. A separate **worker** process consumes that queue and writes click
   records (user agent, IP, timestamp) into Postgres.

The redirect never waits on the database write. That's the whole point of
the architecture.

## Why it's built this way

- **WebSocket-free, plain HTTP** — this project is about background jobs,
  not real-time messaging, so a simple request/response API is enough.
- **Redis as a queue, not a cache** — `LPUSH` on the API side, `BRPOP` on the
  worker side. `BRPOP` blocks until something arrives, so the worker reacts
  instantly instead of polling on a timer. No caching layer was added in
  this version — every redirect still hits Postgres directly.
- **Two separate binaries (`cmd/api`, `cmd/worker`)** — the API and worker
  are independent programs, each with their own `main.go`, sharing logic
  through `internal/`. This mirrors how you'd actually deploy this in
  production: as two separate services that scale independently.
- **Foreign key from `clicks` to `urls`** — `clicks.url_id` references
  `urls.id` with `ON DELETE CASCADE`, so click history is automatically
  cleaned up if a URL is ever deleted. This was a deliberate choice over the
  simpler (but less correct) option of duplicating the short code into every
  click row.

## Architecture

```
   POST /shorten                      GET /{code}
        │                                  │
        ▼                                  ▼
  ┌─────────────────────────────────────────────┐
  │                 API server                    │
  │        (cmd/api — chi router)                 │
  └───────┬───────────────────────────┬───────────┘
          │                           │
          ▼                           ▼
     Postgres (urls)            Redis LPUSH
     save/read short code       ("click_events")
                                       │
                                       ▼
                              ┌──────────────────┐
                              │      Worker       │
                              │  (cmd/worker)      │
                              │  Redis BRPOP  loop  │
                              └────────┬───────────┘
                                       ▼
                              Postgres (clicks)
                              linked via url_id
```

## Tech stack

| Piece | Tool |
|---|---|
| Language | Go |
| HTTP router | [chi](https://github.com/go-chi/chi) |
| Database | PostgreSQL |
| DB driver | [lib/pq](https://github.com/lib/pq) |
| Queue / job broker | Redis (via [go-redis/v9](https://github.com/redis/go-redis)) |
| Short code generation | [go-nanoid](https://github.com/jaevor/go-nanoid) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Config | [godotenv](https://github.com/joho/godotenv) |

## Project structure

```
cmd/
  api/main.go        entrypoint for the HTTP server
  worker/main.go      entrypoint for the queue consumer
internal/
  db/                 Postgres connection setup
  handler/            HTTP handlers (shorten, redirect)
  repository/          all SQL lives here (urls + clicks)
  shortner/            short code generation
  redis_client/        Redis connection + queue push/pop + event type
migrations/            up/down SQL migrations
```

## Database schema

**urls**

| column | type |
|---|---|
| id | BIGSERIAL PRIMARY KEY |
| short_code | TEXT UNIQUE NOT NULL |
| original_url | TEXT NOT NULL |
| created_at | TIMESTAMP DEFAULT NOW() |

**clicks**

| column | type |
|---|---|
| id | BIGSERIAL PRIMARY KEY |
| url_id | BIGINT REFERENCES urls(id) ON DELETE CASCADE |
| user_agent | TEXT |
| ip_address | TEXT |
| clicked_at | TIMESTAMP DEFAULT NOW() |

## Running it locally

Requires Postgres and Redis running locally, and a `.env` file with:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=urlshortner_user
DB_PASSWORD=your_password
DB_NAME=url_shortner
DB_SSLMODE=disable
REDIS_ADDR=localhost:6379
```

Apply migrations:

```bash
migrate -path migrations \
  -database "postgres://urlshortner_user:your_password@localhost:5432/url_shortner?sslmode=disable" \
  up
```

Run the API and the worker in two separate terminals — both need to be
running at the same time, since they're independent processes coordinating
through Redis:

```bash
go run cmd/api/main.go
```

```bash
go run cmd/worker/main.go
```

## Trying it out

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.wikipedia.org"}'
```

```bash
curl -v http://localhost:8080/YOUR_SHORT_CODE
```

Watch the worker's terminal — it should print a "processed click event"
line within milliseconds of the redirect happening. Check Postgres to
confirm the click landed:

```bash
psql -d url_shortner -c "SELECT * FROM clicks;"
```

## What was deliberately left out

This project stayed intentionally small so the async job-processing pattern
stayed the focus:

- **No rate limiting** — for a single-instance app, `httprate` (in-memory,
  per-process limiting) would be the right tool here, not Redis-backed
  limiting, since there's no need to share state across instances that don't
  exist yet.
- **No Docker** — everything runs directly on the host for now.
- **No caching layer** — every redirect hits Postgres directly. A natural
  next step would be caching `short_code → original_url` lookups in Redis
  ahead of the database, which would be a second, distinct use of Redis
  alongside its existing role as the job queue.
- **No auth** — anyone can shorten a URL; this isn't a multi-tenant service.

## What this project demonstrates

- Decoupling a fast user-facing action (redirect) from slower background
  work (persistence) using a queue.
- Using Redis's blocking pop (`BRPOP`) for event-driven job consumption
  instead of a polling loop with `time.Sleep`.
- Running two independently deployable Go binaries from one codebase,
  sharing logic through `internal/` packages.
- A relationally correct schema (foreign key + cascade) instead of
  denormalized duplication.