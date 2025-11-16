#!/bin/bash
set -e

echo "TrackShift TCP MVP Integration Test"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

rm -rf /tmp/trackshift-test
mkdir -p /tmp/trackshift-test/{sender,receiver,sessions}

echo "Creating 500MB test file..."
dd if=/dev/urandom of=/tmp/trackshift-test/sender/testfile.bin bs=1M count=500 status=none

echo "Building binaries..."
make build

echo "Starting receiver..."
./bin/receiver --port 9090 \
  --output-dir /tmp/trackshift-test/receiver \
  --temp-dir /tmp/trackshift-test/receiver/temp \
  --sessions-dir /tmp/trackshift-test/sessions &
RECEIVER_PID=$!

sleep 2

echo "Starting sender..."
./bin/sender --file /tmp/trackshift-test/sender/testfile.bin \
  --receiver localhost:9090 \
  --chunk-size 52428800 \
  --output-dir /tmp/trackshift-test/sessions

wait $RECEIVER_PID || true

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


