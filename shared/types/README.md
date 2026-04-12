# Shared Types

Shared type definitions used across the monorepo for cross-application contracts.

## Contents

### TypeScript (web/mobile consumers)

- `document.ts` — Document, DocumentVersion, DocumentMetadata, DocumentPage types
- `metadata.ts` — Metadata key-value types, extraction/correction types
- `user.ts` — User, Tenant, Organization, Membership types
- `reminder.ts` — ReminderRule, ReminderEvent types
- `search.ts` — SearchRequest, SearchResult, ChunkResult types
- `common.ts` — Pagination, error response, shared utility types

### Go (backend/worker/processing consumers)

- `queue.go` — Queue message schemas for OCR, Processing, and Reminder jobs

## Queue Message Schemas

The queue message types in `queue.go` are the canonical schemas for inter-service
communication via message queues.

### OCRJob

Published to `docvault.ocr.jobs` after a document is uploaded.

```json
{
  "document_id": "doc-123",
  "user_id": "user-456",
  "file_key": "files/doc-123.pdf",
  "language": "en",
  "tenant_id": "tenant-789"
}
```

### ProcessingJob

Published to `docvault.processing.jobs` after OCR completes.

```json
{
  "document_id": "doc-123",
  "user_id": "user-456",
  "language": "en",
  "tenant_id": "tenant-789",
  "total_pages": 10
}
```

### ReminderJob

Published to `docvault.reminder.jobs` for reminder extraction.

```json
{
  "document_id": "doc-123",
  "user_id": "user-456",
  "extracted_text": "Meeting tomorrow at 3pm",
  "language": "en",
  "tenant_id": "tenant-789"
}
```

## Usage (Go)

```go
import "github.com/docvault/shared/types"

// Use types.OCRJob, types.ProcessingJob, types.ReminderJob
```

## Building TypeScript

```bash
cd shared/types
npm install
npm run build
```
