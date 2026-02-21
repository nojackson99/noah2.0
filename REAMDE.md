# Personal Assistant (Learning Project)

## Goal
Build a cloud-backed personal planning assistant that:
- Ingests my personal data (calendar, notes, etc.)
- Helps me plan my week/day
- Starts simple (read-only, CLI-driven)
- Grows into a more agentic system over time

## Philosophy
- Learning-first, not speed-first
- Prefer explicit implementations over magic
- Use managed infra only where it doesn't reduce learning

## Tech Stack
- Go backend (Gin)
- Postgres (Railway in prod, local for dev)
- CLI client (Go) — not yet built
- React UI later (Vercel)
- RAG-style retrieval with pgvector
- Text-only v1

---

## Milestones

### ✅ Week 1 — API Skeleton
**What was built:**
- Go project structure under `api/` with internal packages
- Gin HTTP server with request logging and panic recovery
- `GET /health` — simple liveness check
- `POST /chat` — sends a message to OpenAI and returns the reply
- Config loaded from `.env` via `godotenv` (local) or env vars (Railway)
- Postgres connection pool set up in `api/internal/db/db.go` (not yet wired)

**Key files:**
- `api/main.go` — entrypoint, wires everything together
- `api/internal/config/config.go` — config struct, loaded from env
- `api/internal/httpserver/` — Gin server + routes
- `api/internal/llm/openai.go` — OpenAI HTTP client
- `api/internal/db/db.go` — pgxpool wrapper (dormant until Week 2)

---

### ✅ Week 2 — Google Calendar Ingestion
**What was built:**
- `POST /ingest/calendar` endpoint that fetches events from Google Calendar and stores them in Postgres
- OAuth 2.0 using a stored refresh token (no callback server — personal use only)
- Events are fetched for a 37-day window: 7 days past through 30 days future
- Recurring events are expanded into individual instances (e.g. weekly standup shows up once per week)
- Upserts are idempotent — running ingest multiple times won't create duplicate rows
- DB connection is now active and wired into the server

**How the OAuth works:**
- One-time setup: used Google OAuth Playground to authorize and get a refresh token
- Refresh token is stored in `.env` as `GOOGLE_REFRESH_TOKEN`
- At runtime, the Go `oauth2` library silently exchanges the refresh token for short-lived access tokens automatically
- No user interaction needed after initial setup

**Key files:**
- `api/internal/calendar/client.go` — Google Calendar API wrapper, fetches events
- `api/internal/calendar/store.go` — upserts events into Postgres
- `api/internal/httpserver/routes/ingest.go` — the `POST /ingest/calendar` handler
- `api/db/migrations/001_calendar_events.sql` — schema for the `calendar_events` table

**Database table: `calendar_events`**

| Column | Type | Notes |
|---|---|---|
| `google_event_id` | TEXT (PK) | Used for idempotent upserts |
| `summary` | TEXT | Event title |
| `description` | TEXT | Event body |
| `location` | TEXT | |
| `start_time` | TIMESTAMPTZ | |
| `end_time` | TIMESTAMPTZ | |
| `is_all_day` | BOOLEAN | True for all-day events |
| `status` | TEXT | confirmed, tentative, cancelled |
| `html_link` | TEXT | Link back to Google Calendar |
| `ingested_at` | TIMESTAMPTZ | When first fetched |
| `updated_at` | TIMESTAMPTZ | Updated on every re-ingest |

**To run locally:**
```bash
# One-time: create the DB and run the migration
createdb noah2
psql "postgres://nojackson@localhost:5432/noah2?sslmode=disable" \
  -f api/db/migrations/001_calendar_events.sql

# Start the server
go run api/main.go

# Trigger ingest
curl -s -X POST http://localhost:8080/ingest/calendar
```

**Required env vars (in `.env` at repo root):**
```
DATABASE_URL=postgres://nojackson@localhost:5432/noah2?sslmode=disable
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_REFRESH_TOKEN=...
```

---

### Week 3 — Markdown Notes Ingestion + Embeddings (next)
- Read `.md` files from a local notes folder
- Chunk them into smaller pieces
- Generate embeddings via OpenAI
- Store in Postgres with `pgvector` for semantic search

### Week 4 — Weekly Planning Endpoint
- Query calendar events + relevant notes
- Send to LLM with a planning prompt
- Return a structured weekly plan

---

## Local Dev Setup

**Prerequisites:**
- Go 1.21+
- PostgreSQL 16 (via Homebrew: `brew install postgresql@16`)
- `export PATH="/opt/homebrew/opt/postgresql@16/bin:$PATH"`

**First time:**
```bash
brew services start postgresql@16
createdb noah2
psql "postgres://nojackson@localhost:5432/noah2?sslmode=disable" \
  -f api/db/migrations/001_calendar_events.sql
```

**Every time:**
```bash
go run api/main.go
```