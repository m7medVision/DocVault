"""Processing job use case handler."""

import time
from typing import Any

from opentelemetry import trace

from docvault_shared import telemetry
from docvault_shared.config import config
from docvault_shared.transport.publisher import QueuePublisher

from ..classifier import generate_display_title


CATEGORY_FOLDER_BY_TYPE = {
    "invoice": "Invoices",
    "contract": "Contracts",
    "identity": "Identity",
    "warranty": "Warranties",
    "receipt": "Receipts",
}


class ProcessingJobHandler:
    """Handles document processing job execution (chunk + embed + classify + suggest)."""

    def __init__(
        self,
        chunker: Any,
        embedder: Any,
        classifier: Any,
        pg_repo: Any,
        ocr_repo: Any,
        reminder_publisher: Any,
        connection: Any,
        translator: Any | None = None,
    ):
        self.chunker = chunker
        self.embedder = embedder
        self.classifier = classifier
        self.pg_repo = pg_repo
        self.ocr_repo = ocr_repo
        self.reminder_publisher = reminder_publisher
        self._connection = connection
        if translator is None:
            from ..translation import document_translator

            translator = document_translator
        self.translator = translator

    async def handle(self, message: dict) -> None:
        """Process a single document processing job."""
        start_time = time.time()
        job_type = "processing"

        with telemetry.start_span(
            "process_job",
            kind=trace.SpanKind.CONSUMER,
            attributes={
                "job.type": job_type,
                "document.id": message.get("document_id"),
                "tenant.id": message.get("tenant_id"),
            },
        ) as _:
            required_fields = ["document_id", "version_id", "tenant_id", "org_id"]
            for field in required_fields:
                if field not in message:
                    raise ValueError(f"Missing required field: {field}")

            document_id = message["document_id"]
            version_id = message["version_id"]
            tenant_id = message["tenant_id"]
            org_id = message["org_id"]
            pages = message.get("pages")
            raw_page_ids = message.get("page_ids") or []

            if not pages:
                pages = await self._load_pages_for_legacy_message(document_id, version_id)
                if not pages and self._is_misrouted_ocr_job(message):
                    self._republish_to_ocr_queue(message)
                    telemetry.get_logger().warning(
                        "rerouted_misrouted_ocr_job_from_processing_queue",
                        document_id=document_id,
                        version_id=version_id,
                        source_queue=config.rabbitmq_queue_processing,
                        destination_queue=config.rabbitmq_queue_ocr,
                    )
                    return

                if not pages:
                    raise ValueError("Missing required field: pages")

                if not raw_page_ids:
                    raw_page_ids = [page["id"] for page in pages]

                telemetry.get_logger().info(
                    "loaded_pages_for_legacy_processing_message",
                    document_id=document_id,
                    version_id=version_id,
                    page_count=len(pages),
                )

            if not raw_page_ids:
                raw_page_ids = [page.get("id") for page in pages if page.get("id")]

            low_confidence_pages = set(message.get("low_confidence_pages", []))

            page_id_map: dict[int, str] = {}
            for i, page_data in enumerate(pages):
                if i < len(raw_page_ids):
                    page_id_map[page_data["page_number"]] = raw_page_ids[i]

            searchable_pages = pages
            if low_confidence_pages:
                filtered_pages = [
                    page for page in pages if page["page_number"] not in low_confidence_pages
                ]
                if filtered_pages:
                    searchable_pages = filtered_pages
                    telemetry.get_logger().info(
                        "excluding_low_confidence_pages_from_index",
                        document_id=document_id,
                        excluded_pages=sorted(low_confidence_pages),
                    )
                else:
                    telemetry.get_logger().warning(
                        "all_pages_low_confidence_using_full_text_for_index",
                        document_id=document_id,
                        page_count=len(pages),
                    )

            class PageObj:
                def __init__(self, page_number, text):
                    self.page_number = page_number
                    self.text = text

            source_page_objects = [
                PageObj(p["page_number"], p["text"]) for p in pages if p.get("text")
            ]

            full_text = "\n\n".join(page.text for page in source_page_objects)
            metadata = await self.classifier.extract(
                document_id=document_id,
                text=full_text,
            )

            detected_language = (
                message.get("language") or metadata.get("language") or "en"
            ).lower()
            translated_text_by_page: dict[int, str] = {}

            if source_page_objects and not detected_language.startswith("en"):
                try:
                    translation_result = await self.translator.translate_document(
                        document_id=document_id,
                        pages=source_page_objects,
                        source_language=detected_language,
                    )

                    if translation_result.get("is_translation") and translation_result.get("pages"):
                        translated_pages = translation_result["pages"]
                        translated_text_by_page = {
                            page_number: page_data["translated_text"]
                            for page_number, page_data in translated_pages.items()
                            if page_data.get("translated_text")
                        }

                        if translated_text_by_page:
                            await self.pg_repo.update_page_translations(
                                document_id=document_id,
                                translations=translated_text_by_page,
                            )
                            telemetry.get_logger().info(
                                "translation_completed",
                                document_id=document_id,
                                translated_page_count=len(translated_text_by_page),
                            )

                            translated_full_text = "\n\n".join(
                                translated_text_by_page.get(page.page_number) or page.text
                                for page in source_page_objects
                            )
                            metadata = await self.classifier.extract(
                                document_id=document_id,
                                text=translated_full_text,
                            )
                except Exception as translation_err:
                    telemetry.get_logger().warning(
                        "translation_failed_non_fatal",
                        document_id=document_id,
                        error=str(translation_err),
                    )

            metadata["language"] = detected_language

            page_objects = [
                PageObj(
                    p["page_number"],
                    translated_text_by_page.get(p["page_number"], p["text"]),
                )
                for p in searchable_pages
                if p.get("text")
            ]

            await self.pg_repo.update_processing_stage(document_id, "processing_running")

            chunk_result = self.chunker.chunk_ocr_result(
                document_id,
                page_objects,
                page_ids=page_id_map,
            )
            chunks = chunk_result.chunks

            raw_embeddings = []
            if chunks:
                raw_embeddings = await self.embedder(chunks)

            await self.pg_repo.update_processing_stage(document_id, "indexing")

            await self.pg_repo.delete_by_document(document_id)

            if chunks:
                await self.pg_repo.save_chunks(
                    document_id=document_id,
                    chunks=chunks,
                    embeddings=raw_embeddings,
                )

            await self.pg_repo.upsert_metadata_rows(
                document_id=document_id,
                metadata=metadata,
            )

            await self.ocr_repo.update_document_metadata(
                document_id=document_id,
                doc_type=metadata.get("doc_type"),
                language=metadata.get("language"),
            )

            original_title = await self.pg_repo.get_document_title(document_id) or "document"
            desired_title = generate_display_title(metadata, original_title)
            await self.pg_repo.auto_organize(
                document_id=document_id,
                tenant_id=tenant_id,
                org_id=org_id,
                doc_type=metadata.get("doc_type", "other"),
                desired_title=desired_title,
            )

            try:
                await self.reminder_publisher.publish_after_processing(
                    document_id=document_id,
                    tenant_id=tenant_id,
                    org_id=org_id,
                    pages=page_objects,
                    metadata=metadata,
                )
            except Exception as reminder_err:
                telemetry.get_logger().warning(
                    "reminder_publish_failed_non_fatal",
                    document_id=document_id,
                    error=str(reminder_err),
                )

            await self.pg_repo.update_processing_stage(document_id, "completed")
            await self.ocr_repo.update_document_status(document_id, "processed")

            duration = time.time() - start_time
            telemetry.record_job(
                job_type,
                True,
                duration,
                {
                    "chunk_count": len(chunks),
                    "embedding_count": len(raw_embeddings),
                },
            )

    async def _load_pages_for_legacy_message(
        self,
        document_id: str,
        version_id: str,
    ) -> list[dict]:
        """Load OCR pages from storage for legacy queue messages."""
        if not hasattr(self.ocr_repo, "get_document_pages"):
            return []

        return await self.ocr_repo.get_document_pages(document_id, version_id)

    def _is_misrouted_ocr_job(self, message: dict) -> bool:
        """Detect OCR jobs that were accidentally published to the processing queue."""
        return bool(message.get("storage_key") and message.get("mime_type"))

    def _republish_to_ocr_queue(self, message: dict) -> None:
        """Send a misrouted OCR job back to the correct queue."""
        repaired_message = dict(message)
        repaired_message["retry_count"] = 0
        QueuePublisher(connection=self._connection).publish(
            queue=config.rabbitmq_queue_ocr,
            message=repaired_message,
        )
