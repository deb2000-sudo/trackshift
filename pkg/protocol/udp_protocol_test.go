package protocol

import (
	"bytes"
	"testing"
)

func TestSerializeDeserializePacket(t *testing.T) {
	var sessID [16]byte
	copy(sessID[:], []byte("session-12345678"))

	p := &Packet{
		Version:   currentVer,
		Type:      PacketTypeData,
		SessionID: sessID,
		ChunkID:   42,
		Seq:       7,
		Priority:  1,
		Payload:   []byte("hello world"),
	}

	data, err := SerializePacket(p)
	if err != nil {
		t.Fatalf("SerializePacket error: %v", err)
	}

	got, err := DeserializePacket(data)
	if err != nil {
		t.Fatalf("DeserializePacket error: %v", err)
	}

	if got.Type != p.Type || got.ChunkID != p.ChunkID || got.Seq != p.Seq || got.Priority != p.Priority {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", got, p)
	}
	if !bytes.Equal(got.Payload, p.Payload) {
		t.Fatalf("payload mismatch")
	}
}

func TestChecksumVerificationFailure(t *testing.T) {
	p := &Packet{
		Version:  currentVer,
		Type:     PacketTypeData,
		Payload:  []byte("test"),
		Priority: 0,
	}

	data, err := SerializePacket(p)
	if err != nil {
		t.Fatalf("SerializePacket error: %v", err)
	}

	// Corrupt one byte in payload
	data[len(data)-checksumSize-1] ^= 0xFF

	if _, err := DeserializePacket(data); err == nil {
		t.Fatalf("expected checksum verification error")
	}
}


