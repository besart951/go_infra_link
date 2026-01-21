.PHONY: migrate-up migrate-down migrate-goto migrate-version lint lint-fix

MIGRATIONS_PATH ?= ./migrations
DATABASE_URL ?=

migrate-up:
	go run ./cmd/migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

migrate-down:
	go run ./cmd/migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1

migrate-goto:
	go run ./cmd/migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" goto 1

migrate-version:
	go run ./cmd/migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" version

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...
