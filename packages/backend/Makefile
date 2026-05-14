.PHONY: dev build test lint clean migrate

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

migrate:
	@echo "Running database migrations..."
	@for f in internal/db/migrations/*.sql; do \
		echo "Applying $$f..."; \
		psql "$$DATABASE_URL" < "$$f"; \
	done

.DEFAULT_GOAL := build
