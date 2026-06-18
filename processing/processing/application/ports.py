"""Typed ports (structural Protocols) for the processing pipeline.

These describe only the methods the :class:`ProcessingJobHandler` actually
calls on each injected collaborator, so the composition root in ``main.py`` can
wire concrete implementations while the handler depends on narrow interfaces
rather than ``Any``.
"""

from __future__ import annotations

from typing import Any, Protocol


class ChunkerPort(Protocol):
    """Splits page text into embeddable chunks."""

    def chunk_ocr_result(
        self,
        document_id: str,
        pages: list,
        page_ids: dict[int, str] | None = None,
    ) -> Any:
        """Return a chunk result exposing ``.chunks``."""
        ...


class EmbedderPort(Protocol):
    """Turns a list of chunks into embedding vectors."""

    async def __call__(self, chunks: list) -> list[list[float]]:
        """Embed the given chunks."""
        ...


class ClassifierPort(Protocol):
    """Extracts document metadata from text."""

    async def extract(self, document_id: str, text: str) -> dict:
        """Classify and extract metadata for the document text."""
        ...


class TranslatorPort(Protocol):
    """Translates a document's pages into the target language."""

    async def translate_document(
        self,
        document_id: str,
        pages: list,
        source_language: str,
    ) -> dict:
        """Translate the document pages, returning the translation result dict."""
        ...


class RepositoryPort(Protocol):
    """Persistence + document-state operations used by the pipeline.

    The processing service backs this with a single ``PGVectorRepository`` that
    owns both the pgvector chunk store and the documents/pages tables, so one
    port describes every read/write the handler performs.
    """

    async def get_document_pages(self, document_id: str, version_id: str) -> list[dict]:
        """Load persisted OCR pages for a document version (legacy hydration)."""
        ...

    async def update_processing_stage(self, document_id: str, stage: str) -> None:
        """Advance the document's processing_stage."""
        ...

    async def update_page_translations(
        self,
        document_id: str,
        translations: dict[int, str],
    ) -> None:
        """Persist translated page text."""
        ...

    async def delete_by_document(self, document_id: str) -> None:
        """Remove existing chunks for the document before re-indexing."""
        ...

    async def save_chunks(
        self,
        document_id: str,
        chunks: list,
        embeddings: list[list[float]],
    ) -> Any:
        """Persist chunks with their embeddings."""
        ...

    async def upsert_metadata_rows(self, document_id: str, metadata: dict) -> None:
        """Write extracted metadata rows."""
        ...

    async def update_document_metadata(
        self,
        document_id: str,
        doc_type: str | None = None,
        language: str | None = None,
    ) -> None:
        """Write classified doc_type/language onto the document row."""
        ...

    async def update_document_status(self, document_id: str, status: str) -> None:
        """Update the document's status."""
        ...

    async def get_document_title(self, document_id: str) -> str | None:
        """Return the document's current title."""
        ...

    async def suggest_organization(
        self,
        document_id: str,
        tenant_id: str,
        org_id: str,
        suggested_folder_name: str,
        suggested_filename: str,
        suggestion_confidence: float,
    ) -> None:
        """Persist the non-destructive filing suggestion."""
        ...


class ReminderPublisherPort(Protocol):
    """Publishes reminder jobs after processing completes."""

    async def publish_after_processing(
        self,
        document_id: str,
        tenant_id: str,
        org_id: str,
        pages: list,
        metadata: dict,
    ) -> Any:
        """Publish a reminder job for the processed document."""
        ...


class PublisherPort(Protocol):
    """Publishes a raw message to a named queue (used to reroute misroutes)."""

    def publish(self, queue: str, message: dict) -> None:
        """Publish ``message`` to ``queue``."""
        ...


class PublisherFactory(Protocol):
    """Builds a :class:`PublisherPort` bound to the service connection.

    Injected so the handler does not import or construct ``QueuePublisher``
    directly when it needs to reroute a misrouted OCR job.
    """

    def __call__(self, connection: Any) -> PublisherPort:
        """Return a publisher bound to ``connection``."""
        ...
