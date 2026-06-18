"""OCR job use case handler."""

import time

from opentelemetry import trace

from docvault_shared.models import OCRJob, OCRResult
from docvault_shared import telemetry

from .ports import MessagePublisher, OCRClient, OCRPersistence

REQUIRED_FIELDS = (
    "document_id",
    "version_id",
    "storage_key",
    "tenant_id",
    "org_id",
)


class OCRJobHandler:
    """Handles OCR job execution."""

    def __init__(
        self,
        ocr_service: OCRClient,
        storage_service: OCRPersistence,
        publisher: MessagePublisher,
        processing_queue: str,
    ):
        self.ocr = ocr_service
        self.storage = storage_service
        self._publisher = publisher
        self._processing_queue = processing_queue

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
            job = self._parse_job(message)

            await self.storage.update_document_status(
                job.document_id,
                "processing",
            )

            result = await self.ocr.process_document(job)

            low_confidence_pages = self._flag_low_confidence(job, result)

            page_ids = await self.storage.save_ocr_results(result)

            await self.storage.update_document_status(
                job.document_id,
                "processing",
            )

            self._publish_processing(job, result, page_ids, low_confidence_pages)

            duration = time.time() - start_time
            telemetry.record_job(job_type, True, duration, {"page_count": len(result.pages)})

    def _parse_job(self, message: dict) -> OCRJob:
        """Validate required fields and build the OCRJob value object."""
        for field in REQUIRED_FIELDS:
            if field not in message:
                raise ValueError(f"Missing required field: {field}")

        return OCRJob(
            document_id=message["document_id"],
            version_id=message["version_id"],
            storage_key=message["storage_key"],
            mime_type=message.get("mime_type", "application/octet-stream"),
            tenant_id=message["tenant_id"],
            org_id=message["org_id"],
            language=message.get("language"),
        )

    def _flag_low_confidence(self, job: OCRJob, result: OCRResult) -> list[int]:
        """Flag low-confidence pages and log when any are found."""
        low_confidence_pages = self.ocr.flag_low_confidence_pages(result)
        if low_confidence_pages:
            telemetry.get_logger().warning(
                "low_confidence_pages_found",
                document_id=job.document_id,
                pages=low_confidence_pages,
            )
        return low_confidence_pages

    def _publish_processing(
        self,
        job: OCRJob,
        result: OCRResult,
        page_ids: list[str],
        low_confidence_pages: list[int],
    ) -> None:
        """Publish the downstream processing message."""
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

        self._publisher.publish(queue=self._processing_queue, message=processing_message)
