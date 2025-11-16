package erasure

import (
	"fmt"

	rs "github.com/klauspost/reedsolomon"
)

// ErasureCoder wraps a Reed-Solomon encoder/decoder.
type ErasureCoder struct {
	DataShards   int
	ParityShards int
	ShardSize    int

	codec rs.Encoder
}

// NewErasureCoder creates a new ErasureCoder with the given shard configuration.
func NewErasureCoder(dataShards, parityShards int) (*ErasureCoder, error) {
	if dataShards <= 0 || parityShards <= 0 {
		return nil, fmt.Errorf("dataShards and parityShards must be > 0")
	}
	codec, err := rs.New(dataShards, parityShards)
	if err != nil {
		return nil, err
	}
	return &ErasureCoder{
		DataShards:   dataShards,
		ParityShards: parityShards,
		codec:        codec,
	}, nil
}

// CalculateShardSize returns a shard size that evenly splits dataSize across data shards.
func (e *ErasureCoder) CalculateShardSize(dataSize int64) int {
	if dataSize <= 0 {
		return 0
	}
	shards := int64(e.DataShards)
	size := (dataSize + shards - 1) / shards
	e.ShardSize = int(size)
	return e.ShardSize
}

// Encode splits data into data+parity shards.
func (e *ErasureCoder) Encode(data []byte) ([][]byte, error) {
	if e.DataShards == 0 {
		return nil, fmt.Errorf("erasure coder not initialized")
	}
	shardSize := e.ShardSize
	if shardSize == 0 {
		shardSize = e.CalculateShardSize(int64(len(data)))
	}
	totalShards := e.DataShards + e.ParityShards
	shards := make([][]byte, totalShards)

	// data shards
	for i := 0; i < e.DataShards; i++ {
		start := i * shardSize
		end := start + shardSize
		if start >= len(data) {
			shards[i] = make([]byte, shardSize)
			continue
		}
		if end > len(data) {
			end = len(data)
		}
		shard := make([]byte, shardSize)
		copy(shard, data[start:end])
		shards[i] = shard
	}
	// parity shards
	for i := e.DataShards; i < totalShards; i++ {
		shards[i] = make([]byte, shardSize)
	}

	if err := e.codec.Encode(shards); err != nil {
		return nil, err
	}
	return shards, nil
}

// Decode reconstructs the original data from shards (some of which may be nil).
func (e *ErasureCoder) Decode(shards [][]byte) ([]byte, error) {
	if len(shards) != e.DataShards+e.ParityShards {
		return nil, fmt.Errorf("expected %d shards, got %d", e.DataShards+e.ParityShards, len(shards))
	}
	if err := e.codec.Reconstruct(shards); err != nil {
		return nil, err
	}
	// join data shards and trim padding
	data := make([]byte, 0, len(shards[0])*e.DataShards)
	for i := 0; i < e.DataShards; i++ {
		data = append(data, shards[i]...)
	}
	return data, nil
}

// ValidateShards ensures all shards have equal length and at least DataShards are present.
func (e *ErasureCoder) ValidateShards(shards [][]byte) error {
	if len(shards) != e.DataShards+e.ParityShards {
		return fmt.Errorf("expected %d shards, got %d", e.DataShards+e.ParityShards, len(shards))
	}
	var shardLen int
	present := 0
	for i, sh := range shards {
		if sh == nil {
			continue
		}
		if shardLen == 0 {
			shardLen = len(sh)
		} else if len(sh) != shardLen {
			return fmt.Errorf("shard %d has inconsistent length", i)
		}
		present++
	}
	if present < e.DataShards {
		return fmt.Errorf("not enough shards present: have %d, need %d", present, e.DataShards)
	}
	return nil
}


