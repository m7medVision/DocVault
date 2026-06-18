"""Mistral OCR API client."""

import structlog
import uuid
from datetime import datetime

from mistralai.client import Mistral
from mistralai.client.models import OCRResponse

from docvault_shared.config import config
from docvault_shared.models import OCRJob, OCRPage, OCRResult

from .application.ports import BlobStore

logger = structlog.get_logger(__name__)


def _document_page_id(document_id: str, version_id: str, page_number: int) -> str:
    """Build a deterministic UUID for a document page row."""
    return str(uuid.uuid5(uuid.NAMESPACE_URL, f"{document_id}:{version_id}:{page_number}"))


class MistralOCRClient:
    """Client for Mistral OCR API."""

    def __init__(self, storage: BlobStore):
        """Initialize Mistral OCR client with its blob storage dependency."""
        self.client = Mistral(api_key=config.mistral_api_key)
        self.model = "mistral-ocr-2503"
        self.storage = storage

    async def process_document(self, job: OCRJob) -> OCRResult:
        """Process a document through Mistral OCR."""
        logger.info(
            "starting_ocr",
            document_id=job.document_id,
            storage_key=job.storage_key,
            mime_type=job.mime_type,
        )

        try:
            document_bytes = await self._download(job)
            uploaded_file = self._upload_to_mistral(job, document_bytes)
            try:
                response = self._run_ocr(job, uploaded_file)
            finally:
                self._delete_mistral_upload(job, uploaded_file.id)

            pages = self._parse_ocr_response(response, job.document_id, job.version_id)
            result = OCRResult(
                document_id=job.document_id,
                version_id=job.version_id,
                pages=pages,
                completed_at=datetime.utcnow(),
            )

            logger.info(
                "ocr_completed",
                document_id=job.document_id,
                page_count=len(pages),
            )

            return result

        except Exception as e:
            logger.error(
                "ocr_failed",
                document_id=job.document_id,
                error=str(e),
            )
            raise

    async def _download(self, job: OCRJob) -> bytes:
        """Fetch the document bytes from blob storage."""
        document_bytes = await self.storage.get_object(job.storage_key)
        logger.info(
            "document_downloaded",
            document_id=job.document_id,
            size=len(document_bytes),
        )
        return document_bytes

    def _upload_to_mistral(self, job: OCRJob, document_bytes: bytes):
        """Upload the document bytes to Mistral and return the file handle."""
        file_name = job.storage_key.rsplit("/", 1)[-1] or f"{job.document_id}.bin"
        uploaded_file = self.client.files.upload(
            file={
                "file_name": file_name,
                "content": document_bytes,
                "content_type": job.mime_type,
            },
            purpose="ocr",
        )
        logger.info(
            "document_uploaded_to_mistral",
            document_id=job.document_id,
            file_id=uploaded_file.id,
        )
        return uploaded_file

    def _run_ocr(self, job: OCRJob, uploaded_file) -> OCRResponse:
        """Run OCR over the uploaded file via its signed URL."""
        signed_url = self.client.files.get_signed_url(file_id=uploaded_file.id)
        return self.client.ocr.process(
            model=self.model,
            document={
                "type": "document_url",
                "document_url": signed_url.url,
            },
            id=job.version_id,
        )

    def _delete_mistral_upload(self, job: OCRJob, file_id: str) -> None:
        """Best-effort cleanup of the temporary Mistral upload."""
        try:
            self.client.files.delete(file_id=file_id)
        except Exception as cleanup_error:
            logger.warning(
                "failed_to_delete_mistral_upload",
                document_id=job.document_id,
                file_id=file_id,
                error=str(cleanup_error),
            )

    def _parse_ocr_response(
        self,
        response: OCRResponse,
        document_id: str,
        version_id: str,
    ) -> list[OCRPage]:
        """Parse Mistral OCR response into our page model."""
        pages = []

        if hasattr(response, "pages"):
            for idx, page_data in enumerate(response.pages):
                if isinstance(page_data, dict):
                    page_index = page_data.get("index")
                    page_number = page_data.get("page_number", idx + 1)
                    if page_index is not None:
                        page_number = page_index + 1
                    page_text = page_data.get("text", page_data.get("markdown", ""))
                    confidence = page_data.get("confidence", 1.0)
                else:
                    page_index = getattr(page_data, "index", None)
                    page_number = getattr(page_data, "page_number", idx + 1)
                    if page_index is not None:
                        page_number = page_index + 1
                    page_text = getattr(page_data, "text", getattr(page_data, "markdown", ""))
                    confidence = getattr(page_data, "confidence", 1.0)

                page = OCRPage(
                    page_number=page_number,
                    text=page_text,
                    confidence=confidence,
                    model=self.model,
                )
                pages.append(page)
        else:
            pages.append(
                OCRPage(
                    page_number=1,
                    text=str(response),
                    confidence=0.5,
                    model=self.model,
                )
            )

        return pages

    def flag_low_confidence_pages(self, result: OCRResult) -> list[int]:
        """Flag pages with confidence below threshold."""
        low_confidence = []
        for page in result.pages:
            if page.confidence < config.min_confidence_threshold:
                low_confidence.append(page.page_number)
                logger.warning(
                    "low_confidence_page",
                    document_id=result.document_id,
                    page=page.page_number,
                    confidence=page.confidence,
                    threshold=config.min_confidence_threshold,
                )
        return low_confidence
