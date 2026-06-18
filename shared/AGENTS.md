# shared/AGENTS.md — cross-service shared code

Part of the docvault monorepo. Read root **[../AGENTS.md](../AGENTS.md)** first (commit
identity, conventions).

## What's here

```
shared/
  py/         docvault_shared — Python config + RabbitMQ transport
              (RabbitMQConnection, QueueConsumer/Publisher) + the approved
              embedding model/dim allow-list. Used by ocr/ and processing/.
  types/      shared type definitions
  theme/      shared theme tokens
  telemetry/  shared telemetry/observability helpers
```

## Conventions

- **Changes here ripple across services.** `shared/py` (`docvault_shared`) is imported by both
  the **OCR** and **Processing** Python services — a change to config, transport, or the
  embedding allow-list affects both. Verify each dependent service's gate
  (`uv run ruff check . && uv run pytest -q`) after editing.
- Keep `shared/theme` aligned with the web design tokens in
  [../web/AGENTS.md](../web/AGENTS.md) / [../DESIGN.md](../DESIGN.md); the OKLCH values are the
  source of truth.
- The embedding model/dim (`text-embedding-3-large` @ 1024) is intentionally an allow-list of
  one — changing it requires a pgvector migration in `backend/` and coordination.
