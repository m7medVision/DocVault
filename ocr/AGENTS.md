# ocr/AGENTS.md — OCR service

Part of the docvault monorepo. Read root **[../AGENTS.md](../AGENTS.md)** first (commit
identity, conventions).

## Stack

Python · **uv** · **Mistral OCR cloud API** (`mistral-ocr-2503`, language-agnostic so it
handles Arabic + English) · **RabbitMQ** consumer · **boto3 → MinIO/S3** for file bytes.
Shared config/transport from `shared/py` (`docvault_shared`).

## Layout

```
ocr/ocr/
  main.py            entrypoint — RabbitMQ consumer (console script docvault-ocr)
  ocr.py             MistralOCRClient: download → OCR → parse pages + confidence
  storage.py         S3/MinIO client
  application/       OCRJobHandler (persist pages, publish to processing queue)
tests/
```

## Commands & gate

```sh
just dev-ocr-install        # uv sync
just dev-ocr                # run the consumer
cd ocr && uv run ruff check . && uv run pytest   # gate (run before commit)
```

## Conventions

- **Queue consumer, not an HTTP server.** Consumes `docvault.ocr.jobs`, runs OCR, persists
  `document_pages`, then **publishes to the processing queue**. Each queue has a `.dlq`.
- Low-confidence pages are flagged and excluded from downstream indexing.
- `uv` for everything (`uv sync`, `uv run`). Config + RabbitMQ transport come from
  `docvault_shared` (see [../shared/AGENTS.md](../shared/AGENTS.md)).
