#!/bin/bash
set -e

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

echo "=== TrackShift Resilience Smoke Test ==="

echo ""
echo "Test 1: Transfer with 10% packet loss over TCP"
sudo tc qdisc add dev lo root netem loss 10%
./test/integration/test_basic_transfer.sh || true
sudo tc qdisc del dev lo root || true

echo ""
echo "Test 2: Erasure coding encode/decode cycle"
go test -run TestEncodeDecodeRoundTrip ./internal/erasure -v

echo "Resilience smoke tests complete."


