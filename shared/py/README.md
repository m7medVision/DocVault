# docvault-shared — Shared Python Utilities for DocVault

Common Python utilities shared between DocVault services (OCR, Processing).

## Contents

- `docvault_shared/config` — Configuration loading from environment variables
- `docvault_shared/models` — Shared dataclasses (OCRJob, TextChunk, etc.)
- `docvault_shared/telemetry` — OpenTelemetry + Sentry instrumentation
- `docvault_shared/database` — SQLAlchemy models and persistence (Document, DocumentPage)
- `docvault_shared/transport` — RabbitMQ connection, consumer, publisher

## Usage

```python
from docvault_shared import config
from docvault_shared.transport import RabbitMQConnection, QueueConsumer
from docvault_shared.database import get_ocr_persistence
from docvault_shared import telemetry
```

## Services Using This

- `docvault-ocr` — OCR service
- `docvault-processing` — Processing pipeline service

## Development

```bash
cd shared/py
uv sync --all-extras
uv run ruff check .
uv run mypy .
```
