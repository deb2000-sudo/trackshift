package models

import (
	"errors"
	"time"
)

// ChunkStatus represents the lifecycle state of a single chunk in a transfer.
type ChunkStatus string

const (
	ChunkStatusPending    ChunkStatus = "pending"
	ChunkStatusInProgress ChunkStatus = "in_progress"
	ChunkStatusCompleted  ChunkStatus = "completed"
	ChunkStatusFailed     ChunkStatus = "failed"
)

// SessionStatus represents the lifecycle state of a transfer session.
type SessionStatus string

const (
	SessionStatusCreated      SessionStatus = "created"
	SessionStatusTransferring SessionStatus = "transferring"
	SessionStatusPaused       SessionStatus = "paused"
	SessionStatusCompleted    SessionStatus = "completed"
	SessionStatusFailed       SessionStatus = "failed"
)

// FileMetadata describes the file being transferred.
type FileMetadata struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Hash     string `json:"hash"`      // hex-encoded SHA-256 of full file
	MimeType string `json:"mime_type"` // optional, best-effort
}

// ChunkMetadata describes a single chunk of a file.
type ChunkMetadata struct {
	ID         string       `json:"id"`
	Size       int64        `json:"size"`
	Offset     int64        `json:"offset"`
	SHA256     string       `json:"sha256"`      // hex-encoded SHA-256 of the chunk
	IsParity   bool         `json:"is_parity"`   // true for parity chunks when erasure coding enabled
	Status     ChunkStatus  `json:"status"`      // current status of this chunk
	UpdatedAt  time.Time    `json:"updated_at"`  // last status change time
	CreatedAt  time.Time    `json:"created_at"`  // creation time
	SessionID  string       `json:"session_id"`  // owning session
	Priority   int          `json:"priority"`    // used by priority sender
	RetryCount int          `json:"retry_count"` // number of send retries
	Error      string       `json:"error"`       // last error, if any
}

// TransferSession tracks the state of a file transfer.
type TransferSession struct {
	ID            string                    `json:"id"`
	File          FileMetadata              `json:"file"`
	Status        SessionStatus             `json:"status"`
	Chunks        map[string]*ChunkMetadata `json:"chunks"`          // chunkID -> metadata
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
	CompletedAt   *time.Time                `json:"completed_at,omitempty"`
	TotalChunks   int                       `json:"total_chunks"`
	Completed     int                       `json:"completed"`
	Failed        int                       `json:"failed"`
	BytesSent     int64                     `json:"bytes_sent"`
	BytesReceived int64                     `json:"bytes_received"`
}

// Validate validates the FileMetadata.
func (f *FileMetadata) Validate() error {
	if f.Name == "" {
		return errors.New("file name must not be empty")
	}
	if f.Size <= 0 {
		return errors.New("file size must be greater than zero")
	}
	if f.Hash == "" {
		return errors.New("file hash must not be empty")
	}
	return nil
}

// Validate validates the ChunkMetadata.
func (c *ChunkMetadata) Validate() error {
	if c.ID == "" {
		return errors.New("chunk id must not be empty")
	}
	if c.Size <= 0 {
		return errors.New("chunk size must be greater than zero")
	}
	if c.Offset < 0 {
		return errors.New("chunk offset must be non-negative")
	}
	if c.SHA256 == "" {
		return errors.New("chunk sha256 must not be empty")
	}
	switch c.Status {
	case ChunkStatusPending, ChunkStatusInProgress, ChunkStatusCompleted, ChunkStatusFailed:
	default:
		return errors.New("invalid chunk status")
	}
	return nil
}

// Validate validates the TransferSession.
func (s *TransferSession) Validate() error {
	if s.ID == "" {
		return errors.New("session id must not be empty")
	}
	if err := s.File.Validate(); err != nil {
		return err
	}
	switch s.Status {
	case SessionStatusCreated,
		SessionStatusTransferring,
		SessionStatusPaused,
		SessionStatusCompleted,
		SessionStatusFailed:
	default:
		return errors.New("invalid session status")
	}
	if s.TotalChunks < 0 {
		return errors.New("total chunks must be non-negative")
	}
	if s.Completed < 0 || s.Failed < 0 {
		return errors.New("completed/failed counts must be non-negative")
	}
	return nil
}


