package transport

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/deb2000-sudo/trackshift/pkg/models"
)

// TCPSender sends chunks and associated metadata over a TCP connection.
type TCPSender struct {
	DialTimeout time.Duration
}

// NewTCPSender creates a new TCPSender with sane defaults.
func NewTCPSender() *TCPSender {
	return &TCPSender{
		DialTimeout: 10 * time.Second,
	}
}

// Connect establishes a TCP connection to the given address.
func (s *TCPSender) Connect(address string) (net.Conn, error) {
	d := net.Dialer{Timeout: s.DialTimeout}
	conn, err := d.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("dial tcp %s: %w", address, err)
	}
	return conn, nil
}

// Send sends a single chunk with its metadata over an existing connection.
// Wire format:
//   [4 bytes metadata length][metadata JSON][8 bytes data length][data bytes]
func (s *TCPSender) Send(conn net.Conn, chunk []byte, metadata *models.ChunkMetadata) error {
	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	var buf bytes.Buffer

	// metadata length
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(metaBytes))); err != nil {
		return fmt.Errorf("write meta length: %w", err)
	}
	// metadata
	if _, err := buf.Write(metaBytes); err != nil {
		return fmt.Errorf("write meta: %w", err)
	}
	// data length
	if err := binary.Write(&buf, binary.BigEndian, uint64(len(chunk))); err != nil {
		return fmt.Errorf("write data length: %w", err)
	}
	// data
	if _, err := buf.Write(chunk); err != nil {
		return fmt.Errorf("write data: %w", err)
	}

	if _, err := conn.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("send frame: %w", err)
	}

	return nil
}


