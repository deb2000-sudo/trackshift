# TrackShift - High-Speed Resilient File Transfer

Fast, resilient file mover for unstable networks and very large files (50–100GB+).

## Quick Start

```bash
make build

# Run sender
./bin/sender --file /path/to/large/file.bin --receiver 192.168.1.100:8080

# Run receiver
./bin/receiver --port 8080 --output-dir /path/to/destination/
```

## Project Layout

- `cmd/` – main entrypoints (`sender`, `receiver`, `relay`, `orchestrator`, `dashboard`)
- `internal/` – internal packages (chunker, session, crypto, transport, etc.)
- `pkg/` – shared public packages (`models`, `protocol`, `utils`)
- `configs/` – configuration files
- `scripts/` – helper scripts
- `test/` – integration, performance, and resilience tests
- `docs/` – architecture and operational documentation

## Development

```bash
# Install deps
make deps

# Run tests
make test

# Format code
make fmt

# Lint (if golangci-lint installed)
make lint
```


