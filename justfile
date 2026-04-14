# DocVault — Development Commands
set shell := ["/bin/sh", "-c"]
set dotenv-load := true

ENV_FILE := ".env"
DATABASE_URL := env("DATABASE_URL", "postgresql://docvault:docvault_dev@localhost:5432/docvault")

help:
    @echo ""
    @echo "DocVault — Development Commands"
    @echo ""
    @echo "  just dev-up          Start infra (postgres, redis, rabbitmq, minio)"
    @echo "  just dev-down        Stop all infra services"
    @echo "  just dev-logs        View infra logs"
    @echo "  just dev-ps          Show running services"
    @echo "  just dev-restart     Restart infra services"
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
    @echo "  just dev-clean        Clean up containers + data"
    @echo "  just dev-prune        Remove unused Docker resources"
    @echo ""

dev-up:
    @echo "Starting infrastructure services..."
    docker compose --env-file {{ENV_FILE}} up -d

dev-down:
    @echo "Stopping infrastructure services..."
    docker compose --env-file {{ENV_FILE}} down

dev-logs:
    docker compose --env-file {{ENV_FILE}} logs -f

dev-ps:
    docker compose --env-file {{ENV_FILE}} ps

dev-restart:
    just dev-down
    just dev-up

dev-backend:
    cd backend && go run ./cmd/api

dev-reminder:
    cd reminder && go run ./cmd/reminder

# OCR Service
dev-ocr-install:
    cd ocr && uv sync --all-extras

dev-ocr: dev-ocr-install
    cd ocr && uv run python -m ocr.main

# Processing Pipeline Service
dev-processing-install:
    cd processing && uv sync --all-extras

dev-processing: dev-processing-install
    cd processing && uv run python -m processing.main

dev-web-install:
	cd web && bun install

dev-web: dev-web-install
	cd web && bun run dev

dev-mobile-install:
    cd mobile && bun install

dev-mobile: dev-mobile-install
    cd mobile && bun run start

dev-tmux:
    ./scripts/dev-tmux.sh


dev-setup:
    just dev-up
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
