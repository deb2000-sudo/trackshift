package transport

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"

	"github.com/deb2000-sudo/trackshift/internal/crypto"
	"github.com/deb2000-sudo/trackshift/pkg/models"
)

// TCPReceiver receives chunks and writes them to temporary storage, then can
// assemble them into a final file.
type TCPReceiver struct {
	OutputDir string
	TempDir   string
}

// NewTCPReceiver creates a receiver with the specified output and temp directories.
func NewTCPReceiver(outputDir, tempDir string) (*TCPReceiver, error) {
	if outputDir == "" {
		return nil, fmt.Errorf("outputDir must not be empty")
	}
	if tempDir == "" {
		tempDir = filepath.Join(outputDir, "temp")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return nil, err
	}
	return &TCPReceiver{
		OutputDir: outputDir,
		TempDir:   tempDir,
	}, nil
}

// Receive reads a single framed chunk from conn.
// Returns decompressed chunk data and its metadata.
func (r *TCPReceiver) Receive(conn net.Conn) ([]byte, *models.ChunkMetadata, error) {
	var metaLen uint32
	if err := binary.Read(conn, binary.BigEndian, &metaLen); err != nil {
		// Treat clean connection close as io.EOF so callers can stop without logging an error.
		if err == io.EOF {
			return nil, nil, io.EOF
		}
		return nil, nil, fmt.Errorf("read meta length: %w", err)
	}
	metaBytes := make([]byte, metaLen)
	if _, err := io.ReadFull(conn, metaBytes); err != nil {
		return nil, nil, fmt.Errorf("read meta: %w", err)
	}

	var meta models.ChunkMetadata
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return nil, nil, fmt.Errorf("unmarshal metadata: %w", err)
	}

	var dataLen uint64
	if err := binary.Read(conn, binary.BigEndian, &dataLen); err != nil {
		return nil, nil, fmt.Errorf("read data length: %w", err)
	}

	data := make([]byte, dataLen)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, nil, fmt.Errorf("read data: %w", err)
	}

	decompressed, err := crypto.DecompressChunk(data)
	if err != nil {
		return nil, nil, fmt.Errorf("decompress chunk: %w", err)
	}

	return decompressed, &meta, nil
}

// StoreChunk writes the chunk data to a temp file.
func (r *TCPReceiver) StoreChunk(sessionID string, meta *models.ChunkMetadata, data []byte) (string, error) {
	filename := fmt.Sprintf("%s_%s.part", sessionID, meta.ID)
	path := filepath.Join(r.TempDir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write chunk file: %w", err)
	}
	return path, nil
}

// AssembleFile joins all chunk files into the final output file ordered by offset.
func (r *TCPReceiver) AssembleFile(session *models.TransferSession) (string, error) {
	outPath := filepath.Join(r.OutputDir, session.File.Name)
	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("open output file: %w", err)
	}
	defer out.Close()

	// sort chunks by offset
	chunks := make([]*models.ChunkMetadata, 0, len(session.Chunks))
	for _, c := range session.Chunks {
		chunks = append(chunks, c)
	}
	sort.Slice(chunks, func(i, j int) bool { return chunks[i].Offset < chunks[j].Offset })

	for _, c := range chunks {
		filename := fmt.Sprintf("%s_%s.part", session.ID, c.ID)
		path := filepath.Join(r.TempDir, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read chunk file %s: %w", path, err)
		}
		if _, err := out.Write(data); err != nil {
			return "", fmt.Errorf("write output: %w", err)
		}
	}

	return outPath, nil
}


