package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

// PacketType represents the type of UDP packet.
type PacketType uint8

const (
	PacketTypeData    PacketType = 0x01
	PacketTypeAck     PacketType = 0x02
	PacketTypeNack    PacketType = 0x03
	PacketTypeControl PacketType = 0x04
)

// Packet represents a TrackShift UDP packet.
//
// Header layout (not exported directly, used by serialization):
//   Magic       [4]byte  // "TSFT"
//   Version     uint8
//   Type        uint8
//   SessionID   [16]byte // UUID
//   ChunkID     uint64
//   Seq         uint32
//   Priority    uint8
//   _pad        [3]byte  // padding for alignment / future use
//   Payload     []byte   // up to 64KB
//   Checksum    uint32   // CRC32 over header+payload (checksum field zeroed)
type Packet struct {
	Version   uint8
	Type      PacketType
	SessionID [16]byte
	ChunkID   uint64
	Seq       uint32
	Priority  uint8
	Payload   []byte
	Checksum  uint32
}

var magic = [4]byte{'T', 'S', 'F', 'T'}

const (
	headerSize   = 4 + 1 + 1 + 16 + 8 + 4 + 1 + 3 // 38 bytes
	maxPayload   = 64 * 1024
	currentVer   = 1
	checksumSize = 4
)

// SerializePacket serializes a Packet into bytes suitable for UDP transport.
func SerializePacket(p *Packet) ([]byte, error) {
	if len(p.Payload) > maxPayload {
		return nil, errors.New("payload too large")
	}

	buf := bytes.NewBuffer(make([]byte, 0, headerSize+len(p.Payload)+checksumSize))

	// header
	if _, err := buf.Write(magic[:]); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(p.Version); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(byte(p.Type)); err != nil {
		return nil, err
	}
	if _, err := buf.Write(p.SessionID[:]); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, p.ChunkID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, p.Seq); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(p.Priority); err != nil {
		return nil, err
	}
	// padding
	if _, err := buf.Write([]byte{0, 0, 0}); err != nil {
		return nil, err
	}

	// payload
	if _, err := buf.Write(p.Payload); err != nil {
		return nil, err
	}

	// checksum over header + payload
	checksum := crc32.ChecksumIEEE(buf.Bytes())
	p.Checksum = checksum

	if err := binary.Write(buf, binary.BigEndian, p.Checksum); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializePacket parses bytes into a Packet and verifies checksum.
func DeserializePacket(data []byte) (*Packet, error) {
	if len(data) < headerSize+checksumSize {
		return nil, errors.New("packet too small")
	}

	buf := bytes.NewReader(data)

	var readMagic [4]byte
	if _, err := buf.Read(readMagic[:]); err != nil {
		return nil, err
	}
	if readMagic != magic {
		return nil, errors.New("invalid magic")
	}

	var version uint8
	if err := binary.Read(buf, binary.BigEndian, &version); err != nil {
		return nil, err
	}

	var t uint8
	if err := binary.Read(buf, binary.BigEndian, &t); err != nil {
		return nil, err
	}

	var sessionID [16]byte
	if _, err := buf.Read(sessionID[:]); err != nil {
		return nil, err
	}

	var chunkID uint64
	if err := binary.Read(buf, binary.BigEndian, &chunkID); err != nil {
		return nil, err
	}

	var seq uint32
	if err := binary.Read(buf, binary.BigEndian, &seq); err != nil {
		return nil, err
	}

	priorityByte, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	// skip padding
	if _, err := buf.Seek(3, io.SeekCurrent); err != nil {
		return nil, err
	}

	// remaining bytes in raw slice: payload + checksum
	remaining := data[headerSize:]
	if len(remaining) < checksumSize {
		return nil, errors.New("packet missing checksum")
	}
	payload := remaining[:len(remaining)-checksumSize]
	checksum := binary.BigEndian.Uint32(remaining[len(remaining)-checksumSize:])

	p := &Packet{
		Version:   version,
		Type:      PacketType(t),
		SessionID: sessionID,
		ChunkID:   chunkID,
		Seq:       seq,
		Priority:  priorityByte,
		Payload:   payload,
		Checksum:  checksum,
	}

	if !VerifyChecksum(data, checksum) {
		return nil, errors.New("checksum verification failed")
	}

	return p, nil
}

// CalculateChecksum computes CRC32 checksum of the given data.
func CalculateChecksum(data []byte) uint32 {
	if len(data) <= checksumSize {
		return crc32.ChecksumIEEE(data)
	}
	// Exclude trailing checksum bytes.
	return crc32.ChecksumIEEE(data[:len(data)-checksumSize])
}

// VerifyChecksum verifies that the checksum matches the data.
func VerifyChecksum(data []byte, checksum uint32) bool {
	return CalculateChecksum(data) == checksum
}


