"""Protocol ports describing the dependencies the OCR use case needs.

These are structural (typing.Protocol) interfaces that describe only the
methods actually used by the application layer. Concrete implementations
(MistralOCRClient, MinIOClient, OCRPersistence, QueuePublisher) satisfy them
by shape, so the use case depends on abstractions rather than concretions.
"""

from typing import Protocol, runtime_checkable

from docvault_shared.models import OCRJob, OCRResult


@runtime_checkable
class BlobStore(Protocol):
    """Object storage port: just the read used during OCR."""

    async def get_object(self, storage_key: str) -> bytes:
        """Download an object's bytes by storage key."""
        ...


@runtime_checkable
class OCRClient(Protocol):
    """OCR engine port: run OCR over a job and flag low-confidence pages."""

    async def process_document(self, job: OCRJob) -> OCRResult:
        """Run OCR over the job's document and return the parsed result."""
        ...

    def flag_low_confidence_pages(self, result: OCRResult) -> list[int]:
        """Return the page numbers whose confidence is below threshold."""
        ...


@runtime_checkable
class OCRPersistence(Protocol):
    """Persistence port for document status and OCR page rows."""

    async def update_document_status(self, document_id: str, status: str) -> None:
        """Update a document's processing status."""
        ...

    async def save_ocr_results(self, result: OCRResult) -> list[str]:
        """Persist OCR pages and return their page ids."""
        ...


@runtime_checkable
class MessagePublisher(Protocol):
    """Outbound message port for handing work to the next queue."""

    def publish(self, queue: str, message: dict) -> None:
        """Publish a message to the named queue."""
        ...
