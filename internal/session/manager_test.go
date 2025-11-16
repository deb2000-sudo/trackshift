package session

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/deb2000-sudo/trackshift/pkg/models"
)

func newTempManager(t *testing.T) *SessionManager {
	t.Helper()
	dir := t.TempDir()
	mgr, err := NewSessionManager(dir)
	if err != nil {
		t.Fatalf("NewSessionManager: %v", err)
	}
	return mgr
}

func TestCreateAndGetSession(t *testing.T) {
	mgr := newTempManager(t)

	file := models.FileMetadata{
		Name: "test.bin",
		Size: 1024,
		Hash: "abc",
	}

	s, err := mgr.CreateSession(file)
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}

	got, err := mgr.GetSession(s.ID)
	if err != nil {
		t.Fatalf("GetSession error: %v", err)
	}
	if got.ID != s.ID {
		t.Fatalf("expected ID %s, got %s", s.ID, got.ID)
	}
}

func TestUpdateChunkStatusAndPersistence(t *testing.T) {
	dir := t.TempDir()
	mgr, err := NewSessionManager(dir)
	if err != nil {
		t.Fatalf("NewSessionManager: %v", err)
	}

	file := models.FileMetadata{
		Name: "test.bin",
		Size: 1024,
		Hash: "abc",
	}
	s, err := mgr.CreateSession(file)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if err := mgr.UpdateChunkStatus(s.ID, "chunk-1", models.ChunkStatusCompleted); err != nil {
		t.Fatalf("UpdateChunkStatus: %v", err)
	}

	// Ensure file exists on disk
	path := filepath.Join(dir, s.ID+".json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected session file to exist: %v", err)
	}

	// New manager should load existing session
	mgr2, err := NewSessionManager(dir)
	if err != nil {
		t.Fatalf("NewSessionManager 2: %v", err)
	}

	s2, err := mgr2.GetSession(s.ID)
	if err != nil {
		t.Fatalf("GetSession from mgr2: %v", err)
	}
	if s2.Completed == 0 {
		t.Fatalf("expected completed > 0, got %d", s2.Completed)
	}
}

func TestConcurrentAccess(t *testing.T) {
	mgr := newTempManager(t)

	file := models.FileMetadata{
		Name: "test.bin",
		Size: 1024,
		Hash: "abc",
	}
	s, err := mgr.CreateSession(file)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			chunkID := "chunk-concurrent"
			_ = mgr.UpdateChunkStatus(s.ID, chunkID, models.ChunkStatusInProgress)
		}(i)
	}

	wg.Wait()
}


