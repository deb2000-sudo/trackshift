package models

import "testing"

func TestFileMetadataValidate(t *testing.T) {
	f := FileMetadata{
		Name: "test.bin",
		Size: 1024,
		Hash: "abc",
	}
	if err := f.Validate(); err != nil {
		t.Fatalf("expected valid file metadata, got error: %v", err)
	}

	f.Name = ""
	if err := f.Validate(); err == nil {
		t.Fatalf("expected error for empty name")
	}
}

func TestChunkMetadataValidate(t *testing.T) {
	c := &ChunkMetadata{
		ID:     "chunk-1",
		Size:   1024,
		Offset: 0,
		SHA256: "abc",
		Status: ChunkStatusPending,
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected valid chunk, got error: %v", err)
	}

	c.Status = "bogus"
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error for invalid status")
	}
}

func TestTransferSessionValidate(t *testing.T) {
	s := &TransferSession{
		ID: "session-1",
		File: FileMetadata{
			Name: "test.bin",
			Size: 1024,
			Hash: "abc",
		},
		Status:      SessionStatusCreated,
		TotalChunks: 10,
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("expected valid session, got error: %v", err)
	}

	s.Status = "invalid"
	if err := s.Validate(); err == nil {
		t.Fatalf("expected error for invalid status")
	}
}


