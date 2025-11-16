.PHONY: build test clean deps fmt lint run-sender run-receiver

build:
	go build -o bin/sender ./cmd/sender
	go build -o bin/receiver ./cmd/receiver
	go build -o bin/relay ./cmd/relay
	go build -o bin/orchestrator ./cmd/orchestrator
	go build -o bin/dashboard ./cmd/dashboard

test:
	go test -v ./...

clean:
	rm -rf bin/
	go clean

deps:
	go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run || true

run-sender:
	./bin/sender --config configs/sender.yaml

run-receiver:
	./bin/receiver --config configs/receiver.yaml


