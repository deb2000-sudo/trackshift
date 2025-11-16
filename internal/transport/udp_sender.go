package transport

import (
	"net"
	"sync"
	"time"

	"github.com/deb2000-sudo/trackshift/pkg/models"
	"github.com/deb2000-sudo/trackshift/pkg/protocol"
)

// UDPSenderConfig configures the UDP sender behaviour.
type UDPSenderConfig struct {
	RemoteAddr        string
	MaxParallelStreams int
	RetransmitTimeout  time.Duration
	MaxRetries         int
	WindowSize         int
}

// TransferStats holds simple statistics about a transfer.
type TransferStats struct {
	Sent         uint64
	Acked        uint64
	Retransmits  uint64
	LastRTT      time.Duration
}

// UDPSender implements a basic sliding-window UDP sender.
// This is intentionally conservative for the first implementation and can be
// extended with full RTT-based congestion control later.
type UDPSender struct {
	cfg   UDPSenderConfig
	conn  *net.UDPConn

	mu    sync.RWMutex
	stats TransferStats

	seqMu sync.Mutex
	seq   uint32
}

// NewUDPSender creates a new UDPSender with the given config.
func NewUDPSender(cfg UDPSenderConfig) (*UDPSender, error) {
	if cfg.MaxParallelStreams <= 0 {
		cfg.MaxParallelStreams = 32
	}
	if cfg.RetransmitTimeout == 0 {
		cfg.RetransmitTimeout = 200 * time.Millisecond
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 5
	}
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = 256
	}

	raddr, err := net.ResolveUDPAddr("udp", cfg.RemoteAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	s := &UDPSender{
		cfg:  cfg,
		conn: conn,
	}
	return s, nil
}

// nextSeq atomically increments and returns the next sequence number.
func (s *UDPSender) nextSeq() uint32 {
	s.seqMu.Lock()
	defer s.seqMu.Unlock()
	s.seq++
	return s.seq
}

// Close closes the underlying UDP connection.
func (s *UDPSender) Close() error {
	return s.conn.Close()
}

// SendChunk sends a single chunk as a DATA packet.
// For now this is a simple fire-and-forget send; higher-level reliability will
// be handled by erasure coding and retry logic in later phases.
func (s *UDPSender) SendChunk(sessionID [16]byte, chunkID uint64, data []byte, priority uint8) error {
	seq := s.nextSeq()
	p := &protocol.Packet{
		Version:   1,
		Type:      protocol.PacketTypeData,
		SessionID: sessionID,
		ChunkID:   chunkID,
		Seq:       seq,
		Priority:  priority,
		Payload:   data,
	}
	raw, err := protocol.SerializePacket(p)
	if err != nil {
		return err
	}

	n, err := s.conn.Write(raw)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.stats.Sent += uint64(n)
	s.mu.Unlock()
	return nil
}

// GetStats returns a snapshot of current stats.
func (s *UDPSender) GetStats() TransferStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

// ChunkPriority determines a basic priority based on metadata.
// Lower value means higher priority.
func ChunkPriority(meta *models.ChunkMetadata) uint8 {
	if meta.IsParity {
		return 4
	}
	if meta.Offset == 0 {
		return 2
	}
	return 3
}


