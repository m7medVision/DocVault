# OCR — Document OCR Service (Python)

Python-based OCR service for DocVault using Mistral OCR API.

## Responsibilities

- RabbitMQ consumer for OCR jobs (`docvault.ocr.jobs`)
- Mistral OCR API integration (text extraction from document images/pages)
- Low-confidence page flagging
- Persistence of OCR results to PostgreSQL
- Publishing processing jobs to `docvault.processing.jobs`

## Tech Stack

- Python 3.11
- Mistral OCR 2503 API
- PostgreSQL 16 + pgvector
- RabbitMQ
- MinIO (S3-compatible object storage)
- OpenTelemetry (tracing)

## Architecture Overview

```
docvault.ocr.jobs
       ↓
  QueueConsumer
       ↓
  OCRJobHandler
       ↓
  MistralOCRClient
       ↓
  OCRPersistence → PostgreSQL
       ↓
  QueuePublisher → docvault.processing.jobs
```

## Queue Contracts

**Input queue**: `docvault.ocr.jobs`
**Output queue**: `docvault.processing.jobs`
**DLQ**: `docvault.ocr.jobs.dlq`

## How to Run

```bash
# Install dependencies
cd ocr && uv sync --all-extras

# Run OCR consumer
uv run python -m ocr.main
```

## Environment Variables

Key env vars:
- `RABBITMQ_URL` — broker
- `DATABASE_URL` — PostgreSQL connection
- `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET`
- `MISTRAL_API_KEY` — OCR
- `OTEL_EXPORTER_ENDPOINT` — tracing collector

## Validation

```bash
uv run ruff check .
uv run mypy .
```