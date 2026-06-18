# DocVault

A calm, bilingual (Arabic / English, RTL) document-intelligence vault for GCC legal,
finance, admin, and government users: scan or upload → OCR → extract dates & entities →
translate AR↔EN → surface expiries → answer questions over the content. Everything is
tenant-scoped.

This is a monorepo of cooperating services. The canonical engineering guide is
**[AGENTS.md](./AGENTS.md)**; brand/UX truth lives in **[DESIGN.md](./DESIGN.md)** and
**[PRODUCT.md](./PRODUCT.md)**. Each service also has its own nested `AGENTS.md`.

## Services

| Service | Stack | Dir |
| --- | --- | --- |
| **Backend** | Go · net/http · pgx · sqlc · goose · Casbin | `backend/` |
| **Reminder** | Go · RabbitMQ consumer | `reminder/` |
| **OCR** | Python · uv · Mistral OCR · RabbitMQ | `ocr/` |
| **Processing** | Python · uv · OpenRouter · pgvector · RabbitMQ | `processing/` |
| **Web** | Next.js 15 · React 19 · Tailwind v4 | `web/` |
| **Mobile** | Expo (SDK 55) · expo-router · React Native | `mobile/` |
| **Shared** | Shared config, transport, theme, types (Python/TS) | `shared/` |

**Pipeline:** upload → `docvault.ocr.jobs` → **OCR** → `docvault.processing.jobs` →
**Processing** (classify → translate → chunk → embed → pgvector → suggest folder → publish
reminder) → **Reminder**. Files live in MinIO; metadata + vectors in Postgres.

---

## Prerequisites

- **Docker** (for infra: Postgres+pgvector, RabbitMQ, Redis, MinIO, observability)
- **[`just`](https://github.com/casey/just)** — the task runner (`just help` lists everything)
- **Go** 1.22+ · **uv** (Python) · **pnpm** (frontend) · **goose** (migrations)

---

## Run the project

### 1. Configure environment

```sh
cp .env.example .env      # then fill in API keys as needed (Mistral, OpenRouter, SMTP…)
```

The dev defaults in `.env.example` work out of the box for local infra; the external API
keys (`MISTRAL_API_KEY`, `OPENROUTER_API_KEY`, …) are only needed for the OCR/processing
pipeline.

### 2. Install dependencies

```sh
just dev-setup            # installs uv, pnpm, and go deps across all services
```

### 3. Start infrastructure

```sh
just dev-up               # Postgres+pgvector, RabbitMQ, Redis, MinIO + observability
```

### 4. Apply database migrations

```sh
just db-migrate           # goose migrations in backend/internal/migrate/sql/
```

### 5. Seed sample users (optional, recommended for local dev)

```sh
just db-seed
```

See **[Seeding the database](#seeding-the-database)** below for the accounts this creates.

### 6. Run the services

Each runs in its own terminal:

```sh
just dev-backend          # Go API        → http://localhost:8080
just dev-web              # Next.js web   → http://localhost:3000
just dev-reminder         # reminder worker
just dev-ocr              # OCR service
just dev-processing       # processing pipeline
just dev-mobile           # Expo / React Native
```

Or run everything at once in a tmux workspace:

```sh
just dev-tmux
```

### Stop / reset

```sh
just dev-down             # stop infra
just db-reset             # tear down infra + delete the database volume
just dev-clean            # clean up containers + data
```

---

## Seeding the database

`just db-seed` populates the dev database with a representative set of tenants,
organizations, and users across every role. It reuses the exact same primitives as user
registration (`backend/internal/transport/http/auth_register.go`) — bcrypt password hashing,
the sqlc `Create*` queries, and Casbin policy + role-binding seeding — so seeded users are
indistinguishable from real ones. The command lives at `backend/cmd/seed/main.go`.

```sh
just db-seed
```

It **applies migrations first** (so it works against a fresh database in one step) and is
**idempotent** — re-running skips users that already exist, with no duplicate errors.

### Seeded accounts

All seeded users share the password **`Passw0rd!`**.

| Tenant | Org | Users (role) |
| --- | --- | --- |
| **Acme Legal** | Acme Legal LLP | `owner@acme.test` (owner) · `admin@acme.test` (admin) · `member@acme.test`, `member2@acme.test` (member) · `viewer@acme.test` (viewer) |
| **Gulf Finance** | Gulf Finance Holding | `owner@gulf.test` (owner) · `admin@gulf.test` (admin) · `member@gulf.test` (member) · `viewer@gulf.test` (viewer) |
| **Doha Admin Office** | Doha Admin Office | `owner@doha.test` (owner) · `member@doha.test` (member) |

Emails follow `<role><n>@<slug>.test` (the numeric suffix only appears when a role repeats,
e.g. `member2@acme.test`). The three tenants are fully isolated from one another, which makes
the seed useful for exercising the RBAC/ACL model and tenant scoping.

To log in: start the backend (`just dev-backend`) and the web app (`just dev-web`), then sign
in with any seeded email and `Passw0rd!`.

---

## Database, migrations & sqlc

- **Migrations** are [goose](https://github.com/pressly/goose) files in
  `backend/internal/migrate/sql/`. Apply with `just db-migrate` (`db-status`, `db-rollback`,
  `db-reset`).
- **sqlc is offline / file-based.** Run `just sqlc-generate` after any `.sql` change, then
  confirm `just sqlc-check` is clean. No running DB is needed to regenerate. Never hand-edit
  the generated `backend/internal/db/*.sql.go` or `schema.sql`.

---

## Verification gates

| Area | Gate |
| --- | --- |
| Backend / Reminder (Go) | `cd backend && go build ./... && go test ./...` · `just sqlc-check` clean |
| Backend integration (DB) | `just test-integration` (needs a live DB) |
| Web | `cd web && npx tsc --noEmit` (note: `npm run lint` is broken at baseline) |
| OCR / Processing (Python) | `cd <svc> && uv run ruff check . && uv run pytest -q` |

For everything else — architecture, conventions, the security model, and per-service
guidance — read **[AGENTS.md](./AGENTS.md)**.
