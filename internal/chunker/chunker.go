package chunker

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/deb2000-sudo/trackshift/internal/telemetry"
	"github.com/deb2000-sudo/trackshift/pkg/models"
)

// ChunkerConfig controls how files are split into chunks.
type ChunkerConfig struct {
	MinChunkSize     int64
	MaxChunkSize     int64
	DefaultChunkSize int64

	// Telemetry provides live network stats used by the AI optimizer.
	// It is optional; if nil, the AI service will fall back to defaults.
	Telemetry *telemetry.TelemetryCollector
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

// clampSize ensures a given size respects Min/Max constraints.
func (c *ChunkerConfig) clampSize(size int64) int64 {
	c.normalize()
	if size <= 0 {
		size = c.DefaultChunkSize
	}
	if size < c.MinChunkSize {
		size = c.MinChunkSize
	}
	if size > c.MaxChunkSize {
		size = c.MaxChunkSize
	}
	return size
}

// ChooseChunkSizeStatic implements the legacy/static behavior:
// - use the override if > 0, otherwise DefaultChunkSize
// - then clamp to [MinChunkSize, MaxChunkSize].
func (c *ChunkerConfig) ChooseChunkSizeStatic(override int64) int64 {
	if override <= 0 {
		return c.clampSize(c.DefaultChunkSize)
	}
	return c.clampSize(override)
}

// ChooseChunkSizeAI applies a simple heuristic "AI" policy based on file size.
// This is where a learned model is plugged in.
// It first tries a Hugging Face model to predict an optimal chunk size (in MB)
// based on file metadata. If that call fails for any reason, it falls back
// to a simple heuristic:
// - small files -> smaller chunks (better feedback)
// - huge files  -> larger chunks (reduce overhead)
func (c *ChunkerConfig) ChooseChunkSizeAI(file models.FileMetadata) int64 {
	const (
		MB = 1024 * 1024
		GB = 1024 * 1024 * 1024
	)

	// First, try local ML optimizer service (XGBoost/LightGBM style).
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if predicted, err := predictChunkSizeWithService(ctx, file, c.Telemetry); err == nil && predicted > 0 {
		return c.clampSize(predicted)
	}

	// Next, try Hugging Face.
	if predicted, err := predictChunkSizeWithHF(ctx, file); err == nil && predicted > 0 {
		return c.clampSize(predicted)
	}

	// Fallback: original heuristic based on file size.
	size := file.Size
	var chosen int64

	switch {
	case size <= 100*MB:
		// small file: smaller chunks (8MB) to get quick feedback and progress
		chosen = 8 * MB
	case size <= 1*GB:
		// medium: balance between overhead and responsiveness
		chosen = 32 * MB
	case size <= 10*GB:
		// large files: moderately large chunks
		chosen = 64 * MB
	default:
		// very large: larger chunks to reduce number of round-trips
		chosen = 128 * MB
	}

	return c.clampSize(chosen)
}

// hfRequest represents the JSON payload sent to Hugging Face Inference API.
type hfRequest struct {
	Inputs     string                 `json:"inputs"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// hfResponseItem represents a single item returned by text-generation models.
type hfResponseItem struct {
	GeneratedText string `json:"generated_text"`
}

// predictChunkSizeWithHF calls a Hugging Face text-to-text model (google/flan-t5-small)
// to predict an optimal chunk size in bytes, based on file metadata.
// It expects an environment variable HF_API_TOKEN to be set with the user's key.
func predictChunkSizeWithHF(ctx context.Context, file models.FileMetadata) (int64, error) {
	apiToken := os.Getenv("HF_API_TOKEN")
	if apiToken == "" {
		// If the token is not set, signal the caller to fall back.
		return 0, fmt.Errorf("HF_API_TOKEN not set")
	}

	prompt := "You are a chunk size optimizer for file transfer.\n\n" +
		"Given this file:\n" +
		"- name: " + file.Name + "\n" +
		"- size_bytes: " + strconv.FormatInt(file.Size, 10) + "\n" +
		"- mime_type: " + file.MimeType + "\n\n" +
		"Suggest an optimal chunk size in megabytes as a plain integer (no units, no extra text)."

	reqBody := hfRequest{
		Inputs: prompt,
		Parameters: map[string]interface{}{
			"max_new_tokens": 8,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api-inference.huggingface.co/models/google/flan-t5-small",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return 0, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("huggingface returned status %s", resp.Status)
	}

	var hfResp []hfResponseItem
	if err := json.NewDecoder(resp.Body).Decode(&hfResp); err != nil {
		return 0, err
	}
	if len(hfResp) == 0 {
		return 0, fmt.Errorf("empty response from huggingface")
	}

	text := strings.TrimSpace(hfResp[0].GeneratedText)

	// Extract leading digits to get the integer megabyte value.
	var digits strings.Builder
	for _, r := range text {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		} else if digits.Len() > 0 {
			break
		}
	}

	if digits.Len() == 0 {
		return 0, fmt.Errorf("no integer found in model output: %q", text)
	}

	mb, err := strconv.ParseInt(digits.String(), 10, 64)
	if err != nil {
		return 0, err
	}

	const MB = 1024 * 1024
	return mb * MB, nil
}

// predictChunkSizeWithService calls the local optimizer service (implemented with
// XGBoost/LightGBM in Python) to obtain an optimal chunk size in bytes.
// The service is expected to listen on http://localhost:8000/predict-chunk-size.
func predictChunkSizeWithService(ctx context.Context, file models.FileMetadata, t *telemetry.TelemetryCollector) (int64, error) {
	type requestPayload struct {
		SizeBytes              int64   `json:"size_bytes"`
		MimeType               string  `json:"mime_type"`
		EstimatedBandwidthMbps float64 `json:"estimated_bandwidth_mbps"`
		LatencyMs              float64 `json:"latency_ms"`
	}

	type responsePayload struct {
		ChunkSizeMB float64 `json:"chunk_size_mb"`
	}

	reqBody := requestPayload{
		SizeBytes: file.Size,
		MimeType:  file.MimeType,
	}

	// Use telemetry metrics when available; otherwise leave as zero and let the
	// Python service apply its own defaults.
	if t != nil {
		reqBody.EstimatedBandwidthMbps = t.BandwidthMbps()
		reqBody.LatencyMs = t.LatencyMs()
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://localhost:8000/predict-chunk-size",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return 0, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("optimizer service returned status %s", resp.Status)
	}

	var parsed responsePayload
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, err
	}

	if parsed.ChunkSizeMB <= 0 {
		return 0, fmt.Errorf("invalid chunk_size_mb from service: %f", parsed.ChunkSizeMB)
	}

	const MB = 1024 * 1024
	return int64(parsed.ChunkSizeMB * MB), nil
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
	// Keep legacy behavior by clamping the provided size.
	c.cfg.normalize()
	chunkSize = c.cfg.clampSize(chunkSize)

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
