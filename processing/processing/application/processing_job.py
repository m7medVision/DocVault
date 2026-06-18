"""Processing job use case handler.

The handler is a thin orchestrator: :meth:`ProcessingJobHandler.handle` runs the
pipeline by delegating to small single-responsibility stage methods. The
observable behavior, side effects, and order are identical to the previous
inline implementation.
"""

import time
from dataclasses import dataclass

from opentelemetry import trace

from docvault_shared import telemetry
from docvault_shared.config import config
from docvault_shared.transport.publisher import QueuePublisher

from ..classifier import (
    generate_display_title,
    suggest_folder_path,
    suggestion_confidence,
)
from .ports import (
    ChunkerPort,
    ClassifierPort,
    EmbedderPort,
    PublisherFactory,
    PublisherPort,
    ReminderPublisherPort,
    RepositoryPort,
    TranslatorPort,
)


@dataclass
class PageObj:
    """Lightweight page DTO carrying the text the pipeline operates on."""

    page_number: int
    text: str


def _default_publisher_factory(connection: object) -> PublisherPort:
    """Build a QueuePublisher bound to the service connection."""
    return QueuePublisher(connection=connection)


@dataclass
class _JobContext:
    """Validated, hydrated job state shared across pipeline stages."""

    document_id: str
    version_id: str
    tenant_id: str
    org_id: str
    pages: list[dict]
    page_id_map: dict[int, str]
    searchable_pages: list[dict]
    source_page_objects: list[PageObj]
    message_language: str | None


class ProcessingJobHandler:
    """Handles document processing job execution (chunk + embed + classify + suggest)."""

    def __init__(
        self,
        chunker: ChunkerPort,
        embedder: EmbedderPort,
        classifier: ClassifierPort,
        repo: RepositoryPort,
        reminder_publisher: ReminderPublisherPort,
        connection: object,
        translator: TranslatorPort,
        publisher_factory: PublisherFactory = _default_publisher_factory,
    ):
        self.chunker = chunker
        self.embedder = embedder
        self.classifier = classifier
        self.repo = repo
        self.reminder_publisher = reminder_publisher
        self._connection = connection
        self.translator = translator
        self._publisher_factory = publisher_factory

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
            ctx = await self._build_context(message)
            if ctx is None:
                # Misrouted OCR job: rerouted to the OCR queue, nothing to do.
                return

            metadata, detected_language = await self._classify(ctx)
            metadata, translated_text_by_page = await self._translate(
                ctx, metadata, detected_language
            )
            metadata["language"] = detected_language

            page_objects = self._build_searchable_pages(ctx, translated_text_by_page)
            chunks, raw_embeddings = await self._chunk_and_embed(ctx, page_objects)
            await self._persist(ctx, chunks, raw_embeddings, metadata)
            await self._suggest(ctx, metadata)
            await self._publish_reminder(ctx, page_objects, metadata)

            await self.repo.update_processing_stage(ctx.document_id, "completed")
            await self.repo.update_document_status(ctx.document_id, "processed")

            self._record_telemetry(job_type, start_time, chunks, raw_embeddings)

    async def _build_context(self, message: dict) -> _JobContext | None:
        """Validate, hydrate (legacy), and reroute (misroute) the job message.

        Returns the assembled :class:`_JobContext`, or ``None`` when the message
        was a misrouted OCR job that has been republished to the OCR queue.
        """
        required_fields = ["document_id", "version_id", "tenant_id", "org_id"]
        for field in required_fields:
            if field not in message:
                raise ValueError(f"Missing required field: {field}")

        document_id = message["document_id"]
        version_id = message["version_id"]
        tenant_id = message["tenant_id"]
        org_id = message["org_id"]

        pages, raw_page_ids = await self._hydrate_pages(message)
        if pages is None:
            return None

        page_id_map: dict[int, str] = {}
        for i, page_data in enumerate(pages):
            if i < len(raw_page_ids):
                page_id_map[page_data["page_number"]] = raw_page_ids[i]

        searchable_pages = self._select_searchable_pages(message, pages, document_id)

        source_page_objects = [
            PageObj(p["page_number"], p["text"]) for p in pages if p.get("text")
        ]

        return _JobContext(
            document_id=document_id,
            version_id=version_id,
            tenant_id=tenant_id,
            org_id=org_id,
            pages=pages,
            page_id_map=page_id_map,
            searchable_pages=searchable_pages,
            source_page_objects=source_page_objects,
            message_language=message.get("language"),
        )

    async def _hydrate_pages(self, message: dict) -> tuple[list[dict] | None, list[str]]:
        """Resolve the page list and their ids.

        For legacy messages (no inline ``pages``) the pages are loaded from
        storage; a misrouted OCR job is rerouted and ``(None, [])`` is returned.
        Otherwise raises when pages cannot be resolved.
        """
        document_id = message["document_id"]
        version_id = message["version_id"]
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
                return None, []

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

        return pages, raw_page_ids

    def _select_searchable_pages(
        self,
        message: dict,
        pages: list[dict],
        document_id: str,
    ) -> list[dict]:
        """Drop low-confidence pages from the index set when possible."""
        low_confidence_pages = set(message.get("low_confidence_pages", []))
        if not low_confidence_pages:
            return pages

        filtered_pages = [
            page for page in pages if page["page_number"] not in low_confidence_pages
        ]
        if filtered_pages:
            telemetry.get_logger().info(
                "excluding_low_confidence_pages_from_index",
                document_id=document_id,
                excluded_pages=sorted(low_confidence_pages),
            )
            return filtered_pages

        telemetry.get_logger().warning(
            "all_pages_low_confidence_using_full_text_for_index",
            document_id=document_id,
            page_count=len(pages),
        )
        return pages

    async def _classify(self, ctx: _JobContext) -> tuple[dict, str]:
        """Classify the source text and resolve the detected language."""
        full_text = "\n\n".join(page.text for page in ctx.source_page_objects)
        metadata = await self.classifier.extract(
            document_id=ctx.document_id,
            text=full_text,
        )
        detected_language = (
            ctx.message_language or metadata.get("language") or "en"
        ).lower()
        return metadata, detected_language

    async def _translate(
        self,
        ctx: _JobContext,
        metadata: dict,
        detected_language: str,
    ) -> tuple[dict, dict[int, str]]:
        """Translate non-English documents and re-classify on the translation.

        Non-fatal: translation failures are logged and the original metadata is
        kept. Returns the (possibly updated) metadata and the translated text by
        page number.
        """
        translated_text_by_page: dict[int, str] = {}

        if not ctx.source_page_objects or detected_language.startswith("en"):
            return metadata, translated_text_by_page

        try:
            translation_result = await self.translator.translate_document(
                document_id=ctx.document_id,
                pages=ctx.source_page_objects,
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
                    await self.repo.update_page_translations(
                        document_id=ctx.document_id,
                        translations=translated_text_by_page,
                    )
                    telemetry.get_logger().info(
                        "translation_completed",
                        document_id=ctx.document_id,
                        translated_page_count=len(translated_text_by_page),
                    )

                    translated_full_text = "\n\n".join(
                        translated_text_by_page.get(page.page_number) or page.text
                        for page in ctx.source_page_objects
                    )
                    metadata = await self.classifier.extract(
                        document_id=ctx.document_id,
                        text=translated_full_text,
                    )
        except Exception as translation_err:
            telemetry.get_logger().warning(
                "translation_failed_non_fatal",
                document_id=ctx.document_id,
                error=str(translation_err),
            )

        return metadata, translated_text_by_page

    def _build_searchable_pages(
        self,
        ctx: _JobContext,
        translated_text_by_page: dict[int, str],
    ) -> list[PageObj]:
        """Build the page objects to index, preferring translated text."""
        return [
            PageObj(
                p["page_number"],
                translated_text_by_page.get(p["page_number"], p["text"]),
            )
            for p in ctx.searchable_pages
            if p.get("text")
        ]

    async def _chunk_and_embed(
        self,
        ctx: _JobContext,
        page_objects: list[PageObj],
    ) -> tuple[list, list]:
        """Chunk the pages and embed the resulting chunks."""
        await self.repo.update_processing_stage(ctx.document_id, "processing_running")

        chunk_result = self.chunker.chunk_ocr_result(
            ctx.document_id,
            page_objects,
            page_ids=ctx.page_id_map,
        )
        chunks = chunk_result.chunks

        raw_embeddings: list = []
        if chunks:
            raw_embeddings = await self.embedder(chunks)

        return chunks, raw_embeddings

    async def _persist(
        self,
        ctx: _JobContext,
        chunks: list,
        raw_embeddings: list,
        metadata: dict,
    ) -> None:
        """Persist chunks, metadata rows, and the classified document metadata."""
        await self.repo.update_processing_stage(ctx.document_id, "indexing")

        await self.repo.delete_by_document(ctx.document_id)

        if chunks:
            await self.repo.save_chunks(
                document_id=ctx.document_id,
                chunks=chunks,
                embeddings=raw_embeddings,
            )

        await self.repo.upsert_metadata_rows(
            document_id=ctx.document_id,
            metadata=metadata,
        )

        await self.repo.update_document_metadata(
            document_id=ctx.document_id,
            doc_type=metadata.get("doc_type"),
            language=metadata.get("language"),
        )

    async def _suggest(self, ctx: _JobContext, metadata: dict) -> None:
        """Write the non-destructive folder/filename suggestion."""
        original_title = await self.repo.get_document_title(ctx.document_id) or "document"
        suggested_filename = generate_display_title(metadata, original_title)
        await self.repo.suggest_organization(
            document_id=ctx.document_id,
            tenant_id=ctx.tenant_id,
            org_id=ctx.org_id,
            suggested_folder_name=suggest_folder_path(metadata),
            suggested_filename=suggested_filename,
            suggestion_confidence=suggestion_confidence(metadata),
        )

    async def _publish_reminder(
        self,
        ctx: _JobContext,
        page_objects: list[PageObj],
        metadata: dict,
    ) -> None:
        """Publish the reminder job (non-fatal on failure)."""
        try:
            await self.reminder_publisher.publish_after_processing(
                document_id=ctx.document_id,
                tenant_id=ctx.tenant_id,
                org_id=ctx.org_id,
                pages=page_objects,
                metadata=metadata,
            )
        except Exception as reminder_err:
            telemetry.get_logger().warning(
                "reminder_publish_failed_non_fatal",
                document_id=ctx.document_id,
                error=str(reminder_err),
            )

    def _record_telemetry(
        self,
        job_type: str,
        start_time: float,
        chunks: list,
        raw_embeddings: list,
    ) -> None:
        """Record the job-completion telemetry."""
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
        if not hasattr(self.repo, "get_document_pages"):
            return []

        return await self.repo.get_document_pages(document_id, version_id)

    def _is_misrouted_ocr_job(self, message: dict) -> bool:
        """Detect OCR jobs that were accidentally published to the processing queue."""
        return bool(message.get("storage_key") and message.get("mime_type"))

    def _republish_to_ocr_queue(self, message: dict) -> None:
        """Send a misrouted OCR job back to the correct queue."""
        repaired_message = dict(message)
        repaired_message["retry_count"] = 0
        self._publisher_factory(self._connection).publish(
            queue=config.rabbitmq_queue_ocr,
            message=repaired_message,
        )
