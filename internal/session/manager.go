package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/deb2000-sudo/trackshift/pkg/models"
)

// SessionManager manages in-memory sessions and persists them to disk.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*models.TransferSession
	baseDir  string
}

// SessionCheckpoint is a lightweight snapshot of session progress.
type SessionCheckpoint struct {
	SessionID       string    `json:"session_id"`
	CompletedChunks []string  `json:"completed_chunks"`
	PendingChunks   []string  `json:"pending_chunks"`
	TotalChunks     int       `json:"total_chunks"`
	LastUpdateTime  time.Time `json:"last_update_time"`
}

// NewSessionManager creates a new SessionManager using baseDir for persistence.
// Existing session files in baseDir are loaded on startup.
func NewSessionManager(baseDir string) (*SessionManager, error) {
	if baseDir == "" {
		return nil, errors.New("baseDir must not be empty")
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating sessions dir: %w", err)
	}

	mgr := &SessionManager{
		sessions: make(map[string]*models.TransferSession),
		baseDir:  baseDir,
	}
	if err := mgr.loadExisting(); err != nil {
		return nil, err
	}
	return mgr, nil
}

// loadExisting loads any existing session JSON files from baseDir.
func (m *SessionManager) loadExisting() error {
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return fmt.Errorf("read sessions dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-len(".json")]
		s, err := m.LoadSession(id)
		if err != nil {
			// best-effort: log-style error via fmt, but continue
			fmt.Fprintf(os.Stderr, "failed to load session %s: %v\n", id, err)
			continue
		}
		m.sessions[id] = s
	}
	return nil
}

// CreateSession creates and persists a new transfer session.
func (m *SessionManager) CreateSession(fileInfo models.FileMetadata) (*models.TransferSession, error) {
	if err := fileInfo.Validate(); err != nil {
		return nil, err
	}
	id := uuid.NewString()
	now := time.Now()

	s := &models.TransferSession{
		ID:          id,
		File:        fileInfo,
		Status:      models.SessionStatusCreated,
		Chunks:      make(map[string]*models.ChunkMetadata),
		CreatedAt:   now,
		UpdatedAt:   now,
		TotalChunks: 0,
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.sessions[id] = s
	m.mu.Unlock()

	if err := m.SaveSession(s); err != nil {
		return nil, err
	}
	return s, nil
}

// GetSession returns a session by ID.
func (m *SessionManager) GetSession(id string) (*models.TransferSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %s not found", id)
	}
	return s, nil
}

// UpdateChunkStatus updates the status of a chunk in a session and persists the session.
func (m *SessionManager) UpdateChunkStatus(sessionID, chunkID string, status models.ChunkStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session %s not found", sessionID)
	}

	chunk, ok := s.Chunks[chunkID]
	if !ok {
		// lazily create metadata entry if not present
		chunk = &models.ChunkMetadata{
			ID:        chunkID,
			CreatedAt: time.Now(),
		}
		s.Chunks[chunkID] = chunk
	}

	chunk.Status = status
	chunk.UpdatedAt = time.Now()

	switch status {
	case models.ChunkStatusCompleted:
		s.Completed++
	case models.ChunkStatusFailed:
		s.Failed++
	}
	s.UpdatedAt = time.Now()

	return m.saveLocked(s)
}

// SaveSession persists the given session to disk.
func (m *SessionManager) SaveSession(session *models.TransferSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveLocked(session)
}

// saveLocked must be called with m.mu locked.
func (m *SessionManager) saveLocked(session *models.TransferSession) error {
	if err := session.Validate(); err != nil {
		return err
	}
	path := filepath.Join(m.baseDir, session.ID+".json")

	tmpPath := path + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open temp session file: %w", err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(session); err != nil {
		f.Close()
		return fmt.Errorf("encode session: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp session file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("atomic rename session file: %w", err)
	}
	return nil
}

// LoadSession loads a session from disk by ID.
func (m *SessionManager) LoadSession(id string) (*models.TransferSession, error) {
	path := filepath.Join(m.baseDir, id+".json")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open session file: %w", err)
	}
	defer f.Close()

	var s models.TransferSession
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, fmt.Errorf("decode session: %w", err)
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return &s, nil
}

// ListSessions returns all known sessions in memory.
func (m *SessionManager) ListSessions() []*models.TransferSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]*models.TransferSession, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, s)
	}
	return out
}

// PersistCheckpoint writes a checkpoint file for the given session.
func (m *SessionManager) PersistCheckpoint(sessionID string) error {
	m.mu.RLock()
	s, ok := m.sessions[sessionID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session %s not found", sessionID)
	}

	var completed, pending []string
	for id, ch := range s.Chunks {
		switch ch.Status {
		case models.ChunkStatusCompleted:
			completed = append(completed, id)
		default:
			pending = append(pending, id)
		}
	}

	cp := SessionCheckpoint{
		SessionID:       s.ID,
		CompletedChunks: completed,
		PendingChunks:   pending,
		TotalChunks:     s.TotalChunks,
		LastUpdateTime:  time.Now(),
	}

	path := filepath.Join(m.baseDir, s.ID+".checkpoint.json")
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(&cp); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// GetMissingChunks returns IDs of chunks that are not completed.
func (m *SessionManager) GetMissingChunks(sessionID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[sessionID]
	if !ok {
		return nil
	}
	var missing []string
	for id, ch := range s.Chunks {
		if ch.Status != models.ChunkStatusCompleted {
			missing = append(missing, id)
		}
	}
	return missing
}


