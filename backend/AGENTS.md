# backend/AGENTS.md — Go API

Part of the docvault monorepo. Read the root **[../AGENTS.md](../AGENTS.md)** first — the
**commit-identity rule (never attribute commits to Claude/AI)**, conventional commits, and
tenant-scoping rules all apply here.

## Stack

Go (1.22+ `net/http` ServeMux with method+path patterns) · `pgx/v5` + `pgxpool` ·
**sqlc** (generated `internal/db`) · **goose** migrations · **Casbin** RBAC · JWT auth ·
pgvector retrieval. Module `github.com/docvault/backend`.

## Layout

```
cmd/api/                 entrypoint
internal/
  transport/http/        handlers + routes.go (route → Casbin obj/act wiring)
  usecase/               business logic (services)
  repository/            data access (wraps sqlc Queries)
  query/*.sql            sqlc query sources  ← you edit these
  db/                    sqlc-GENERATED (schema.sql, *.sql.go) ← never hand-edit
  migrate/sql/           goose migrations (NNN_*.sql, +goose Up/Down)
  domain/                domain models
  authz/                 Casbin model.conf, policy.go, seeder, ACL helpers
  middleware/            jwt.go, casbin.go, rbac.go (GetTenantID/OrgID/UserID/Role, HasMinRole)
  search/                embedder (OpenRouter) + vector literal helpers
```

## Commands & gate

```sh
just dev-backend            # run
go build ./... && go test ./...     # unit gate (run before every commit)
just sqlc-generate          # after ANY .sql change (offline; regenerates internal/db)
just sqlc-check             # must be clean (no diff) once staged
just db-migrate             # apply migrations to the dev DB
just test-integration       # DB-backed tests behind //go:build integration
```

## Conventions

- **Schema/query change → migration + `just sqlc-generate` + `just sqlc-check` clean.** Stage
  the regenerated `internal/db/*` alongside the `.sql` change in the same commit. Never edit
  generated files by hand. Queries use `sqlc.arg()/sqlc.narg()` with explicit `::type` casts;
  array params as `::uuid[]` / `::text[]`.
- **Tenant/org scope every query.** There is no RLS; `tenant_id` (+ `org_id`) must be in the
  `WHERE` of every read/write. `owner_id`/`created_by` are provenance only — authz is Casbin +
  the ACL layer, not ownership columns.
- **Permissions:** route-level `Authorize(enforcer, object, action)` (coarse, by resource
  type) **plus** row-level checks in handlers: `requireDocVisible` / `requireDocWritable` /
  `requireFolderVisible` (admins short-circuit; return **404 NOT_FOUND**, never 403, for
  invisible resources so existence isn't leaked). New ACL/visibility SQL goes in
  `internal/query/acl.sql` and mirrors the existing cycle-protected recursive folder CTEs
  (path accumulator + depth cap) and the `g.org_id` grant scoping.
- **Retrieval is the leak seam.** The visibility predicate in `internal/query/search.sql`
  (`filtered_chunks`) is what stops restricted content reaching `/search` and `/chat` —
  change it with care and re-run `just test-integration`.
- **Folder reparent must go through the advisory-locked `repository.Reparent`** (per-tenant
  `pg_advisory_xact_lock`); the raw `MoveFolder` SQL is intentionally not exposed because a
  bare statement does not prevent the concurrent-cycle race under READ COMMITTED.
- **Handlers:** pull `tenantID/orgID/userID/role` from `middleware`, validate, call the
  usecase, write an audit event via `auditSvc` for mutations, return JSON.
- **DB-backed tests** carry `//go:build integration`, seed inside a rolled-back `pgx` tx
  (`sqldb.New(tx)`), and are excluded from `go test ./...`.

The **Reminder** service is a separate Go module (queue consumer, no HTTP) — see
[../reminder/AGENTS.md](../reminder/AGENTS.md).
