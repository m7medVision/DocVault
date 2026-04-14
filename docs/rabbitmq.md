# RabbitMQ Flow

## Queue Topology

- `docvault.ocr.jobs`: backend publishes new uploads, OCR service consumes them.
- `docvault.processing.jobs`: OCR publishes extracted pages, processing service consumes them.
- `docvault.reminder.jobs`: processing publishes reminder work, reminder service consumes it.
- Each main queue uses a matching dead-letter queue at `<queue>.dlq`.

## Stage Ownership

1. Backend creates the document and version records, then publishes an OCR job.
2. OCR reads the source file, stores OCR pages, then publishes a processing job.
3. Processing chunks, classifies, and indexes the document, then publishes a reminder job.
4. Reminder extracts reminder rules and stores them in Postgres.

## Retry Rule

- Consumers keep retry state in the message payload with `retry_count`.
- On retryable failure, the consumer waits for backoff, republishes the message with `retry_count + 1`, and acknowledges the original delivery.
- On permanent failure or max retries, the consumer publishes one structured DLQ envelope and acknowledges the original delivery.

## DLQ Envelope

DLQ messages use a wrapper so failures are inspectable:

```json
{
  "original_queue": "docvault.reminder.jobs",
  "original_message": {"...": "..."},
  "error": "...",
  "retry_count": 3,
  "failed_at": "2026-04-14T16:00:00Z"
}
```

For malformed payloads, `original_message` can be `null` and `original_body` contains the raw body.
