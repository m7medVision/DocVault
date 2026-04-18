# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Tech Stack & Commands

Use `just` as the primary task runner.

| Service | Language | Dev Command | Lint/Verify |
| :--- | :--- | :--- | :--- |
| **Backend** | Go | `just dev-backend` | `go test ./...` |
| **Reminder** | Go | `just dev-reminder` | `go test ./...` |
| **OCR** | Python | `just dev-ocr` | `uv run ruff check .` |
| **Processing**| Python | `just dev-processing`| `uv run ruff check .` |
| **Web** | Next.js | `just dev-web` | `bun run lint` |
| **Mobile** | Expo | `just dev-mobile` | - |

### Infrastructure
- **Start/Stop:** `just dev-up` / `just dev-down` (Docker Compose).
- **Database:** `just db-migrate` (uses `goose` in `backend/internal/migrate/sql`).
- **Dependencies:** `just dev-setup` installs all deps (uv, bun, go).

## Architecture & Workflows

### Multi-Service Flow
1. **Infrastructure**: Postgres (pgvector), RabbitMQ, Redis, MinIO.
2. **OCR Pipeline**: `docvault.ocr.jobs` -> OCR Service -> `docvault.processing.jobs` -> Processing Service.
3. **Storage**: MinIO for files, Postgres for metadata/vectors.

### Critical Conventions
- **Migrations**: Always use `goose` via `just db-migrate`. New migrations go in `backend/internal/migrate/sql`.
- **Python**: Uses `uv` for dependency management (`uv sync`, `uv run`).
- **Frontend**: Uses `bun` for package management.
- **Non-Interactive**: ALWAYS use `-f` with `rm`, `cp`, `mv` and `-y` with package managers.

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
