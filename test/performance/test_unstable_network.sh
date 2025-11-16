#!/bin/bash

set -e

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

echo "Testing transfer over simulated unstable network..."

sudo tc qdisc add dev lo root netem delay 100ms loss 5% duplicate 1% corrupt 0.1%

./test/integration/test_basic_transfer.sh || true

sudo tc qdisc del dev lo root || true

echo "Network simulation removed."


