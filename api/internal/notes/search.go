package notes

import (
	"context"
	"fmt"

	"github.com/noah-jackson/noah2.0/api/internal/db"
	"github.com/pgvector/pgvector-go"
)

// SearchChunks returns the content of the top-k note chunks most similar
// to the given embedding, using cosine similarity.
func SearchChunks(ctx context.Context, database *db.DB, embedding []float32, k int) ([]string, error) {
	const q = `
		SELECT content
		FROM note_chunks
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $2
	`

	rows, err := database.Pool.Query(ctx, q, pgvector.NewVector(embedding), k)
	if err != nil {
		return nil, fmt.Errorf("search note chunks: %w", err)
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return nil, fmt.Errorf("scan note chunk: %w", err)
		}
		results = append(results, content)
	}

	return results, rows.Err()
}
