# reminder/AGENTS.md — reminder service

Part of the docvault monorepo. Read root **[../AGENTS.md](../AGENTS.md)** first (commit
identity, conventions).

## Stack

Go · **RabbitMQ** consumer (no HTTP surface) · **OpenRouter** for date extraction. Separate Go
module from `backend/`.

## Layout

```
cmd/reminder/            entrypoint
internal/
  reminder/              OpenRouterDateExtractor (structured JSON: expiry/renewal/due dates)
  transport/             RabbitMQ consumer + connection
```

## Commands & gate

```sh
just dev-reminder
cd reminder && go test ./...   # gate (run before commit)
```

## Conventions

- **Queue consumer only.** Consumes reminder jobs published by the processing service and
  derives reminder dates (expiry/renewal from start + duration).
- Conventional commits, green gate before commit, and the root **commit-identity rule** apply.
