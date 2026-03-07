-- Run with:
--   psql $DATABASE_URL -f api/db/migrations/002_note_chunks.sql

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS note_chunks (
    id           BIGSERIAL PRIMARY KEY,
    file_path    TEXT NOT NULL,
    chunk_index  INT  NOT NULL,
    content      TEXT NOT NULL,
    embedding    vector(1536),
    ingested_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (file_path, chunk_index)
);

CREATE INDEX IF NOT EXISTS note_chunks_embedding_idx
    ON note_chunks USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 10);
