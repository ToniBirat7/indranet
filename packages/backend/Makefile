.PHONY: dev build test lint clean

dev:
	@echo "Starting backend with hot reload..."
	air -c .air.toml

build:
	go build -o bin/server ./cmd/server

test:
	go test -v -race -count=1 ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

# Migrations run automatically on server startup via embedded SQL files.
# To reset: use scripts/reset-db.sh and restart the backend.

.DEFAULT_GOAL := build
