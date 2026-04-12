"""Data models for DocVault services."""

from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Optional


class DocumentStatus(str, Enum):
    """Document processing status."""

    PENDING = "pending"
    PROCESSING = "processing"
    PROCESSED = "processed"
    FAILED = "failed"


class ProcessingJobType(str, Enum):
    """Type of processing job."""

    OCR = "ocr"
    CHUNKING = "chunking"
    EMBEDDING = "embedding"
    METADATA_EXTRACTION = "metadata_extraction"


@dataclass
class OCRJob:
    """OCR job from RabbitMQ queue."""

    document_id: str
    version_id: str
    storage_key: str
    mime_type: str
    tenant_id: str
    org_id: str
    language: Optional[str] = None


@dataclass
class OCRResult:
    """Result from Mistral OCR API."""

    document_id: str
    version_id: str
    pages: list["OCRPage"]
    completed_at: datetime


@dataclass
class OCRPage:
    """OCR result for a single page."""

    page_number: int
    text: str
    confidence: float
    model: str = "mistral-ocr-2503"


@dataclass
class ChunkResult:
    """Result from text chunking."""

    document_id: str
    chunks: list["TextChunk"]


@dataclass
class TextChunk:
    """A chunk of text for embedding."""

    chunk_index: int
    page_id: str
    page_number: int
    text: str


@dataclass
class EmbeddingResult:
    """Result from embedding generation."""

    document_id: str
    chunks: list["EmbeddedChunk"]


@dataclass
class EmbeddedChunk:
    """A text chunk with its embedding vector."""

    chunk_index: int
    page_id: str
    embedding: list[float]
    token_count: int


@dataclass
class ProcessingError:
    """Error during document processing."""

    document_id: str
    stage: ProcessingJobType
    error_message: str
    retry_count: int = 0
    timestamp: datetime = None
