package notes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Chunk struct {
	FilePath   string
	ChunkIndex int
	Content    string
}

// ReadChunks walks notesDir recursively, reads all .md files, and splits them
// into paragraph-based chunks (split on double newline). Chunks under 50
// characters are skipped.
func ReadChunks(notesDir string) ([]Chunk, error) {
	if _, err := os.Stat(notesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("notes directory does not exist: %s", notesDir)
	}

	var chunks []Chunk

	err := filepath.WalkDir(notesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		paragraphs := strings.Split(string(data), "\n\n")
		chunkIdx := 0
		for _, p := range paragraphs {
			content := strings.TrimSpace(p)
			if len(content) < 50 {
				continue
			}
			chunks = append(chunks, Chunk{
				FilePath:   path,
				ChunkIndex: chunkIdx,
				Content:    content,
			})
			chunkIdx++
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk notes dir: %w", err)
	}

	return chunks, nil
}
