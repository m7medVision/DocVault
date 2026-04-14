"""OCR job use case handler."""

import time
from typing import Any

from opentelemetry import trace

from docvault_shared.config import config
from docvault_shared.models import OCRJob
from docvault_shared import telemetry
from docvault_shared.transport.publisher import QueuePublisher


class OCRJobHandler:
    """Handles OCR job execution."""

    def __init__(
        self,
        ocr_service: Any,
        storage_service: Any,
        connection: Any,
    ):
        self.ocr = ocr_service
        self.storage = storage_service
        self._connection = connection

    async def handle(self, message: dict) -> None:
        """Process a single OCR job message."""
        start_time = time.time()
        job_type = "ocr"

        with telemetry.start_span(
            "process_ocr_job",
            kind=trace.SpanKind.CONSUMER,
            attributes={
                "job.type": job_type,
                "document.id": message.get("document_id"),
                "tenant.id": message.get("tenant_id"),
            },
        ) as _:
            required_fields = [
                "document_id",
                "version_id",
                "storage_key",
                "tenant_id",
                "org_id",
            ]
            for field in required_fields:
                if field not in message:
                    raise ValueError(f"Missing required field: {field}")

            job = OCRJob(
                document_id=message["document_id"],
                version_id=message["version_id"],
                storage_key=message["storage_key"],
                mime_type=message.get("mime_type", "application/octet-stream"),
                tenant_id=message["tenant_id"],
                org_id=message["org_id"],
                language=message.get("language"),
            )

            await self.storage.update_document_status(
                job.document_id,
                "processing",
            )

            result = await self.ocr.process_document(job)

            low_confidence_pages = self.ocr.flag_low_confidence_pages(result)
            if low_confidence_pages:
                telemetry.get_logger().warning(
                    "low_confidence_pages_found",
                    document_id=job.document_id,
                    pages=low_confidence_pages,
                )

            page_ids = await self.storage.save_ocr_results(result)

            await self.storage.update_document_status(
                job.document_id,
                "processing",
            )

            processing_message = {
                "document_id": job.document_id,
                "version_id": job.version_id,
                "tenant_id": job.tenant_id,
                "org_id": job.org_id,
                "language": job.language,
                "retry_count": 0,
                "page_ids": page_ids,
                "pages": [
                    {
                        "page_number": p.page_number,
                        "text": p.text,
                        "confidence": p.confidence,
                        "model": p.model,
                    }
                    for p in result.pages
                ],
                "low_confidence_pages": low_confidence_pages,
            }

            publisher = QueuePublisher(connection=self._connection)
            publisher.publish(queue=config.rabbitmq_queue_processing, message=processing_message)

            duration = time.time() - start_time
            telemetry.record_job(job_type, True, duration, {"page_count": len(result.pages)})
