package chunker

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/deb2000-sudo/trackshift/pkg/models"
)

// ChunkerConfig controls how files are split into chunks.
type ChunkerConfig struct {
	MinChunkSize     int64
	MaxChunkSize     int64
	DefaultChunkSize int64
}

// normalize ensures sane defaults for the config.
func (c *ChunkerConfig) normalize() {
	if c.MinChunkSize == 0 {
		c.MinChunkSize = 5 * 1024 * 1024 // 5MB
	}
	if c.MaxChunkSize == 0 {
		c.MaxChunkSize = 200 * 1024 * 1024 // 200MB
	}
	if c.DefaultChunkSize == 0 {
		c.DefaultChunkSize = 50 * 1024 * 1024 // 50MB
	}
	if c.DefaultChunkSize < c.MinChunkSize {
		c.DefaultChunkSize = c.MinChunkSize
	}
	if c.DefaultChunkSize > c.MaxChunkSize {
		c.DefaultChunkSize = c.MaxChunkSize
	}
}

// Chunker defines the interface for splitting files into chunks.
type Chunker interface {
	ChunkFile(path string, chunkSize int64) ([]*models.ChunkMetadata, error)
	CalculateChunkHash(chunk []byte) [32]byte
}

type fileChunker struct {
	cfg ChunkerConfig
}

// NewChunker creates a new Chunker with the given config.
func NewChunker(cfg ChunkerConfig) Chunker {
	cfg.normalize()
	return &fileChunker{cfg: cfg}
}

// ChunkFile splits the file at path into chunks of up to chunkSize bytes.
// If chunkSize is <= 0, the DefaultChunkSize from config is used.
func (c *fileChunker) ChunkFile(path string, chunkSize int64) ([]*models.ChunkMetadata, error) {
	if chunkSize <= 0 {
		chunkSize = c.cfg.DefaultChunkSize
	}
	if chunkSize < c.cfg.MinChunkSize {
		chunkSize = c.cfg.MinChunkSize
	}
	if chunkSize > c.cfg.MaxChunkSize {
		chunkSize = c.cfg.MaxChunkSize
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(f)
	var (
		offset int64
		index  int
		result []*models.ChunkMetadata
		now    = time.Now()
	)

	buf := make([]byte, chunkSize)
	for {
		n, readErr := io.ReadFull(reader, buf)
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			if n == 0 {
				break
			}
		} else if readErr != nil {
			return nil, readErr
		}

		chunk := buf[:n]
		hash := c.CalculateChunkHash(chunk)

		meta := &models.ChunkMetadata{
			ID:         fmt.Sprintf("%d", index),
			Size:       int64(n),
			Offset:     offset,
			SHA256:     fmt.Sprintf("%x", hash[:]),
			IsParity:   false,
			Status:     models.ChunkStatusPending,
			CreatedAt:  now,
			UpdatedAt:  now,
			SessionID:  "",
			Priority:   0,
			RetryCount: 0,
		}
		result = append(result, meta)

		offset += int64(n)
		index++

		if offset >= info.Size() {
			break
		}

		if readErr == io.EOF {
			break
		}
	}

	return result, nil
}

// CalculateChunkHash computes the SHA-256 hash for a given chunk.
func (c *fileChunker) CalculateChunkHash(chunk []byte) [32]byte {
	return sha256.Sum256(chunk)
}


