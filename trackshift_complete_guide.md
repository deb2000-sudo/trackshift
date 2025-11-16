# TrackShift: Complete Development Guide for Cursor AI
## Ubuntu 24.04 | Step-by-Step from Zero to Production

---

## 🎯 PHASE 0: Environment Setup (Day 1)

### Step 1: Install Required Tools

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Go (1.21+)
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc
go version

# Install Docker
sudo apt install -y docker.io docker-compose
sudo usermod -aG docker $USER
newgrp docker

# Install build essentials
sudo apt install -y build-essential git curl wget

# Install network testing tools
sudo apt install -y iperf3 tcpdump wireshark
```

### Step 2: Create Project Structure

```bash
# Create project directory
mkdir -p ~/projects/trackshift
cd ~/projects/trackshift

# Initialize Git
git init
git branch -M main

# Create Go module
go mod init github.com/yourusername/trackshift

# Create directory structure
mkdir -p {cmd,internal,pkg,api,configs,scripts,test,docs}
mkdir -p cmd/{sender,receiver,relay,orchestrator}
mkdir -p internal/{chunker,transport,session,crypto}
mkdir -p pkg/{protocol,utils,models}
mkdir -p test/{integration,performance}
```

### Step 3: Create Initial Files

Create `.gitignore`:
```bash
cat > .gitignore << 'EOF'
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test files
*.test
*.out
coverage.txt

# Go
vendor/
go.work

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Temp files
tmp/
*.tmp
*.log
EOF
```

Create `README.md`:
```bash
cat > README.md << 'EOF'
# TrackShift - High-Speed Resilient File Transfer

Fast, resilient file mover for unstable networks (50-100GB files).

## Quick Start
```bash
# Build all components
make build

# Run sender
./bin/sender --file /path/to/large/file.bin --receiver 192.168.1.100:8080

# Run receiver
./bin/receiver --port 8080 --output /path/to/destination/
```

## Architecture
- Dynamic Chunk Swarm Protocol (DCSP)
- UDP/QUIC transport with erasure coding
- AI-driven adaptive chunking
- Edge relay grid for global distribution

## Development Status
- [ ] Phase 1: Core MVP (In Progress)
- [ ] Phase 2: UDP Transport
- [ ] Phase 3: Edge Grid
- [ ] Phase 4: Resilience
- [ ] Phase 5: AI Optimization
- [ ] Phase 6: Production Hardening
EOF
```

Create `Makefile`:
```bash
cat > Makefile << 'EOF'
.PHONY: build test clean run-sender run-receiver

# Build all binaries
build:
	go build -o bin/sender ./cmd/sender
	go build -o bin/receiver ./cmd/receiver
	go build -o bin/relay ./cmd/relay

# Run tests
test:
	go test -v ./...

# Run with race detector
test-race:
	go test -race -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Run sender (example)
run-sender:
	./bin/sender --config configs/sender.yaml

# Run receiver (example)
run-receiver:
	./bin/receiver --config configs/receiver.yaml

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run
EOF
```

### Step 4: Open in Cursor

```bash
# Open the project in Cursor
cursor ~/projects/trackshift
```

---

## 🚀 PHASE 1: Core MVP (Week 1-2)

### Day 1-2: Project Models & Data Structures

**In Cursor, create this prompt:**

```
I'm building TrackShift, a high-speed file transfer system. Create the core data models in Go.

Requirements:
1. File: pkg/models/models.go
2. Define these structs with proper tags and documentation:
   - FileMetadata (name, size, hash, mime type)
   - ChunkMetadata (ID, size, offset, SHA256, parity flag)
   - TransferSession (session ID, file info, chunk map, status)
   - ChunkStatus enum (Pending, InProgress, Completed, Failed)
   - SessionStatus enum (Created, Transferring, Paused, Completed, Failed)

3. Include JSON tags for serialization
4. Add validation methods
5. Use proper Go conventions and documentation

Generate complete, production-ready code.
```

**Expected Output:** Cursor will generate `pkg/models/models.go`

**Verify it works:**
```bash
go build ./pkg/models
```

---

### Day 3-4: File Chunker Implementation

**Cursor Prompt:**

```
Create the file chunker for TrackShift in internal/chunker/chunker.go

Requirements:
1. ChunkerConfig struct with:
   - MinChunkSize (5MB default)
   - MaxChunkSize (200MB default)
   - DefaultChunkSize (50MB)
   
2. Chunker interface with methods:
   - ChunkFile(filepath string, chunkSize int64) ([]*ChunkMetadata, error)
   - CalculateChunkHash(chunk []byte) [32]byte
   
3. Implementation that:
   - Opens file and reads in chunks
   - Calculates SHA-256 for each chunk
   - Creates ChunkMetadata for each
   - Handles errors gracefully
   - Uses buffered I/O for efficiency

4. Include comprehensive error handling
5. Add unit tests in chunker_test.go

Use the models from pkg/models/models.go
```

**After Cursor generates, create a test file:**

```bash
# Create a 100MB test file
dd if=/dev/urandom of=test_files/test_100mb.bin bs=1M count=100
```

**Test the chunker:**
```bash
go test -v ./internal/chunker
```

---

### Day 5-6: Session Manager

**Cursor Prompt:**

```
Create session manager for TrackShift in internal/session/manager.go

Requirements:
1. SessionManager struct that:
   - Stores active sessions in memory (map)
   - Persists session state to disk (JSON files in sessions/ directory)
   - Uses sync.RWMutex for thread safety

2. Methods:
   - CreateSession(fileInfo FileMetadata) (*TransferSession, error)
   - GetSession(sessionID string) (*TransferSession, error)
   - UpdateChunkStatus(sessionID, chunkID string, status ChunkStatus) error
   - SaveSession(session *TransferSession) error
   - LoadSession(sessionID string) (*TransferSession, error)
   - ListSessions() []*TransferSession

3. Features:
   - Auto-generate UUID for session IDs
   - Atomic file writes for persistence
   - Session recovery on restart
   - Cleanup old sessions (configurable TTL)

4. Include tests for:
   - Session creation and retrieval
   - Concurrent access
   - Persistence across restarts

Use models from pkg/models/models.go
```

**Test it:**
```bash
go test -v ./internal/session
```

---

### Day 7-8: Basic Compression & Hashing

**Cursor Prompt:**

```
Create compression and hashing utilities for TrackShift in internal/crypto/crypto.go

Requirements:
1. CompressChunk(data []byte) ([]byte, error)
   - Use zstd compression (github.com/klauspost/compress/zstd)
   - Compression level: Default (balance speed/ratio)

2. DecompressChunk(data []byte) ([]byte, error)
   - Decompress zstd data
   - Validate output

3. HashChunk(data []byte) [32]byte
   - SHA-256 hash
   - Return as [32]byte array

4. VerifyChunk(data []byte, expectedHash [32]byte) bool
   - Calculate hash and compare

5. Error handling for all operations
6. Benchmark tests to measure compression ratio and speed

Add dependency:
go get github.com/klauspost/compress/zstd
```

**Run benchmarks:**
```bash
go test -bench=. -benchmem ./internal/crypto
```

---

### Day 9-10: Basic TCP Sender (MVP)

**Cursor Prompt:**

```
Create basic TCP sender for TrackShift MVP in cmd/sender/main.go

Requirements:
1. Command-line flags:
   - --file (input file path)
   - --receiver (receiver address:port)
   - --chunk-size (default 50MB)
   - --output-dir (session state directory)

2. Main flow:
   a. Parse flags
   b. Create session using SessionManager
   c. Chunk file using Chunker
   d. For each chunk:
      - Compress chunk
      - Calculate hash
      - Send over TCP to receiver
      - Update session status
   e. Send completion signal
   f. Print transfer statistics

3. Features:
   - Progress bar (use github.com/schollz/progressbar/v3)
   - Transfer speed calculation
   - ETA estimation
   - Graceful shutdown on Ctrl+C
   - Session state persistence for resume

4. Error handling and logging (use log/slog)

Also create internal/transport/tcp_sender.go with:
- TCPSender struct
- Send(conn net.Conn, chunk []byte, metadata ChunkMetadata) error
- Connect(address string) (net.Conn, error)

Dependencies:
go get github.com/schollz/progressbar/v3
go get github.com/google/uuid
```

---

### Day 11-12: Basic TCP Receiver (MVP)

**Cursor Prompt:**

```
Create TCP receiver for TrackShift in cmd/receiver/main.go

Requirements:
1. Command-line flags:
   - --port (listening port, default 8080)
   - --output-dir (where to save received files)
   - --temp-dir (temporary chunk storage)

2. Main flow:
   a. Start TCP listener
   b. Accept connections
   c. For each chunk received:
      - Decompress
      - Verify hash
      - Store temporarily
   d. When all chunks received:
      - Assemble file
      - Verify final file hash
      - Move to output directory
      - Cleanup temp files

3. Features:
   - Handle multiple concurrent transfers
   - Progress indication
   - Chunk validation
   - Resume capability (check for existing chunks)
   - Graceful shutdown

Also create internal/transport/tcp_receiver.go with:
- TCPReceiver struct
- Receive(conn net.Conn) ([]byte, ChunkMetadata, error)
- AssembleFile(chunks []*ChunkMetadata, outputPath string) error

Use SessionManager to track incoming transfers
```

---

### Day 13-14: Integration Testing & MVP Validation

**Create integration test script:**

```bash
cat > test/integration/test_basic_transfer.sh << 'EOF'
#!/bin/bash
set -e

echo "TrackShift MVP Integration Test"

# Cleanup
rm -rf /tmp/trackshift-test
mkdir -p /tmp/trackshift-test/{sender,receiver,sessions}

# Create test file (500MB)
echo "Creating 500MB test file..."
dd if=/dev/urandom of=/tmp/trackshift-test/sender/testfile.bin bs=1M count=500

# Start receiver in background
echo "Starting receiver..."
./bin/receiver --port 9090 \
  --output-dir /tmp/trackshift-test/receiver \
  --temp-dir /tmp/trackshift-test/receiver/temp &
RECEIVER_PID=$!

sleep 2

# Start sender
echo "Starting sender..."
./bin/sender --file /tmp/trackshift-test/sender/testfile.bin \
  --receiver localhost:9090 \
  --chunk-size 52428800 \
  --output-dir /tmp/trackshift-test/sessions

# Wait for transfer to complete
wait

# Verify file integrity
echo "Verifying transfer..."
SENDER_HASH=$(sha256sum /tmp/trackshift-test/sender/testfile.bin | cut -d' ' -f1)
RECEIVER_HASH=$(sha256sum /tmp/trackshift-test/receiver/testfile.bin | cut -d' ' -f1)

if [ "$SENDER_HASH" = "$RECEIVER_HASH" ]; then
  echo "✓ Transfer successful! Hashes match."
  echo "  Sender:   $SENDER_HASH"
  echo "  Receiver: $RECEIVER_HASH"
else
  echo "✗ Transfer failed! Hash mismatch."
  exit 1
fi

# Cleanup
kill $RECEIVER_PID 2>/dev/null || true
EOF

chmod +x test/integration/test_basic_transfer.sh
```

**Run the test:**
```bash
make build
./test/integration/test_basic_transfer.sh
```

**✅ PHASE 1 COMPLETE CHECKPOINT:**
- [ ] Can transfer 500MB+ files over TCP
- [ ] Chunk integrity verified with SHA-256
- [ ] Session state persisted
- [ ] Basic progress indication works
- [ ] Files match byte-for-byte

---

## 🔥 PHASE 2: UDP Transport & Parallelism (Week 3-4)

### Day 15-16: UDP Protocol Design

**Cursor Prompt:**

```
Design custom UDP protocol for TrackShift in pkg/protocol/udp_protocol.go

Requirements:
1. Define packet structure:
   - Header (32 bytes):
     * Magic number (4 bytes): 0xTRKSHIFT
     * Version (1 byte)
     * Packet type (1 byte): DATA, ACK, NACK, CONTROL
     * Session ID (16 bytes): UUID
     * Chunk ID (8 bytes)
     * Sequence number (4 bytes)
   - Payload (variable, max 64KB)
   - Checksum (4 bytes): CRC32

2. Packet types as constants:
   const (
     PacketTypeData = 0x01
     PacketTypeAck = 0x02
     PacketTypeNack = 0x03
     PacketTypeControl = 0x04
   )

3. Functions:
   - SerializePacket(packet *Packet) ([]byte, error)
   - DeserializePacket(data []byte) (*Packet, error)
   - CalculateChecksum(data []byte) uint32
   - VerifyChecksum(packet *Packet) bool

4. Include thorough validation and error handling
5. Add unit tests for serialization/deserialization
```

---

### Day 17-19: UDP Sender Implementation

**Cursor Prompt:**

```
Create UDP sender with parallel streams for TrackShift in internal/transport/udp_sender.go

Requirements:
1. UDPSender struct with:
   - Connection pool (multiple UDP sockets)
   - Send queue (buffered channel)
   - ACK tracker (map[uint32]time.Time)
   - Retransmission scheduler
   - Stats tracker (sent, acked, retransmitted)

2. Configuration:
   - MaxParallelStreams (default 32, max 256)
   - RetransmitTimeout (default 200ms)
   - MaxRetries (default 10)
   - WindowSize (default 256 packets)

3. Methods:
   - NewUDPSender(config *Config) *UDPSender
   - SendChunk(chunk []byte, metadata ChunkMetadata) error
   - StartWorkers(numWorkers int) - parallel senders
   - HandleAcks() - ACK receiver goroutine
   - Retransmit() - retransmission scheduler
   - GetStats() *TransferStats

4. Features:
   - Sliding window flow control
   - Adaptive retransmission (RTT-based)
   - Congestion avoidance (basic AIMD)
   - Packet reordering handling

5. Thread-safe implementation with proper synchronization

Include extensive logging for debugging
```

---

### Day 20-21: UDP Receiver Implementation

**Cursor Prompt:**

```
Create UDP receiver for TrackShift in internal/transport/udp_receiver.go

Requirements:
1. UDPReceiver struct with:
   - UDP listener socket
   - Receive buffer (out-of-order packet reassembly)
   - ACK sender
   - Duplicate detection (bitmap/bloom filter)
   - Chunk assembler

2. Methods:
   - NewUDPReceiver(port int) *UDPReceiver
   - Start() - main receive loop
   - HandlePacket(packet *Packet) - process incoming packets
   - SendAck(seqNum uint32, sessionID string)
   - AssembleChunk(chunkID string) ([]byte, bool) - returns chunk when complete

3. Features:
   - Out-of-order packet handling
   - Fast ACK generation
   - Duplicate packet detection
   - Missing packet tracking (send NACKs)
   - Statistics collection

4. Handle multiple concurrent sessions

Add comprehensive error handling
```

---

### Day 22-23: Update Sender/Receiver Mains for UDP

**Cursor Prompt:**

```
Update cmd/sender/main.go to use UDP transport instead of TCP.

Changes needed:
1. Add flag --protocol (tcp or udp, default udp)
2. Add flag --parallel-streams (default 32)
3. Replace TCPSender with UDPSender
4. Add statistics dashboard:
   - Packets sent/acked/lost
   - Effective throughput
   - Retransmission rate
   - RTT (round-trip time)

5. Maintain backward compatibility with TCP

Update the transfer loop to:
- Create UDPSender with config
- Start worker goroutines
- Feed chunks to send queue
- Monitor ACKs and handle retransmissions
- Display real-time stats

Also update cmd/receiver/main.go similarly for UDP protocol.
```

---

### Day 24-25: Performance Testing & Tuning

**Create performance test script:**

```bash
cat > test/performance/benchmark_udp.sh << 'EOF'
#!/bin/bash

echo "TrackShift UDP Performance Benchmark"

# Test configurations
FILE_SIZE=5000  # 5GB
CHUNK_SIZE=100  # 100MB
PARALLEL_STREAMS=(1 4 8 16 32 64 128)

# Create test file
echo "Creating ${FILE_SIZE}MB test file..."
dd if=/dev/urandom of=/tmp/test_5gb.bin bs=1M count=$FILE_SIZE 2>/dev/null

# Test each parallelism level
for STREAMS in "${PARALLEL_STREAMS[@]}"; do
  echo ""
  echo "Testing with $STREAMS parallel streams..."
  
  # Start receiver
  ./bin/receiver --protocol udp --port 9090 \
    --output-dir /tmp/receiver &
  RECEIVER_PID=$!
  sleep 1
  
  # Start sender and capture stats
  START_TIME=$(date +%s)
  ./bin/sender --protocol udp \
    --file /tmp/test_5gb.bin \
    --receiver localhost:9090 \
    --parallel-streams $STREAMS \
    --chunk-size $((CHUNK_SIZE * 1024 * 1024))
  END_TIME=$(date +%s)
  
  DURATION=$((END_TIME - START_TIME))
  THROUGHPUT=$((FILE_SIZE / DURATION))
  
  echo "Duration: ${DURATION}s"
  echo "Throughput: ${THROUGHPUT} MB/s"
  
  # Cleanup
  kill $RECEIVER_PID 2>/dev/null
  rm -f /tmp/receiver/*
  sleep 2
done

# Cleanup
rm -f /tmp/test_5gb.bin
EOF

chmod +x test/performance/benchmark_udp.sh
```

**Run benchmarks:**
```bash
make build
./test/performance/benchmark_udp.sh
```

**Network simulation test (with packet loss):**

```bash
cat > test/performance/test_unstable_network.sh << 'EOF'
#!/bin/bash

# Simulate unstable network with netem
sudo tc qdisc add dev lo root netem delay 100ms loss 5% duplicate 1% corrupt 0.1%

echo "Testing transfer over simulated unstable network..."
echo "- 100ms latency"
echo "- 5% packet loss"
echo "- 1% packet duplication"

# Run transfer test
./test/integration/test_basic_transfer.sh

# Remove network simulation
sudo tc qdisc del dev lo root

echo "Network simulation removed."
EOF

chmod +x test/performance/test_unstable_network.sh
```

**✅ PHASE 2 COMPLETE CHECKPOINT:**
- [ ] UDP protocol implementation working
- [ ] Parallel streams (32+) functional
- [ ] Achieves >1GB/s on localhost
- [ ] Handles packet loss gracefully
- [ ] Retransmission works correctly
- [ ] 10x faster than TCP baseline

---

## 🌐 PHASE 3: Edge Grid & Orchestrator (Week 5-6)

### Day 26-28: Simple Edge Relay

**Cursor Prompt:**

```
Create edge relay server for TrackShift in cmd/relay/main.go

Requirements:
1. Relay acts as UDP packet forwarder
2. Command-line flags:
   - --listen-port (incoming from senders)
   - --forward-address (next relay or receiver)
   - --relay-id (unique identifier)
   - --orchestrator-url (register with orchestrator)

3. Functionality:
   - Listen for UDP packets
   - Minimal state tracking (session routing table)
   - Forward packets to configured destination
   - Send heartbeats to orchestrator
   - Report metrics (packets forwarded, bytes, latency)

4. Features:
   - Packet buffering (small buffer for burst handling)
   - Basic routing logic (forward based on session ID)
   - Health checks
   - Graceful shutdown

Create lightweight, stateless relay that can run in containers.

Also create internal/relay/forwarder.go with core relay logic.
```

---

### Day 29-31: Orchestrator Service

**Cursor Prompt:**

```
Create orchestrator service for TrackShift in cmd/orchestrator/main.go

This is a REST API service that coordinates transfers.

Requirements:
1. HTTP server (use net/http or fiber)
2. Endpoints:
   
   POST /api/v1/session
   - Create new transfer session
   - Allocate relay path
   - Return session token and relay endpoints
   
   GET /api/v1/session/:id
   - Get session status
   - Return chunk map, progress, stats
   
   PUT /api/v1/session/:id/chunk/:chunkId
   - Update chunk status
   
   GET /api/v1/relays
   - List available relays
   
   POST /api/v1/relays/register
   - Register new relay (from relay heartbeat)
   
   GET /api/v1/metrics
   - System-wide metrics

3. Data storage:
   - In-memory for MVP (can add Redis later)
   - Relay registry (list of active relays with health)
   - Session state
   - Metrics aggregation

4. Features:
   - Relay selection algorithm (choose closest relay)
   - Session lifecycle management
   - Health monitoring of relays
   - Basic authentication (JWT tokens)

5. Use internal/orchestrator/ package for business logic

Add Swagger/OpenAPI docs for API.
```

**Install dependencies:**
```bash
go get github.com/gofiber/fiber/v2
go get github.com/golang-jwt/jwt/v5
```

---

### Day 32-33: Update Sender/Receiver for Orchestrator

**Cursor Prompt:**

```
Update TrackShift sender to use orchestrator:

1. Add flags:
   - --orchestrator-url (default http://localhost:8000)
   - --auth-token (JWT token)

2. Modified flow:
   a. Contact orchestrator to create session
   b. Receive relay endpoints and session token
   c. Connect to assigned relay (instead of direct receiver)
   d. Send chunks through relay
   e. Report progress to orchestrator periodically

3. Add HTTP client in internal/client/orchestrator_client.go:
   - CreateSession(fileInfo) (Session, error)
   - UpdateProgress(sessionID, chunkID, status) error
   - GetSession(sessionID) (Session, error)

Similarly update receiver to register with orchestrator and accept connections via relay.
```

---

### Day 34-35: Multi-Relay Testing

**Create Docker Compose for testing:**

```bash
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  orchestrator:
    build:
      context: .
      dockerfile: Dockerfile.orchestrator
    ports:
      - "8000:8000"
    environment:
      - RELAY_CHECK_INTERVAL=5s
    networks:
      - trackshift-net

  relay-1:
    build:
      context: .
      dockerfile: Dockerfile.relay
    environment:
      - RELAY_ID=relay-us-east-1
      - LISTEN_PORT=9001
      - ORCHESTRATOR_URL=http://orchestrator:8000
    ports:
      - "9001:9001/udp"
    networks:
      - trackshift-net
    depends_on:
      - orchestrator

  relay-2:
    build:
      context: .
      dockerfile: Dockerfile.relay
    environment:
      - RELAY_ID=relay-us-west-1
      - LISTEN_PORT=9002
      - ORCHESTRATOR_URL=http://orchestrator:8000
    ports:
      - "9002:9002/udp"
    networks:
      - trackshift-net
    depends_on:
      - orchestrator

networks:
  trackshift-net:
    driver: bridge
EOF
```

**Create Dockerfiles:**

```bash
# Dockerfile.orchestrator
cat > Dockerfile.orchestrator << 'EOF'
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /orchestrator ./cmd/orchestrator

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /orchestrator /orchestrator
EXPOSE 8000
CMD ["/orchestrator"]
EOF

# Dockerfile.relay
cat > Dockerfile.relay << 'EOF'
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /relay ./cmd/relay

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /relay /relay
EXPOSE 9001/udp
CMD ["/relay"]
EOF
```

**Test multi-relay setup:**
```bash
docker-compose up -d
# Test transfer through relay infrastructure
./bin/sender --file test.bin --orchestrator-url http://localhost:8000
```

**✅ PHASE 3 COMPLETE CHECKPOINT:**
- [ ] Orchestrator service running
- [ ] Relays registering with orchestrator
- [ ] Transfers routed through relay
- [ ] Session management working
- [ ] Multi-relay path functional

---

## 🛡️ PHASE 4: Resilience & Erasure Coding (Week 7-8)

### Day 36-38: Reed-Solomon Implementation

**Cursor Prompt:**

```
Implement erasure coding with Reed-Solomon for TrackShift in internal/erasure/reed_solomon.go

Requirements:
1. Use github.com/klauspost/reedsolomon library

2. ErasureCoder struct:
   - DataShards int (default 10)
   - ParityShards int (default 3)
   - ShardSize int

3. Methods:
   - Encode(data []byte) ([][]byte, error)
     * Splits data into shards
     * Generates parity shards
     * Returns all shards (data + parity)
   
   - Decode(shards [][]byte) ([]byte, error)
     * Reconstructs missing shards
     * Returns original data
   
   - CalculateShardSize(dataSize int64) int
   - ValidateShards(shards [][]byte) error

4. Features:
   - Automatic padding for shard alignment
   - Support for missing/corrupted shards
   - Efficient memory usage

5. Integration:
   - Update Chunker to create parity chunks
   - Update Receiver to reconstruct from available chunks

6. Tests:
   - Encode/decode cycle
   - Recovery with 1, 2, 3 missing shards
   - Performance benchmarks

go get github.com/klauspost/reedsolomon
```

---

### Day 39-41: Intelligent Retry Logic

**Cursor Prompt:**

```
Implement intelligent retry and backoff logic for TrackShift in internal/transport/retry_manager.go

Requirements:
1. RetryManager struct:
   - MaxRetries int
   - BaseBackoff time.Duration (100ms)
   - MaxBackoff time.Duration (30s)
   - BackoffMultiplier float64 (2.0)
   - JitterFactor float64 (0.1)

2. Retry strategies:
   - Exponential backoff with jitter
   - Adaptive backoff based on RTT
   - Circuit breaker pattern (stop retrying if too many failures)

3. Methods:
   - ShouldRetry(attempt int, err error) bool
   - NextBackoff(attempt int, rtt time.Duration) time.Duration
   - RecordSuccess(identifier string)
   - RecordFailure(identifier string, err error)
   - GetCircuitState(identifier string) CircuitState

4. Circuit breaker states:
   - Closed (normal operation)
   - Open (too many failures, reject requests)
   - HalfOpen (testing if recovered)

5. Integration:
   - Use in UDPSender for packet retransmission
   - Use in OrchestrationClient for API calls
   - Track per-chunk retry state

Include metrics for monitoring retry behavior.
```

---

### Day 42-43: Resume After Network Change

**Cursor Prompt:**

```
Implement resume capability after IP address change for TrackShift.

Update internal/session/manager.go:

1. Add methods:
   - PersistCheckpoint(sessionID string) error
     * Save current state to disk every 10 seconds
   
   - ResumeSession(sessionID string, newReceiverAddr string) error
     * Load saved state
     * Update receiver address
     * Resume from last checkpoint
   
   - GetMissingChunks(sessionID string) []string
     * Return list of chunks not completed

2. SessionCheckpoint struct:
   - SessionID
   - CompletedChunks []string
   - PendingChunks []string
   - TotalChunks int
   - LastUpdateTime time.Time
   - ReceiverAddress string

3. Features:
   - Atomic checkpoint writes
   - Corruption detection (checksum checkpoint file)
   - Automatic cleanup of old checkpoints
   - Fast resume (skip completed chunks)

Update cmd/sender/main.go to:
- Add --resume flag with session ID
- Detect network changes (monitor local IP)
- Automatically reconnect and resume

Test scenarios:
- Kill and restart sender
- Change WiFi network mid-transfer
- Simulate mobile handoff
```

---

### Day 44-45: Integration & Stress Testing

**Create resilience test suite:**

```bash
cat > test/resilience/test_recovery.sh << 'EOF'
#!/bin/bash
set -e

echo "=== TrackShift Resilience Test Suite ==="

# Test 1: Recovery with packet loss
echo ""
echo "Test 1: Transfer with 10% packet loss"
sudo tc qdisc add dev lo root netem loss 10%
./test/integration/test_basic_transfer.sh
sudo tc qdisc del dev lo root
echo "✓ Test 1 passed"

# Test 2: Recovery with missing chunks (erasure coding)
echo ""
echo "Test 2: Erasure coding recovery"
cat > /tmp/test_erasure.go << 'GOEOF'
package main

import (
    "fmt"
    "os"
    "trackshift/internal/erasure"
)

func main() {
    // Create test data
    data := make([]byte, 1024*1024) // 1MB
    for i := range data {
        data[i] = byte(i % 256)
    }
    
    // Encode with erasure coding
    ec := erasure.NewErasureCoder(10, 3)
    shards, err := ec.Encode(data)
    if err != nil {
        fmt.Printf("Encode failed: %v\n", err)
        os.Exit(1)
    }
    
    // Simulate losing 3 shards
    shards[2] = nil
    shards[5] = nil
    shards[9] = nil
    
    // Attempt recovery
    recovered, err := ec.Decode(shards)
    if err != nil {
        fmt.Printf("Decode failed: %v\n", err)
        os.Exit(1)
    }
    
    // Verify
    if len(recovered) != len(data) {
        fmt.Printf("Size mismatch: %d vs %d\n", len(recovered), len(data))
        os.Exit(1)
    }
    
    for i := range data {
        if data[i] != recovered[i] {
            fmt.Printf("Data mismatch at byte %d\n", i)
            os.Exit(1)
        }
    }
    
    fmt.Println("✓ Erasure coding recovery successful")
}
GOEOF

go run /tmp/test_erasure.go
rm /tmp/test_erasure.go

# Test 3: Resume after interruption
echo ""
echo "Test 3: Resume after sender crash"
./bin/receiver --port 9090 --output-dir /tmp/receiver &
RECEIVER_PID=$!
sleep 1

# Start transfer, kill after 30% progress
timeout 10s ./bin/sender --file /tmp/test_5gb.bin --receiver localhost:9090 || true
sleep 2

# Resume transfer
./bin/sender --resume --receiver localhost:9090
kill $RECEIVER_PID

# Verify file
if sha256sum -c /tmp/receiver/testfile.bin.sha256; then
    echo "✓ Test 3 passed - Resume successful"
else
    echo "✗ Test 3 failed"
    exit 1
fi

echo ""
echo "=== All resilience tests passed ==="
EOF

chmod +x test/resilience/test_recovery.sh
```

**Run resilience tests:**
```bash
./test/resilience/test_recovery.sh
```

**✅ PHASE 4 COMPLETE CHECKPOINT:**
- [ ] Erasure coding (Reed-Solomon) working
- [ ] Can recover from 3 lost chunks (10+3 config)
- [ ] Intelligent retry with exponential backoff
- [ ] Circuit breaker prevents retry storms
- [ ] Resume works after crash
- [ ] Resume works after IP address change
- [ ] Passes all resilience tests

---

## 🤖 PHASE 5: AI Optimization (Week 9-10)

### Day 46-48: Telemetry Collection

**Cursor Prompt:**

```
Create telemetry system for TrackShift in internal/telemetry/collector.go

Requirements:
1. TelemetryCollector struct that tracks:
   - Network metrics:
     * Bandwidth (current, average, peak)
     * Latency (RTT)
     * Packet loss rate
     * Jitter
   - Transfer metrics:
     * Chunk success rate
     * Retransmission rate
     * Effective throughput
     * Time to completion per chunk
   - System metrics:
     * CPU usage
     * Memory usage
     * Disk I/O

2. Methods:
   - RecordMetric(name string, value float64, tags map[string]string)
   - GetMetrics(timeWindow time.Duration) *MetricsSnapshot
   - ExportPrometheus() string (Prometheus format)
   - StartCollection() (background goroutine)
   - StopCollection()

3. MetricsSnapshot struct:
   - Timestamp
   - NetworkStats
   - TransferStats
   - SystemStats
   - Aggregated statistics (min, max, avg, p50, p95, p99)

4. Features:
   - Time-series storage (circular buffer)
   - Efficient aggregation
   - Export to Prometheus
   - Real-time monitoring

5. Integration points:
   - UDPSender: record packet stats
   - UDPReceiver: record receive stats
   - SessionManager: record session stats
   - Orchestrator: aggregate from all components

Use github.com/prometheus/client_golang for Prometheus integration

go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

---

### Day 49-51: AI Optimizer Service (Python)

**Create Python microservice:**

```bash
mkdir -p ai-optimizer
cd ai-optimizer

cat > requirements.txt << 'EOF'
fastapi==0.104.1
uvicorn==0.24.0
numpy==1.24.3
scikit-learn==1.3.0
pydantic==2.4.2
prometheus-client==0.18.0
EOF

cat > optimizer.py << 'EOF'
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import numpy as np
from typing import Optional
import logging

app = FastAPI(title="TrackShift AI Optimizer")
logging.basicConfig(level=logging.INFO)

class NetworkMetrics(BaseModel):
    bandwidth_mbps: float
    latency_ms: float
    packet_loss_pct: float
    jitter_ms: float
    cpu_usage_pct: float

class OptimizationParams(BaseModel):
    chunk_size_mb: int
    parallel_streams: int
    parity_ratio: float
    compression_level: int

class Optimizer:
    """Simple heuristic optimizer (can be replaced with ML model)"""
    
    def optimize(self, metrics: NetworkMetrics) -> OptimizationParams:
        # Chunk size optimization
        if metrics.latency_ms < 50:
            # Low latency - use larger chunks
            chunk_size = 200
        elif metrics.latency_ms < 200:
            chunk_size = 100
        else:
            # High latency - smaller chunks for faster feedback
            chunk_size = 50
        
        # Parallel streams optimization
        if metrics.bandwidth_mbps > 1000:
            # High bandwidth - more streams
            parallel_streams = 128
        elif metrics.bandwidth_mbps > 100:
            parallel_streams = 64
        else:
            parallel_streams = 32
        
        # Adjust for packet loss
        if metrics.packet_loss_pct > 5:
            # High loss - fewer streams, more parity
            parallel_streams = max(16, parallel_streams // 2)
            parity_ratio = 0.4
        elif metrics.packet_loss_pct > 2:
            parity_ratio = 0.3
        else:
            parity_ratio = 0.2
        
        # Compression level based on CPU
        if metrics.cpu_usage_pct < 50:
            compression_level = 9  # High compression
        elif metrics.cpu_usage_pct < 75:
            compression_level = 5  # Balanced
        else:
            compression_level = 1  # Fast compression
        
        return OptimizationParams(
            chunk_size_mb=chunk_size,
            parallel_streams=parallel_streams,
            parity_ratio=parity_ratio,
            compression_level=compression_level
        )

optimizer = Optimizer()

@app.post("/optimize", response_model=OptimizationParams)
async def optimize_transfer(metrics: NetworkMetrics):
    """Get optimized parameters based on current network conditions"""
    try:
        params = optimizer.optimize(metrics)
        logging.info(f"Optimized params: {params}")
        return params
    except Exception as e:
        logging.error(f"Optimization failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    return {"status": "healthy"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8001)
EOF

cat > Dockerfile << 'EOF'
FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY optimizer.py .

EXPOSE 8001
CMD ["python", "optimizer.py"]
EOF
```

**Start the optimizer:**
```bash
cd ai-optimizer
pip install -r requirements.txt
python optimizer.py &
cd ..
```

---

### Day 52-53: Integrate AI Optimizer

**Cursor Prompt:**

```
Integrate AI optimizer into TrackShift sender.

Create internal/optimizer/client.go:

Requirements:
1. OptimizerClient struct:
   - BaseURL string (AI optimizer service URL)
   - HTTPClient *http.Client
   - Cache (cache recent recommendations)

2. Methods:
   - GetOptimizedParams(metrics *telemetry.MetricsSnapshot) (*OptimizationParams, error)
     * Send current metrics to AI service
     * Receive optimized parameters
     * Cache result for 30 seconds
   
   - ApplyParams(params *OptimizationParams) error
     * Update sender configuration
     * Adjust chunk size
     * Adjust parallel streams
     * Update erasure coding ratio

3. Integration in cmd/sender/main.go:
   - Add flag --ai-optimizer-url (default http://localhost:8001)
   - Add flag --adaptive (enable AI optimization)
   - Background goroutine that:
     * Collects metrics every 10 seconds
     * Queries optimizer
     * Applies new parameters dynamically
     * Logs parameter changes

4. Graceful parameter transitions:
   - Don't change mid-chunk
   - Apply at chunk boundaries
   - Smooth transitions (no abrupt changes)

5. Add optimization metrics:
   - Parameter change frequency
   - Optimization effectiveness (throughput before/after)
   - Decision logging

Create similar integration for receiver.
```

---

### Day 54-55: Priority Channels

**Cursor Prompt:**

```
Implement priority channels for TrackShift in internal/transport/priority_queue.go

Requirements:
1. Priority levels (0 = highest):
   - 0: Control messages (session control, ACKs)
   - 1: Metadata and manifests
   - 2: Critical chunks (first/last chunks)
   - 3: Normal data chunks
   - 4: Parity chunks

2. PriorityQueue implementation:
   - Based on heap data structure
   - O(log n) insert and remove
   - Support for dynamic priority updates

3. PrioritySender wrapper:
   - Wraps UDPSender
   - Uses PriorityQueue for send ordering
   - Ensures high-priority packets sent first
   - Fair queuing (prevent starvation of low priority)

4. Integration:
   - Mark chunk priority in metadata
   - Send first chunk with high priority (enables early validation)
   - Send metadata chunks first
   - Interleave priorities to prevent blocking

5. Monitoring:
   - Track queue depth per priority
   - Measure priority inversion events
   - Latency per priority level

Update UDPSender to use PrioritySender wrapper.
Add priority field to Packet header.
```

---

### Day 56: Adaptive Testing & Benchmarking

**Create adaptive behavior test:**

```bash
cat > test/performance/test_adaptive.sh << 'EOF'
#!/bin/bash

echo "=== Testing AI Adaptive Optimization ==="

# Start AI optimizer
cd ai-optimizer
python optimizer.py &
OPTIMIZER_PID=$!
sleep 2
cd ..

# Test scenario 1: Good network -> Bad network
echo ""
echo "Scenario 1: Network degradation"

# Start with good network
sudo tc qdisc add dev lo root netem delay 10ms loss 0%
./bin/receiver --port 9090 &
RECEIVER_PID=$!
sleep 1

# Start adaptive sender
./bin/sender \
  --file /tmp/test_1gb.bin \
  --receiver localhost:9090 \
  --adaptive \
  --ai-optimizer-url http://localhost:8001 &
SENDER_PID=$!

# After 30 seconds, degrade network
sleep 30
echo "Degrading network..."
sudo tc qdisc change dev lo root netem delay 200ms loss 10%

# Wait for completion
wait $SENDER_PID

# Check if transfer completed
if [ -f /tmp/receiver/test_1gb.bin ]; then
    echo "✓ Scenario 1 passed - Adapted to degraded network"
else
    echo "✗ Scenario 1 failed"
fi

# Cleanup
kill $RECEIVER_PID 2>/dev/null
sudo tc qdisc del dev lo root

# Scenario 2: Bad network -> Good network
echo ""
echo "Scenario 2: Network improvement"
sudo tc qdisc add dev lo root netem delay 200ms loss 10%
./bin/receiver --port 9090 &
RECEIVER_PID=$!
sleep 1

./bin/sender \
  --file /tmp/test_1gb.bin \
  --receiver localhost:9090 \
  --adaptive \
  --ai-optimizer-url http://localhost:8001 &
SENDER_PID=$!

sleep 30
echo "Improving network..."
sudo tc qdisc change dev lo root netem delay 10ms loss 0%

wait $SENDER_PID

if [ -f /tmp/receiver/test_1gb.bin ]; then
    echo "✓ Scenario 2 passed - Scaled up on better network"
else
    echo "✗ Scenario 2 failed"
fi

# Cleanup
kill $RECEIVER_PID 2>/dev/null
kill $OPTIMIZER_PID 2>/dev/null
sudo tc qdisc del dev lo root

echo ""
echo "=== Adaptive optimization tests complete ==="
EOF

chmod +x test/performance/test_adaptive.sh
```

**✅ PHASE 5 COMPLETE CHECKPOINT:**
- [ ] Telemetry collection working
- [ ] Prometheus metrics exported
- [ ] AI optimizer service running
- [ ] Adaptive parameter adjustment working
- [ ] Priority channels implemented
- [ ] Demonstrable throughput improvement with adaptation
- [ ] Passes adaptive behavior tests

---

## 🔒 PHASE 6: Production Hardening (Week 11-12)

### Day 57-58: Security Hardening

**Cursor Prompt:**

```
Implement security hardening for TrackShift.

Create internal/security/encryption.go:

Requirements:
1. End-to-end encryption:
   - Use AES-256-GCM for chunk encryption
   - Implement key exchange (ECDH)
   - Per-session encryption keys
   - Support for customer-managed keys (KMS integration)

2. EncryptionManager struct:
   - GenerateKeyPair() (public, private []byte)
   - DeriveSharedSecret(theirPublic, myPrivate []byte) []byte
   - EncryptChunk(data []byte, key []byte) ([]byte, error)
   - DecryptChunk(encrypted []byte, key []byte) ([]byte, error)

3. Authentication:
   - Implement JWT-based auth in internal/security/auth.go
   - Token generation and validation
   - Role-based access control (sender, receiver, admin)

4. mTLS for control plane:
   - Update orchestrator to require client certificates
   - Certificate generation and management
   - Automatic certificate rotation

5. Security features:
   - Rate limiting (prevent DDoS)
   - Input validation (all user inputs)
   - SQL injection prevention (if using database)
   - Secrets management (no hardcoded secrets)

6. Update all components:
   - Sender: encrypt chunks before send
   - Receiver: decrypt after receive
   - Orchestrator: require auth tokens
   - Relays: validate packet signatures

Use crypto/aes, crypto/ecdh, crypto/rand from standard library.
Add github.com/golang-jwt/jwt/v5 for JWT.
```

---

### Day 59-60: Monitoring & Alerting

**Cursor Prompt:**

```
Create comprehensive monitoring dashboard for TrackShift.

Create cmd/dashboard/main.go:

Requirements:
1. Web dashboard (use Fiber or Gin):
   - Real-time transfer monitoring
   - System health overview
   - Relay status map
   - Performance charts
   - Alert management

2. Pages:
   - /dashboard - Overview with live stats
   - /sessions - List all active sessions
   - /sessions/:id - Detailed session view with chunk map
   - /relays - Relay health and geography
   - /metrics - Prometheus metrics viewer
   - /alerts - Alert configuration and history

3. Real-time updates:
   - WebSocket for live data
   - Server-sent events (SSE) alternative
   - Auto-refresh every 5 seconds

4. Alert system in internal/alerting/alerting.go:
   - Alert rules:
     * Transfer stalled (no progress for 2 minutes)
     * High packet loss (>10%)
     * Relay unhealthy
     * Low disk space
   - Alert channels:
     * Email (SMTP)
     * Webhook (Slack, Discord, PagerDuty)
     * Log file
   - Alert throttling (prevent spam)

5. Frontend (simple HTML/JS):
   - Use Chart.js for visualizations
   - Bootstrap for UI
   - Real-time chart updates

go get github.com/gofiber/fiber/v2
go get github.com/gofiber/websocket/v2
```

**Create basic dashboard:**

```bash
mkdir -p cmd/dashboard/static

cat > cmd/dashboard/static/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>TrackShift Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js"></script>
</head>
<body>
    <nav class="navbar navbar-dark bg-dark">
        <div class="container-fluid">
            <span class="navbar-brand">TrackShift Dashboard</span>
        </div>
    </nav>
    
    <div class="container-fluid mt-4">
        <div class="row">
            <div class="col-md-3">
                <div class="card">
                    <div class="card-body">
                        <h5>Active Transfers</h5>
                        <h2 id="activeTransfers">0</h2>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card">
                    <div class="card-body">
                        <h5>Total Throughput</h5>
                        <h2 id="throughput">0 MB/s</h2>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card">
                    <div class="card-body">
                        <h5>Active Relays</h5>
                        <h2 id="activeRelays">0</h2>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card">
                    <div class="card-body">
                        <h5>Packet Loss</h5>
                        <h2 id="packetLoss">0%</h2>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="row mt-4">
            <div class="col-md-12">
                <div class="card">
                    <div class="card-body">
                        <canvas id="throughputChart"></canvas>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="row mt-4">
            <div class="col-md-12">
                <div class="card">
                    <div class="card-body">
                        <h5>Active Sessions</h5>
                        <table class="table" id="sessionsTable">
                            <thead>
                                <tr>
                                    <th>Session ID</th>
                                    <th>File</th>
                                    <th>Progress</th>
                                    <th>Speed</th>
                                    <th>ETA</th>
                                </tr>
                            </thead>
                            <tbody></tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    </div>
    
    <script>
        // WebSocket connection
        const ws = new WebSocket('ws://localhost:8000/ws');
        
        // Initialize chart
        const ctx = document.getElementById('throughputChart').getContext('2d');
        const chart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Throughput (MB/s)',
                    data: [],
                    borderColor: 'rgb(75, 192, 192)',
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: { beginAtZero: true }
                }
            }
        });
        
        // Handle WebSocket messages
        ws.onmessage = function(event) {
            const data = JSON.parse(event.data);
            updateDashboard(data);
        };
        
        function updateDashboard(data) {
            document.getElementById('activeTransfers').textContent = data.activeTransfers;
            document.getElementById('throughput').textContent = data.throughput.toFixed(2) + ' MB/s';
            document.getElementById('activeRelays').textContent = data.activeRelays;
            document.getElementById('packetLoss').textContent = data.packetLoss.toFixed(2) + '%';
            
            // Update chart
            chart.data.labels.push(new Date().toLocaleTimeString());
            chart.data.datasets[0].data.push(data.throughput);
            if (chart.data.labels.length > 20) {
                chart.data.labels.shift();
                chart.data.datasets[0].data.shift();
            }
            chart.update();
            
            // Update sessions table
            updateSessionsTable(data.sessions);
        }
        
        function updateSessionsTable(sessions) {
            const tbody = document.querySelector('#sessionsTable tbody');
            tbody.innerHTML = '';
            sessions.forEach(session => {
                const row = tbody.insertRow();
                row.innerHTML = `
                    <td>${session.id.substring(0, 8)}...</td>
                    <td>${session.filename}</td>
                    <td>
                        <div class="progress">
                            <div class="progress-bar" style="width: ${session.progress}%">
                                ${session.progress.toFixed(1)}%
                            </div>
                        </div>
                    </td>
                    <td>${session.speed.toFixed(2)} MB/s</td>
                    <td>${session.eta}</td>
                `;
            });
        }
    </script>
</body>
</html>
EOF
```

---

### Day 61-62: Database Integration (Optional but Recommended)

**Cursor Prompt:**

```
Add PostgreSQL database for persistent storage in TrackShift.

Create internal/database/postgres.go:

Requirements:
1. Use GORM or pgx for database access
2. Tables:
   - sessions (id, sender_id, receiver_id, filename, size, status, created_at, updated_at)
   - chunks (id, session_id, chunk_id, status, hash, size, created_at)
   - relays (id, relay_id, address, region, status, last_heartbeat)
   - metrics (id, timestamp, metric_name, value, tags)
   - audit_logs (id, timestamp, action, user_id, session_id, details)

3. Database interface:
   - Connect(dsn string) error
   - SaveSession(*models.TransferSession) error
   - GetSession(id string) (*models.TransferSession, error)
   - UpdateChunkStatus(sessionID, chunkID string, status ChunkStatus) error
   - RegisterRelay(*models.Relay) error
   - RecordMetric(*models.Metric) error
   - QueryMetrics(filter MetricFilter) ([]*models.Metric, error)

4. Migration system:
   - Use golang-migrate or GORM AutoMigrate
   - Version control for schema
   - Rollback capability

5. Connection pooling and optimization:
   - Max connections: 25
   - Connection timeout: 30s
   - Query timeout: 10s
   - Prepared statements

6. Update SessionManager to use database instead of file storage

go get gorm.io/gorm
go get gorm.io/driver/postgres
```

**Setup PostgreSQL:**
```bash
# Using Docker
docker run --name trackshift-db \
  -e POSTGRES_PASSWORD=trackshift \
  -e POSTGRES_DB=trackshift \
  -p 5432:5432 \
  -d postgres:15

# Wait for PostgreSQL to start
sleep 5

# Update orchestrator configuration
cat > configs/orchestrator.yaml << 'EOF'
database:
  host: localhost
  port: 5432
  user: postgres
  password: trackshift
  dbname: trackshift
  sslmode: disable

server:
  port: 8000
  
relay:
  check_interval: 5s
  unhealthy_threshold: 3
  
auth:
  jwt_secret: change-me-in-production
  token_expiry: 24h
EOF
```

---

### Day 63-64: S3 Integration

**Cursor Prompt:**

```
Add S3-compatible storage integration for TrackShift.

Create internal/storage/s3.go:

Requirements:
1. S3Client wrapper:
   - Use AWS SDK for Go v2
   - Support AWS S3, MinIO, DigitalOcean Spaces, etc.

2. Methods:
   - UploadChunk(bucket, key string, data []byte) error
   - DownloadChunk(bucket, key string) ([]byte, error)
   - UploadFile(bucket, key, filepath string) error
   - ListChunks(bucket, prefix string) ([]string, error)
   - DeleteChunk(bucket, key string) error

3. Features:
   - Multipart upload for large files
   - Concurrent chunk uploads
   - Retry logic with exponential backoff
   - Presigned URLs for direct upload/download
   - Server-side encryption

4. Integration points:
   - Receiver: option to upload completed files to S3
   - Sender: option to read files from S3
   - Orchestrator: store session metadata in S3

5. Configuration:
   - S3 endpoint
   - Access key / Secret key
   - Bucket name
   - Region
   - Storage class

go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
```

**Add CLI flags for S3:**
```bash
# Receiver with S3 output
./bin/receiver --port 9090 \
  --s3-bucket my-bucket \
  --s3-endpoint https://s3.amazonaws.com \
  --s3-access-key ACCESS_KEY \
  --s3-secret-key SECRET_KEY

# Sender with S3 input
./bin/sender --s3-object s3://my-bucket/large-file.bin \
  --receiver localhost:9090
```

---

### Day 65-66: CLI Enhancements & Documentation

**Cursor Prompt:**

```
Create enhanced CLI with better UX for TrackShift.

Update cmd/sender/main.go with cobra CLI framework:

Requirements:
1. Command structure:
   trackshift send <file> --to <receiver>
   trackshift receive --port <port>
   trackshift resume <session-id>
   trackshift list - list active sessions
   trackshift status <session-id> - check session status
   trackshift config - manage configuration
   trackshift version - show version info

2. Global flags:
   --config - config file path
   --verbose - enable debug logging
   --quiet - suppress output
   --json - JSON output format

3. Interactive mode:
   - Prompt for missing required flags
   - Confirmation before large transfers
   - Progress bar with detailed stats
   - Color-coded output (success/error/warning)

4. Configuration file support:
   - YAML configuration
   - Environment variables
   - Flag precedence: flags > env vars > config file > defaults

5. Shell completion:
   - Bash
   - Zsh
   - Fish
   - PowerShell

go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/fatih/color
```

**Generate documentation:**
```bash
# Use Cursor to generate comprehensive docs

mkdir -p docs

# README with quick start
# ARCHITECTURE.md with technical details
# API.md with REST API documentation
# DEPLOYMENT.md with infrastructure guide
# SECURITY.md with security best practices
# CONTRIBUTING.md with development guide
```

---

### Day 67-68: Performance Optimization & Profiling

**Create profiling test:**

```bash
cat > test/performance/profile.sh << 'EOF'
#!/bin/bash

echo "=== TrackShift Performance Profiling ==="

# Build with profiling enabled
go build -o bin/sender-prof ./cmd/sender

# Run with CPU profiling
echo "Running CPU profile..."
./bin/sender-prof \
  --file /tmp/test_10gb.bin \
  --receiver localhost:9090 \
  --cpuprofile=/tmp/cpu.prof

# Run with memory profiling
echo "Running memory profile..."
./bin/sender-prof \
  --file /tmp/test_10gb.bin \
  --receiver localhost:9090 \
  --memprofile=/tmp/mem.prof

# Analyze profiles
echo ""
echo "CPU Profile Top Functions:"
go tool pprof -top /tmp/cpu.prof

echo ""
echo "Memory Profile Top Allocations:"
go tool pprof -top /tmp/mem.prof

# Generate flame graphs
go tool pprof -http=:8080 /tmp/cpu.prof &
echo "CPU flame graph available at http://localhost:8080"

# Run benchmarks
echo ""
echo "Running benchmarks..."
go test -bench=. -benchmem ./...

echo ""
echo "Profiling complete!"
EOF

chmod +x test/performance/profile.sh
```

**Optimization checklist:**
```bash
cat > OPTIMIZATION_CHECKLIST.md << 'EOF'
# TrackShift Optimization Checklist

## Completed Optimizations
- [ ] Zero-copy I/O where possible
- [ ] Buffer pooling (