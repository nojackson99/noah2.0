package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/noah-jackson/noah2.0/api/internal/db"
	"github.com/noah-jackson/noah2.0/api/internal/llm"
	"github.com/noah-jackson/noah2.0/api/internal/notes"
)

// RegisterIngestNotes registers POST /ingest/notes.
// It reads markdown files from notesDir, generates embeddings, and upserts into Postgres.
func RegisterIngestNotes(r *gin.Engine, notesDir string, database *db.DB, llmClient *llm.Client) {
	r.POST("/ingest/notes", func(c *gin.Context) {
		if notesDir == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "NOTES_DIR is not configured"})
			return
		}

		chunks, err := notes.ReadChunks(notesDir)
		if err != nil {
			log.Printf("ERROR read note chunks: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read notes"})
			return
		}

		// Count unique files
		seen := make(map[string]struct{})
		for _, ch := range chunks {
			seen[ch.FilePath] = struct{}{}
		}
		filesFound := len(seen)

		var chunksWithEmbeddings []notes.ChunkWithEmbedding
		for _, ch := range chunks {
			embedding, err := llmClient.Embed(c.Request.Context(), ch.Content)
			if err != nil {
				log.Printf("ERROR embed chunk %s[%d]: %v", ch.FilePath, ch.ChunkIndex, err)
				c.JSON(http.StatusBadGateway, gin.H{"error": "failed to generate embedding"})
				return
			}
			chunksWithEmbeddings = append(chunksWithEmbeddings, notes.ChunkWithEmbedding{
				Chunk:     ch,
				Embedding: embedding,
			})
		}

		upserted, err := notes.UpsertChunks(c.Request.Context(), database, chunksWithEmbeddings)
		if err != nil {
			log.Printf("ERROR upsert note chunks: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store chunks"})
			return
		}

		log.Printf("INFO ingested %d note chunks from %d files", upserted, filesFound)
		c.JSON(http.StatusOK, gin.H{
			"files_found":      filesFound,
			"chunks_upserted":  upserted,
		})
	})
}
