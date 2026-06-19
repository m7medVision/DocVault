# AGENTS.md — docvault monorepo

The canonical guide for any agent (Claude Code, Cursor, CI bots, …) working in this repo.
Each service has its **own nested `AGENTS.md`** — read the one for the area you are touching.
Brand/UX truth lives in **[DESIGN.md](./DESIGN.md)** and **[PRODUCT.md](./PRODUCT.md)**.

**docvault** is a calm, bilingual (Arabic/English, RTL) document-intelligence vault for GCC
legal / finance / admin / gov users: scan or upload → OCR → extract dates & entities →
translate AR↔EN → surface expiries → answer questions over the content. Tenant-scoped.

---

## 0. Golden rules (read before you do anything)

1. **Commit identity is sacred.** Never attribute a commit/branch/PR to Claude, Anthropic,
   or any AI — no `Co-Authored-By: Claude` trailer, no AI author/committer name or
   `*@anthropic.com` email, not in commits, not in PRs, **ever**. Use the repo's human git
   identity. If you can't, don't commit — ask. (See [CLAUDE.md](./CLAUDE.md).)
2. **Green before commit.** Run the area's gate (below) and only commit if it passes. Never
   commit broken code.
3. **Conventional Commits only:** `feat:` `fix:` `chore:` `test:` `ci:` `docs:` (+ scope,
   e.g. `feat(authz): …`). One logical change per commit.
4. **Stage explicitly.** Never `git add -A` / `git add .`; stage the exact paths you changed.
5. **Branch off `master`.** Don't commit directly to `master`. Don't push or open PRs unless
   asked.
6. **Tenant scoping is non-negotiable.** Every data access is scoped by `tenant_id` (+ `org_id`).
   There is no Postgres RLS — isolation is application-enforced, so a missing `WHERE` clause
   is a data leak. Treat retrieval/visibility code as security-critical.
7. **Reuse before you build.** Prefer existing components, utilities, and patterns.

---

## 1. Services (each has its own AGENTS.md)

| Service | Lang / Stack | Dir | Guide |
| --- | --- | --- | --- |
| **Backend** | Go · net/http · pgx · sqlc · goose · Casbin | `backend/` | [backend/AGENTS.md](./backend/AGENTS.md) |
| **Reminder** | Go · RabbitMQ consumer | `reminder/` | [reminder/AGENTS.md](./reminder/AGENTS.md) |
| **OCR** | Python · uv · Mistral OCR · RabbitMQ | `ocr/` | [ocr/AGENTS.md](./ocr/AGENTS.md) |
| **Processing** | Python · uv · OpenRouter · pgvector · RabbitMQ | `processing/` | [processing/AGENTS.md](./processing/AGENTS.md) |
| **Web** | Next.js 15 · React 19 · Tailwind v4 · shadcn-style | `web/` | [web/AGENTS.md](./web/AGENTS.md) |
| **Mobile** | Expo (SDK 55) · expo-router · React Native | `mobile/` | [mobile/AGENTS.md](./mobile/AGENTS.md) |
| **Shared** | Python/TS shared config, transport, theme, types | `shared/` | [shared/AGENTS.md](./shared/AGENTS.md) |

---

## 2. Quickstart

`just` is the task runner. `just help` lists everything.

```sh
just dev-setup     # install all deps (uv, pnpm, go)
just dev-up        # start infra (Postgres+pgvector, RabbitMQ, Redis, MinIO) + observability
just dev-backend   # run a service (also: dev-reminder, dev-ocr, dev-processing, dev-web, dev-mobile)
just dev-down      # stop infra
```

**Multi-service flow:** upload → `docvault.ocr.jobs` → **OCR** → `docvault.processing.jobs`
→ **Processing** (classify → translate → chunk → embed→pgvector → suggest folder → publish
reminder) → **Reminder**. Files in MinIO; metadata + vectors in Postgres.

---

## 3. Database, migrations & sqlc

- **Migrations:** [goose](https://github.com/pressly/goose) files in
  `backend/internal/migrate/sql/` (`-- +goose Up` / `Down`). Apply with `just db-migrate`
  (`db-status`, `db-rollback`, `db-reset`).
- **sqlc is offline / file-based.** `just sqlc-generate` runs `sqlc-schema` (concatenates the
  `Up` blocks of every migration into `backend/internal/db/schema.sql`) then `sqlc generate`.
  **No running DB is needed to regenerate.** Run it after any `.sql` change, then confirm
  `just sqlc-check` is clean. **Never hand-edit** `backend/internal/db/*.sql.go` or
  `schema.sql`.
- **Queries** live in `backend/internal/query/*.sql` using `sqlc.arg()/sqlc.narg()` with casts.

---

## 4. Verification gates (per area)

| Area | Gate |
| --- | --- |
| Backend / Reminder (Go) | `cd backend && go build ./... && go test ./...` · `just sqlc-check` clean |
| Backend integration (DB) | `just test-integration` (tests behind `//go:build integration`; needs live DB) |
| Web | `cd web && npx tsc --noEmit` — **`npm run lint` is broken at baseline; do NOT use it** |
| OCR / Processing (Python) | `cd <svc> && uv run ruff check . && uv run pytest -q` |

---

## 5. Design system & theming (summary — detail in [web/AGENTS.md](./web/AGENTS.md))

- **shadcn-style components** live in `web/components/ui/` (Radix primitives + `cva` variants
  + the `cn()` merge from `web/lib/utils.ts`). **Reuse them; never add a second component or
  icon library.** Icons: **lucide-react** only.
- **Tailwind v4**, utility-first. Style with design tokens, never raw hex.
- **Tokens are OKLCH CSS variables** in `web/app/globals.css` (light + dark): warm-paper
  background, **deep ink-blue** `--primary`, **date-amber** `--brand-amber` reserved for
  dates/reminders/expiry. Neutrals are tinted toward the brand hue.
- **Type:** Noto Sans (Latin) + Noto Sans Arabic — one family, two scripts.
- **i18n/RTL:** `next-intl`, locales `en`/`ar`; all strings in `web/messages/{en,ar}.json`
  (`eslint-plugin-i18next` forbids hardcoded copy); RTL via `dir`, logical Tailwind utilities.
- Honor the **bans** in [DESIGN.md](./DESIGN.md) (no glassmorphism, no gradient text, no em
  dashes in copy, no "AI magic"/"magic" marketing language, lucide only, …).

---

## 6. Security & permissions model

- **Auth:** JWT (access+refresh); claims carry `tenant_id`, `org_id`, `role`.
- **RBAC:** Casbin with domains (`owner > admin > member > viewer`), policy seeded per tenant.
- **Granular ACL (org-open + restrict):** by default org members see all org docs; a document
  or folder can be `is_restricted`, after which only its owner, org admins, or explicitly
  granted users/groups (with folder inheritance) may see it. The **visibility predicate lives
  in the single retrieval seam** `backend/internal/query/search.sql` (`filtered_chunks`),
  which closes `/search` **and** `/chat`. Row-level reads use `requireDocVisible`/
  `requireFolderVisible` and return **404, not 403**, for invisible resources.
- When touching retrieval, listing, or sharing: assume it's a leak vector and test it.

---

## 7. Conventions recap

- Non-interactive shells: always `-f` with `rm/cp/mv`, `-y` with package managers.
- Python: `uv` (`uv sync`, `uv run`). Frontend: `pnpm` for deps, `npm run` for scripts.
- Keep changes at the altitude of the surrounding code; match its idioms and comment density.
