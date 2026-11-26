# Load .env variables
include .env
export

# Build once, reuse
.PHONY: run build test test-coverage test-db-up test-db-down test-integration swagger migration-build migration-create migration-up migration-down migration-status

# Disposable Postgres used by integration tests (see docker-compose.test.yml)
TEST_DATABASE_URL ?= postgres://postgres:postgres@localhost:55432/auth_test?sslmode=disable

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

# Start / stop the disposable Postgres for integration tests
test-db-up:
	docker compose -f docker-compose.test.yml up -d --wait

test-db-down:
	docker compose -f docker-compose.test.yml down -v

# Run unit + integration tests against TEST_DATABASE_URL (run `make test-db-up` first)
test-integration:
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -tags=integration -count=1 -p 1 ./...

# Regenerate the Swagger spec into api/swagger from the handler annotations.
# The swag CLI is pinned here instead of in go.mod; keep it in step with github.com/swaggo/swag in go.mod.
SWAG_VERSION ?= v1.16.6
swagger:
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/main.go -o api/swagger --parseInternal --exclude docs

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
	@echo "  make test               - Run all tests"
	@echo "  make test-coverage      - Run all tests with coverage report"
	@echo "  make test-db-up         - Start disposable Postgres for integration tests"
	@echo "  make test-db-down       - Stop disposable Postgres for integration tests"
	@echo "  make test-integration   - Run unit + integration tests (needs test-db-up)"
	@echo "  make swagger            - Regenerate Swagger docs into api/swagger"
	@echo "  make clean              - Remove built binaries"

.DEFAULT_GOAL := help
