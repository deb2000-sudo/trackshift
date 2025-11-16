package orchestrator

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/deb2000-sudo/trackshift/pkg/models"
)

// Service implements a minimal in-memory orchestrator.
type Service struct {
	mu       sync.RWMutex
	sessions map[string]*models.TransferSession
	relays   map[string]*RelayInfo
}

// RelayInfo holds basic information about a registered relay.
type RelayInfo struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	Region    string    `json:"region,omitempty"`
	LastSeen  time.Time `json:"last_seen"`
}

// NewService creates a new orchestrator Service.
func NewService() *Service {
	return &Service{
		sessions: make(map[string]*models.TransferSession),
		relays:   make(map[string]*RelayInfo),
	}
}

// RegisterRoutes registers HTTP handlers on the given mux.
func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/session", s.handleSessionCreate)
	mux.HandleFunc("/api/v1/session/", s.handleSessionGet)
	mux.HandleFunc("/api/v1/relays/register", s.handleRelayRegister)
	mux.HandleFunc("/api/v1/relays", s.handleRelaysList)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

// handleSessionCreate handles POST /api/v1/session
func (s *Service) handleSessionCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		File models.FileMetadata `json:"file"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := req.File.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := uuid.NewString()
	now := time.Now()
	sess := &models.TransferSession{
		ID:        id,
		File:      req.File,
		Status:    models.SessionStatusCreated,
		Chunks:    make(map[string]*models.ChunkMetadata),
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.mu.Lock()
	s.sessions[id] = sess
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, sess)
}

// handleSessionGet handles GET /api/v1/session/:id
func (s *Service) handleSessionGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// url path: /api/v1/session/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/session/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := parts[0]

	s.mu.RLock()
	sess, ok := s.sessions[id]
	s.mu.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

// handleRelayRegister handles POST /api/v1/relays/register
func (s *Service) handleRelayRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID      string `json:"id"`
		Address string `json:"address"`
		Region  string `json:"region,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.ID == "" || req.Address == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	info := &RelayInfo{
		ID:       req.ID,
		Address:  req.Address,
		Region:   req.Region,
		LastSeen: time.Now(),
	}

	s.mu.Lock()
	s.relays[req.ID] = info
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, info)
}

// handleRelaysList handles GET /api/v1/relays
func (s *Service) handleRelaysList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	s.mu.RLock()
	out := make([]*RelayInfo, 0, len(s.relays))
	for _, v := range s.relays {
		out = append(out, v)
	}
	s.mu.RUnlock()
	writeJSON(w, http.StatusOK, out)
}


