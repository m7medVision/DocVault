"""Page-aware text chunking service for document processing."""

import hashlib
import structlog
from dataclasses import dataclass
from typing import Iterator, Optional

from docvault_shared.config import config
from docvault_shared.models import ChunkResult, TextChunk

logger = structlog.get_logger(__name__)


@dataclass
class ChunkConfig:
    """Configuration for chunking strategy."""

    chunk_size: int
    chunk_overlap: int
    min_chunk_size: int
    max_chunk_size: int
    respect_sentences: bool
    respect_paragraphs: bool


class ChunkingService:
    """Service for splitting text into overlapping chunks for embedding."""

    def __init__(self, config_override: Optional[ChunkConfig] = None):
        """Initialize the chunking service with configuration."""
        if config_override:
            self.config = config_override
        else:
            self.config = ChunkConfig(
                chunk_size=config.chunk_size,
                chunk_overlap=config.chunk_overlap,
                min_chunk_size=50,
                max_chunk_size=1000,
                respect_sentences=True,
                respect_paragraphs=True,
            )

        self._sentence_endings = ".!?"

    def chunk_text(
        self,
        text: str,
        page_id: str,
        page_number: int,
        document_id: Optional[str] = None,
    ) -> list[TextChunk]:
        """Split text into overlapping chunks with page awareness."""
        if not text or not text.strip():
            logger.debug("empty_text_skipped", page_id=page_id)
            return []

        chunks = []
        start = 0
        chunk_index = 0

        text_length = len(text)

        while start < text_length:
            end = start + self.config.chunk_size

            if end >= text_length:
                end = text_length
                chunk_text = text[start:end].strip()

                if len(chunk_text) >= self.config.min_chunk_size:
                    chunk = self._create_chunk(
                        text=chunk_text,
                        page_id=page_id,
                        page_number=page_number,
                        chunk_index=chunk_index,
                        start_pos=start,
                        end_pos=end,
                    )
                    chunks.append(chunk)
                break

            split_point = self._find_split_point(text, start, end)

            chunk_text = text[start:split_point].strip()

            if len(chunk_text) < self.config.min_chunk_size:
                split_point = min(start + self.config.min_chunk_size, text_length)
                chunk_text = text[start:split_point].strip()

            if len(chunk_text) > 0:
                chunk = self._create_chunk(
                    text=chunk_text,
                    page_id=page_id,
                    page_number=page_number,
                    chunk_index=chunk_index,
                    start_pos=start,
                    end_pos=split_point,
                )
                chunks.append(chunk)
                chunk_index += 1

            start = split_point - self.config.chunk_overlap
            if start <= 0:
                start = split_point

            if start >= end:
                start = end

        logger.debug(
            "text_chunked",
            page_id=page_id,
            page_number=page_number,
            chunk_count=len(chunks),
            text_length=text_length,
        )

        return chunks

    def _find_split_point(self, text: str, start: int, end: int) -> int:
        """Find the best point to split text."""
        if self.config.respect_sentences:
            search_start = max(start, end - 100)
            for i in range(end - 1, search_start, -1):
                if text[i] in self._sentence_endings and (
                    i + 1 >= len(text) or text[i + 1] in " \n"
                ):
                    return i + 1

        if self.config.respect_paragraphs:
            search_start = max(start, end - 100)
            for i in range(end - 1, search_start, -1):
                if text[i] == "\n" and i > 0 and text[i - 1] == "\n":
                    return i + 1
                if text[i] == "\n" and i + 2 < len(text):
                    next_char = text[i + 1]
                    if next_char.isupper() and i + 2 < len(text) and text[i + 2].isupper():
                        return i + 1

        for i in range(end - 1, start, -1):
            if text[i] in " \t\n":
                return i

        return end

    def _create_chunk(
        self,
        text: str,
        page_id: str,
        page_number: int,
        chunk_index: int,
        start_pos: int,
        end_pos: int,
    ) -> TextChunk:
        """Create a TextChunk with metadata."""
        return TextChunk(
            chunk_index=chunk_index,
            page_id=page_id,
            page_number=page_number,
            text=text,
        )

    def chunk_ocr_result(
        self,
        document_id: str,
        pages: list,
        page_ids: dict[int, str] | None = None,
    ) -> ChunkResult:
        """Chunk all pages from an OCR result."""
        page_ids = page_ids or {}
        all_chunks = []
        global_chunk_index = 0

        for page in pages:
            page_id = page_ids.get(page.page_number, f"{document_id}-page-{page.page_number}")
            page_chunks = self.chunk_text(
                text=page.text,
                page_id=page_id,
                page_number=page.page_number,
                document_id=document_id,
            )

            for chunk in page_chunks:
                chunk.chunk_index = global_chunk_index
                global_chunk_index += 1

            all_chunks.extend(page_chunks)

        logger.info(
            "document_chunked",
            document_id=document_id,
            page_count=len(pages),
            total_chunks=len(all_chunks),
        )

        return ChunkResult(document_id=document_id, chunks=all_chunks)

    def get_chunk_hash(self, chunk: TextChunk) -> str:
        """Generate a reproducible hash for a chunk."""
        content = f"{chunk.text}:{chunk.page_id}:{chunk.chunk_index}"
        return hashlib.sha256(content.encode()).hexdigest()[:16]


class ProcessingChunker(ChunkingService):
    """Enhanced chunker for the processing pipeline."""

    async def process_chunking_job(self, message: dict) -> ChunkResult:
        """Process a chunking job from the queue."""
        document_id = message["document_id"]
        pages = message["pages"]
        logger.info("starting_chunking", document_id=document_id)

        class PageObj:
            def __init__(self, page_number, text):
                self.page_number = page_number
                self.text = text

        page_objects = [PageObj(p["page_number"], p["text"]) for p in pages]

        result = self.chunk_ocr_result(document_id, page_objects)

        logger.info(
            "chunking_completed",
            document_id=document_id,
            total_chunks=len(result.chunks),
        )

        return result


chunking_service = ChunkingService()
processing_chunker = ProcessingChunker()
