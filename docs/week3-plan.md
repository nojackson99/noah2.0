# Week 3 Plan: Markdown Notes Ingestion + Embeddings

## Context

Weeks 1–2 built an API skeleton and Google Calendar ingestion. Week 3 adds the second data source: local markdown notes. The goal is to read `.md` files from a configured directory, chunk them, generate OpenAI embeddings, and store them in Postgres using `pgvector` — setting up semantic search that Week 4's planning endpoint will query.

---

## Scope

**New endpoint:** `POST /ingest/notes`
**Pattern:** mirrors calendar ingest — no request body, triggered by HTTP POST, returns `{files_found, chunks_upserted}`.

---

## Setup Step (before coding)

pgvector is not yet installed. This must be done once:

```bash
brew install pgvector          # installs the Postgres extension
psql "postgres://nojackson@localhost:5432/noah2?sslmode=disable" \
  -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

---

## Files to Create

### 1. `api/db/migrations/002_note_chunks.sql`

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE note_chunks (
    id           BIGSERIAL PRIMARY KEY,
    file_path    TEXT NOT NULL,
    chunk_index  INT  NOT NULL,
    content      TEXT NOT NULL,
    embedding    vector(1536),
    ingested_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (file_path, chunk_index)
);

CREATE INDEX ON note_chunks USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 10);
```

- `UNIQUE(file_path, chunk_index)` enables idempotent re-ingest (same pattern as `ON CONFLICT` in calendar store).
- `vector(1536)` matches `text-embedding-3-small` dimensions.
- IVFFlat index makes cosine similarity queries fast (needed by Week 4).

---

### 2. `api/internal/notes/reader.go`

Walk `NOTES_DIR` recursively; for each `.md` file, read content and split into chunks by double-newline (paragraph-based). Skip chunks under 50 characters.

```go
type Chunk struct {
    FilePath   string
    ChunkIndex int
    Content    string
}

func ReadChunks(notesDir string) ([]Chunk, error)
```

Returns all chunks from all files. Errors if the directory doesn't exist.

---

### 3. `api/internal/notes/store.go`

Upserts chunks into `note_chunks`. Mirrors `calendar/store.go` exactly.

```go
func UpsertChunks(ctx context.Context, database *db.DB, chunks []ChunkWithEmbedding) (int64, error)
```

Uses `ON CONFLICT (file_path, chunk_index) DO UPDATE SET content=..., embedding=..., updated_at=NOW()`.

---

### 4. `api/internal/httpserver/routes/ingest_notes.go`

```go
func RegisterIngestNotes(r *gin.Engine, notesDir string, database *db.DB, llmClient *llm.Client)
```

Handler logic:
1. Call `notes.ReadChunks(notesDir)` → get raw chunks
2. For each chunk, call `llmClient.Embed(ctx, chunk.Content)` → get `[]float32`
3. Call `notes.UpsertChunks(ctx, database, chunksWithEmbeddings)`
4. Return `{"files_found": N, "chunks_upserted": M}`

---

## Files to Modify

### 5. `api/internal/llm/openai.go`

Add an `Embed` method to the existing `Client` struct. Calls `POST https://api.openai.com/v1/embeddings` with model `text-embedding-3-small`. Returns `[]float32`.

```go
func (c *Client) Embed(ctx context.Context, text string) ([]float32, error)
```

Reuses the existing `httpClient` and `apiKey` fields. Hardcode model as `"text-embedding-3-small"` (no config needed — embeddings model is not the same concept as chat model).

---

### 6. `api/internal/config/config.go`

Add one field:
```go
NotesDir string
```

Load from env var `NOTES_DIR` with empty string default (handler will return a clear error if unset).

---

### 7. `api/main.go`

Wire in the new route:
```go
routes.RegisterIngestNotes(srv.Engine, cfg.NotesDir, database, openai)
```

The existing `openai` client is already constructed — it gets reused for embeddings (same `Client`, new method).

---

## pgvector Go Driver

The `pgx` driver (already in use) can store `[]float32` as a `vector` column using the `pgvector-go` package:

```bash
go get github.com/pgvector/pgvector-go
```

Use `pgvector.NewVector(embedding)` when binding the parameter in the store.

---

## `.env` Addition

```
NOTES_DIR=/path/to/your/notes
```

---

## Verification

1. Run the migration: `psql ... -f api/db/migrations/002_note_chunks.sql`
2. Add `NOTES_DIR=/path/to/notes` to `.env`
3. Start server: `go run api/main.go`
4. Trigger ingest: `curl -X POST http://localhost:8080/ingest/notes`
5. Confirm response: `{"files_found": N, "chunks_upserted": M}`
6. Spot-check DB: `SELECT file_path, chunk_index, LEFT(content, 80) FROM note_chunks LIMIT 10;`
7. Verify embeddings exist: `SELECT COUNT(*) FROM note_chunks WHERE embedding IS NOT NULL;`
