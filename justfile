# DocVault — Development Commands
set shell := ["/bin/sh", "-c"]
set dotenv-load := true

ENV_FILE := ".env"
DATABASE_URL := env("DATABASE_URL", "postgresql://docvault:docvault_dev@localhost:5432/docvault")
SQLC_VERSION := "v1.31.1"

help:
    @echo ""
    @echo "DocVault — Development Commands"
    @echo ""
    @echo "  just dev-up          Start infra + observability"
    @echo "  just dev-down        Stop infra + observability"
    @echo "  just dev-logs        View Docker service logs"
    @echo "  just dev-ps          Show running services"
    @echo "  just dev-restart     Restart infra services"
    @echo "  just obs-up          Start Grafana observability stack"
    @echo "  just obs-down        Stop Grafana observability stack"
    @echo "  just obs-logs        View observability logs"
    @echo ""
    @echo "  just dev-backend     Run backend API (Go)"
    @echo "  just dev-reminder    Run reminder service (Go)"
    @echo "  just dev-ocr         Run OCR service (Python)"
    @echo "  just dev-processing   Run processing pipeline (Python)"
    @echo "  just dev-web          Run web frontend (Bun/Next.js)"
    @echo "  just dev-mobile       Run mobile (Bun/Expo)"
    @echo ""
    @echo "  just dev-all          Run all services (separate terminals!)"
    @echo "  just dev-tmux         Run all services in a tmux workspace"
    @echo "  just dev-setup        Initial setup (install deps)"
    @echo "  just db-migrate      Run database migrations (goose)"
    @echo "  just db-rollback     Rollback last migration"
    @echo "  just db-status       Show migration status"
    @echo "  just db-seed         Seed sample tenants/orgs/users (idempotent)"
    @echo "  just sqlc-install    Install pinned sqlc"
    @echo "  just sqlc-generate   Generate SQLC code"
    @echo "  just sqlc-check      Verify SQLC code is fresh"
    @echo "  just dev-clean        Clean up containers + data"
    @echo "  just dev-prune        Remove unused Docker resources"
    @echo ""

dev-up:
    @echo "Starting infrastructure and observability services..."
    docker compose --env-file {{ENV_FILE}} up -d

dev-down:
    @echo "Stopping infrastructure and observability services..."
    docker compose --env-file {{ENV_FILE}} down

dev-logs:
    docker compose --env-file {{ENV_FILE}} logs -f

dev-ps:
    docker compose --env-file {{ENV_FILE}} ps

dev-restart:
    just dev-down
    just dev-up

obs-up:
    @echo "Starting observability stack..."
    docker compose --env-file {{ENV_FILE}} up -d grafana tempo loki prometheus otel-collector promtail

obs-down:
    @echo "Stopping observability stack..."
    docker compose --env-file {{ENV_FILE}} stop grafana tempo loki prometheus otel-collector promtail

obs-logs:
    docker compose --env-file {{ENV_FILE}} logs -f grafana tempo loki prometheus otel-collector promtail

dev-backend:
    cd backend && air -c .air.toml

dev-reminder:
    cd reminder && air -c .air.toml

# OCR Service
dev-ocr-install:
    cd ocr && uv sync --all-extras

dev-ocr: dev-ocr-install
    cd ocr && uv run watchfiles "python -m ocr.main" .

# Processing Pipeline Service
dev-processing-install:
    cd processing && uv sync --all-extras

dev-processing: dev-processing-install
    cd processing && uv run watchfiles "python -m processing.main" .

dev-web-install:
	cd web && pnpm install

dev-web: dev-web-install
	cd web && pnpm run dev

dev-mobile-install:
    cd mobile && pnpm install

dev-mobile: dev-mobile-install
    cd mobile && pnpm run start

dev-tmux:
    ./scripts/dev-tmux.sh


dev-setup:
    just dev-up
    @echo "Installing air (Go hot-reload)..."
    go install github.com/air-verse/air@latest 2>/dev/null || echo "air already installed or go not found"
    just dev-ocr-install
    just dev-processing-install
    just dev-web-install
    just dev-mobile-install

dev-clean:
    docker compose --env-file {{ENV_FILE}} down -v

dev-prune:
    docker system prune -f

db-migrate:
    goose -dir backend/internal/migrate/sql postgres "{{DATABASE_URL}}" up

db-rollback:
    goose -dir backend/internal/migrate/sql postgres "{{DATABASE_URL}}" down

db-status:
    goose -dir backend/internal/migrate/sql postgres "{{DATABASE_URL}}" status

db-reset:
    docker compose --env-file {{ENV_FILE}} down -v

# Seed the dev database with sample tenants/orgs/users (idempotent; applies migrations first)
db-seed:
    cd backend && go run ./cmd/seed

sqlc-install:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@{{SQLC_VERSION}}

sqlc-schema:
	mkdir -p backend/internal/db
	cd backend && sh scripts/sqlc-schema.sh > internal/db/schema.sql

sqlc-generate: sqlc-schema
	cd backend && sqlc generate

sqlc-check: sqlc-generate
	git diff --exit-code -- backend/sqlc.yaml backend/internal/query backend/internal/db

# Run DB-backed integration tests (require a live Postgres at DATABASE_URL).
test-integration:
	cd backend && go test -tags=integration ./...
