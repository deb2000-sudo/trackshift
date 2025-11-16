package crypto

import (
	"bytes"
	"testing"
)

func TestCompressDecompressRoundTrip(t *testing.T) {
	data := bytes.Repeat([]byte{1, 2, 3, 4, 5}, 1024*10) // 50KB

	comp, err := CompressChunk(data)
	if err != nil {
		t.Fatalf("CompressChunk error: %v", err)
	}
	if len(comp) == 0 {
		t.Fatalf("expected compressed data, got empty slice")
	}

	decomp, err := DecompressChunk(comp)
	if err != nil {
		t.Fatalf("DecompressChunk error: %v", err)
	}

	if !bytes.Equal(data, decomp) {
		t.Fatalf("round-trip mismatch")
	}
}

func TestHashAndVerifyChunk(t *testing.T) {
	data := []byte("hello world")
	hash := HashChunk(data)

	if !VerifyChunk(data, hash) {
		t.Fatalf("expected hash verification to succeed")
	}

	if VerifyChunk([]byte("other"), hash) {
		t.Fatalf("expected hash verification to fail for different data")
	}
}

func BenchmarkCompressChunk(b *testing.B) {
	data := bytes.Repeat([]byte("TrackShift compression benchmark"), 1024)
	for i := 0; i < b.N; i++ {
		if _, err := CompressChunk(data); err != nil {
			b.Fatalf("CompressChunk error: %v", err)
		}
	}
}

func BenchmarkDecompressChunk(b *testing.B) {
	data := bytes.Repeat([]byte("TrackShift compression benchmark"), 1024)
	comp, err := CompressChunk(data)
	if err != nil {
		b.Fatalf("CompressChunk error: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DecompressChunk(comp); err != nil {
			b.Fatalf("DecompressChunk error: %v", err)
		}
	}
}


