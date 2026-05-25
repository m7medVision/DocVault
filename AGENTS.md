# Agent Instructions

## Tech Stack & Commands

Use `just` as the primary task runner.

| Service | Language | Dev Command | Lint/Verify |
| :--- | :--- | :--- | :--- |
| **Backend** | Go | `just dev-backend` | `go test ./...` |
| **Reminder** | Go | `just dev-reminder` | `go test ./...` |
| **OCR** | Python | `just dev-ocr` | `uv run ruff check .` |
| **Processing**| Python | `just dev-processing`| `uv run ruff check .` |
| **Web** | Next.js | `just dev-web` | `npm run lint` |
| **Mobile** | Expo | `just dev-mobile` | - |

### Infrastructure
- **Start/Stop:** `just dev-up` / `just dev-down` (Docker Compose).
- **Database:** `just db-migrate` (uses `goose` in `backend/internal/migrate/sql`).
- **Dependencies:** `just dev-setup` installs all deps (uv, pnpm, go).

## Architecture & Workflows

### Multi-Service Flow
1. **Infrastructure**: Postgres (pgvector), RabbitMQ, Redis, MinIO.
2. **OCR Pipeline**: `docvault.ocr.jobs` -> OCR Service -> `docvault.processing.jobs` -> Processing Service.
3. **Storage**: MinIO for files, Postgres for metadata/vectors.

### Critical Conventions
- **Migrations**: Always use `goose` via `just db-migrate`. New migrations go in `backend/internal/migrate/sql`.
- **Python**: Uses `uv` for dependency management (`uv sync`, `uv run`).
- **Frontend**: Uses `pnpm` for dependency management and `npm run ...` for package scripts.
- **Non-Interactive**: ALWAYS use `-f` with `rm`, `cp`, `mv` and `-y` with package managers.
