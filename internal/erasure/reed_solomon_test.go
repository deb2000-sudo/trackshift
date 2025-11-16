package erasure

import "testing"

func TestEncodeDecodeRoundTrip(t *testing.T) {
	ec, err := NewErasureCoder(10, 3)
	if err != nil {
		t.Fatalf("NewErasureCoder: %v", err)
	}

	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 251)
	}

	shards, err := ec.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// lose up to 3 shards
	shards[2] = nil
	shards[5] = nil
	shards[9] = nil

	if err := ec.ValidateShards(shards); err != nil {
		t.Fatalf("ValidateShards: %v", err)
	}

	recovered, err := ec.Decode(shards)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if len(recovered) < len(data) {
		t.Fatalf("recovered size too small: %d < %d", len(recovered), len(data))
	}
	for i := range data {
		if data[i] != recovered[i] {
			t.Fatalf("data mismatch at %d", i)
		}
	}
}


