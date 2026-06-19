# processing/AGENTS.md — enrichment & indexing service

Part of the docvault monorepo. Read root **[../AGENTS.md](../AGENTS.md)** first (commit
identity, conventions).

## Stack

Python · **uv** · **OpenRouter** for all generation/embeddings · **pgvector** via SQLAlchemy ·
**RabbitMQ** consumer. Shared config from `shared/py` (`docvault_shared`).

Models (via OpenRouter, OpenAI-compatible): embeddings **`openai/text-embedding-3-large`
truncated to 1024 dims** (the only approved embedding model/dim) · classify
`google/gemini-2.5-flash` · translate `mistralai/mistral-large`.

## Layout

```
processing/processing/
  main.py                      entrypoint — RabbitMQ consumer (docvault-processing)
  application/processing_job.py  orchestrator
  classifier.py                doc_type + metadata + suggest_folder_path (nested path)
  translation.py               AR→EN per page
  chunker.py                   char-window chunker (overlap, sentence boundaries)
  embeddings.py                OpenRouter embeddings (must be 1024-dim)
  database.py                  SQLAlchemy models + pgvector repo + folder suggestions
tests/
```

## Pipeline (order)

classify → translate (if not English) → chunk → embed → index into `extracted_text_chunks`
(pgvector) → **suggest** a nested folder path (writes the suggestion columns; does **not**
auto-file) → publish a reminder job.

## Commands & gate

```sh
just dev-processing-install
just dev-processing
cd processing && uv run ruff check . && uv run pytest -q   # gate (run before commit)
```

## Conventions

- **Embeddings must be 1024-dim `text-embedding-3-large` via OpenRouter** — the vector column
  is `vector(1024)`; do not change the model/dim without a migration + backend coordination.
- **Suggest, never auto-file.** `database.py` writes `suggested_folder_name` (a `/`-joined
  nested path), `suggested_filename`, `suggestion_confidence`, `suggestion_create_new` and
  sets `processing_stage="suggesting"`; the user accepts/dismisses via the backend. It must
  not mutate `folder_id`/`title` or create folders.
- The **SQLAlchemy `Document` model must map every column it writes** (e.g. the suggestion
  columns) — they're owned by the backend's goose migrations.
- `uv` for everything; config + RabbitMQ transport from `docvault_shared`.
