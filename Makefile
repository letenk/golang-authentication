# Load .env variables
include .env
export

# Build once, reuse
.PHONY: run build test test-coverage migration-build migration-create migration-up migration-down migration-status bob-gen

run:
	go run cmd/main.go

build:
	go build -o main cmd/main.go

# Run all tests
test:
	go test ./... -count=1

# Run all tests with coverage report
test-coverage:
	go test ./... -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out

# Build migration binary (only when needed)
migration-build:
	@echo "Building migration binary..."
	go build -v -o migration migrations/cmd/main.go
	@echo "✅ Migration binary built successfully"

# Create new migration
migration-create:
	@if [ ! -f migration ]; then $(MAKE) migration-build; fi
	./migration create $(name) $(type)

# Run migrations up
migration-up:
	@if [ ! -f migration ]; then $(MAKE) migration-build; fi
	./migration postgres up

# Rollback migration
migration-down:
	@if [ ! -f migration ]; then $(MAKE) migration-build; fi
	./migration postgres down

# Check migration status
migration-status:
	@if [ ! -f migration ]; then $(MAKE) migration-build; fi
	./migration postgres status

# Show current migration version
migration-version:
	@if [ ! -f migration ]; then $(MAKE) migration-build; fi
	./migration postgres version

# Generate Bob ORM models from database schema
bob-gen:
	PSQL_DSN=postgres://$(DB_CONFIGS_USER):$(DB_CONFIGS_PASSWORD)@$(DB_CONFIGS_HOST):$(DB_CONFIGS_PORT)/$(DB_CONFIGS_NAME)?sslmode=$(DB_CONFIGS_SSLMODE) go run github.com/stephenafamo/bob/gen/bobgen-psql@latest

# Clean built binaries
clean:
	rm -f main migration
	@echo "✅ Cleaned binaries"

# Help command
help:
	@echo "Available commands:"
	@echo "  make build              - Build main application"
	@echo "  make run                - Run application"
	@echo "  make migration-build    - Build migration binary"
	@echo "  make migration-create   - Create new migration (name=xxx type=sql)"
	@echo "  make migration-up       - Run all pending migrations"
	@echo "  make migration-down     - Rollback last migration"
	@echo "  make migration-status   - Show migration status"
	@echo "  make migration-version  - Show current migration version"
	@echo "  make bob-gen            - Generate Bob ORM models from DB schema"
	@echo "  make test               - Run all tests"
	@echo "  make test-coverage      - Run all tests with coverage report"
	@echo "  make clean              - Remove built binaries"

.DEFAULT_GOAL := help
