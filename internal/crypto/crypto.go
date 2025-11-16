package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/klauspost/compress/zstd"
)

// CompressChunk compresses the given data using zstd with a default level.
func CompressChunk(data []byte) ([]byte, error) {
	enc, err := zstd.NewWriter(nil)
	if err != nil {
		return nil, fmt.Errorf("create zstd encoder: %w", err)
	}
	defer enc.Close()

	out := enc.EncodeAll(data, nil)
	return out, nil
}

// DecompressChunk decompresses zstd-compressed data.
func DecompressChunk(data []byte) ([]byte, error) {
	dec, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("create zstd decoder: %w", err)
	}
	defer dec.Close()

	out, err := dec.DecodeAll(data, nil)
	if err != nil {
		return nil, fmt.Errorf("zstd decode: %w", err)
	}
	return out, nil
}

// HashChunk returns the SHA-256 hash of data as a fixed array.
func HashChunk(data []byte) [32]byte {
	return sha256.Sum256(data)
}

// VerifyChunk hashes data and compares it to expectedHash.
func VerifyChunk(data []byte, expectedHash [32]byte) bool {
	actual := HashChunk(data)
	return actual == expectedHash
}


