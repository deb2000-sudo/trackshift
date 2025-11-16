package chunker

import (
	"os"
	"testing"
)

func writeTempFile(t *testing.T, size int64) string {
	t.Helper()

	f, err := os.CreateTemp("", "chunker_test_*.bin")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close()

	buf := make([]byte, 1024*1024) // 1MB buffer
	var written int64
	for written < size {
		n := size - written
		if n > int64(len(buf)) {
			n = int64(len(buf))
		}
		if _, err := f.Write(buf[:n]); err != nil {
			t.Fatalf("write: %v", err)
		}
		written += n
	}

	return f.Name()
}

func TestChunkFileBasic(t *testing.T) {
	// 10MB file, 5MB chunk size -> expect 2 chunks
	filePath := writeTempFile(t, 10*1024*1024)
	defer os.Remove(filePath)

	c := NewChunker(ChunkerConfig{})
	chunks, err := c.ChunkFile(filePath, 5*1024*1024)
	if err != nil {
		t.Fatalf("ChunkFile error: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}

	if chunks[0].Offset != 0 {
		t.Fatalf("expected first chunk offset 0, got %d", chunks[0].Offset)
	}
	if chunks[1].Offset != chunks[0].Size {
		t.Fatalf("expected second chunk offset %d, got %d", chunks[0].Size, chunks[1].Offset)
	}
}


