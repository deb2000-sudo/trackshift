package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// HashFileSHA256 returns the hex-encoded SHA-256 hash of a file at the given path.
func HashFileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// HashBytesSHA256 returns the hex-encoded SHA-256 hash of the given bytes.
func HashBytesSHA256(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// HumanBytes returns a human-readable representation of a byte count.
func HumanBytes(n int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	f := float64(n)
	switch {
	case f >= TB:
		return fmt.Sprintf("%.2fTB", f/TB)
	case f >= GB:
		return fmt.Sprintf("%.2fGB", f/GB)
	case f >= MB:
		return fmt.Sprintf("%.2fMB", f/MB)
	case f >= KB:
		return fmt.Sprintf("%.2fKB", f/KB)
	default:
		return fmt.Sprintf("%dB", n)
	}
}


