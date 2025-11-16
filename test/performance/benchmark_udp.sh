#!/bin/bash

echo "TrackShift UDP Performance Benchmark"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

FILE_SIZE=5000  # 5GB
CHUNK_SIZE=100  # 100MB
PARALLEL_STREAMS=(1 4 8 16 32)

echo "Creating ${FILE_SIZE}MB test file..."
dd if=/dev/urandom of=/tmp/trackshift-udp-test.bin bs=1M count=$FILE_SIZE status=none

for STREAMS in "${PARALLEL_STREAMS[@]}"; do
  echo ""
  echo "Testing with $STREAMS parallel streams..."

  ./bin/receiver --protocol udp --port 9090 \
    --output-dir /tmp/trackshift-udp-recv &
  RECEIVER_PID=$!
  sleep 1

  START_TIME=$(date +%s)
  ./bin/sender --protocol udp \
    --file /tmp/trackshift-udp-test.bin \
    --receiver localhost:9090 \
    --parallel-streams "$STREAMS" \
    --chunk-size $((CHUNK_SIZE * 1024 * 1024))
  END_TIME=$(date +%s)

  DURATION=$((END_TIME - START_TIME))
  THROUGHPUT=$((FILE_SIZE / DURATION))

  echo "Duration: ${DURATION}s"
  echo "Throughput: ${THROUGHPUT} MB/s"

  kill "$RECEIVER_PID" 2>/dev/null || true
  rm -rf /tmp/trackshift-udp-recv
  sleep 2
done

rm -f /tmp/trackshift-udp-test.bin


