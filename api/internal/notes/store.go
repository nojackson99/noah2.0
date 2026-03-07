package notes

import (
	"context"
	"fmt"

	"github.com/noah-jackson/noah2.0/api/internal/db"
	"github.com/pgvector/pgvector-go"
)

type ChunkWithEmbedding struct {
	Chunk
	Embedding []float32
}

// UpsertChunks inserts or updates note chunks in the database.
// Keyed on (file_path, chunk_index), so re-ingesting is idempotent.
func UpsertChunks(ctx context.Context, database *db.DB, chunks []ChunkWithEmbedding) (int64, error) {
	const q = `
		INSERT INTO note_chunks (file_path, chunk_index, content, embedding, ingested_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (file_path, chunk_index) DO UPDATE SET
			content    = EXCLUDED.content,
			embedding  = EXCLUDED.embedding,
			updated_at = NOW()
	`

	var upserted int64
	for _, c := range chunks {
		tag, err := database.Pool.Exec(ctx, q,
			c.FilePath,
			c.ChunkIndex,
			c.Content,
			pgvector.NewVector(c.Embedding),
		)
		if err != nil {
			return upserted, fmt.Errorf("upsert chunk %s[%d]: %w", c.FilePath, c.ChunkIndex, err)
		}
		upserted += tag.RowsAffected()
	}

	return upserted, nil
}
