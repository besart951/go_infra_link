.PHONY: migrate-up migrate-down migrate-goto migrate-version lint lint-fix

MIGRATIONS_PATH ?= ./migrations

migrate-up:
	go run ./cmd/migrate -path $(MIGRATIONS_PATH) up

migrate-down:
	go run ./cmd/migrate -path $(MIGRATIONS_PATH) down 1

migrate-goto:
	go run ./cmd/migrate -path $(MIGRATIONS_PATH) goto 1

migrate-version:
	go run ./cmd/migrate -path $(MIGRATIONS_PATH) version

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...
