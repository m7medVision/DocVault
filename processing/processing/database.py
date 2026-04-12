"""Database access for the Processing Service with pgvector support."""

import structlog
import uuid
from datetime import datetime
from typing import Optional

from sqlalchemy import Column, DateTime, Float, ForeignKey, Integer, String, Text, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import Session, sessionmaker
from sqlalchemy.types import UserDefinedType

from docvault_shared.config import config
from docvault_shared.models import TextChunk

logger = structlog.get_logger(__name__)

Base = declarative_base()


class Vector(UserDefinedType):
    """Custom type for pgvector vectors."""

    cache_ok = True

    def __init__(self, dimensions: int = 1024):
        super().__init__()
        self.dimensions = dimensions

    def get_col_spec(self):
        return f"vector({self.dimensions})"

    def bind_processor(self, dialect):
        def process(value):
            if value is not None:
                return list(value)
            return value

        return process

    def result_processor(self, dialect, coltype):
        def process(value):
            if value is not None:
                return list(value)
            return value

        return process


class ExtractedTextChunk(Base):
    """Model for extracted_text_chunks table with pgvector."""

    __tablename__ = "extracted_text_chunks"

    id = Column(String, primary_key=True)
    document_id = Column(String, ForeignKey("documents.id"), nullable=False)
    page_id = Column(String, ForeignKey("document_pages.id"), nullable=False)
    chunk_index = Column(Integer, nullable=False)
    chunk_text = Column(Text, nullable=False)
    embedding = Column(Vector(1024))
    created_at = Column(DateTime)


class Document(Base):
    """Model for documents table."""

    __tablename__ = "documents"

    id = Column(String, primary_key=True)
    tenant_id = Column(String, nullable=False)
    org_id = Column(String, nullable=False)
    folder_id = Column(String)
    owner_id = Column(String, nullable=False)
    title = Column(String, nullable=False)
    doc_type = Column(String)
    status = Column(String, nullable=False)
    language = Column(String)
    created_at = Column(DateTime)


class DocumentPage(Base):
    """Model for document_pages table."""

    __tablename__ = "document_pages"

    id = Column(String, primary_key=True)
    document_id = Column(String, ForeignKey("documents.id"), nullable=False)
    version_id = Column(String, nullable=False)
    page_number = Column(Integer, nullable=False)
    ocr_text = Column(Text)
    translated_text = Column(Text)
    confidence = Column(Float)
    ocr_model = Column(String)
    created_at = Column(DateTime)


def _chunk_id(document_id: str, page_id: str, chunk_index: int) -> str:
    """Build a deterministic UUID for a chunk row."""
    return str(uuid.uuid5(uuid.NAMESPACE_URL, f"{document_id}:{page_id}:{chunk_index}"))


class PGVectorRepository:
    """pgvector repository for document chunks."""

    def __init__(self, db_url: Optional[str] = None):
        self.db_url = db_url or config.database_url
        self.engine = create_engine(
            self.db_url,
            pool_pre_ping=True,
            pool_size=5,
            max_overflow=10,
        )
        self.SessionLocal = sessionmaker(bind=self.engine)

    def create_tables(self) -> None:
        """Create all tables in the database."""
        Base.metadata.create_all(self.engine)
        logger.info("pgvector_tables_created")

    def get_session(self) -> Session:
        """Get a new database session."""
        return self.SessionLocal()

    async def upsert_chunk(
        self,
        document_id: str,
        chunk: TextChunk,
        embedding: list[float],
    ) -> str:
        """Insert or update a text chunk with embedding."""
        session = self.get_session()
        chunk_id_val = _chunk_id(document_id, chunk.page_id, chunk.chunk_index)

        try:
            text_chunk = ExtractedTextChunk(
                id=chunk_id_val,
                document_id=document_id,
                page_id=chunk.page_id,
                chunk_index=chunk.chunk_index,
                chunk_text=chunk.text,
                embedding=embedding,
                created_at=datetime.utcnow(),
            )
            session.merge(text_chunk)
            session.commit()
            logger.debug("chunk_upserted", chunk_id=chunk_id_val)
            return chunk_id_val

        except Exception as e:
            session.rollback()
            logger.error("chunk_upsert_failed", chunk_id=chunk_id_val, error=str(e))
            raise

        finally:
            session.close()

    async def save_chunks(
        self,
        document_id: str,
        chunks: list[TextChunk],
        embeddings: list[list[float]],
    ) -> list[str]:
        """Save text chunks with embeddings to pgvector."""
        session = self.get_session()
        chunk_ids = []

        try:
            for chunk, embedding in zip(chunks, embeddings):
                chunk_id_val = _chunk_id(document_id, chunk.page_id, chunk.chunk_index)

                text_chunk = ExtractedTextChunk(
                    id=chunk_id_val,
                    document_id=document_id,
                    page_id=chunk.page_id,
                    chunk_index=chunk.chunk_index,
                    chunk_text=chunk.text,
                    embedding=embedding,
                    created_at=datetime.utcnow(),
                )
                session.merge(text_chunk)
                chunk_ids.append(chunk_id_val)

            session.commit()
            logger.info("chunks_saved", document_id=document_id, count=len(chunk_ids))
            return chunk_ids

        except Exception as e:
            session.rollback()
            logger.error("save_chunks_failed", document_id=document_id, error=str(e))
            raise

        finally:
            session.close()

    async def delete_by_document(self, doc_id: str) -> None:
        """Delete all chunks for a document."""
        session = self.get_session()

        try:
            session.query(ExtractedTextChunk).filter(
                ExtractedTextChunk.document_id == doc_id
            ).delete()
            session.commit()
            logger.info("chunks_deleted", document_id=doc_id)

        except Exception as e:
            session.rollback()
            logger.error("delete_chunks_failed", document_id=doc_id, error=str(e))
            raise

        finally:
            session.close()

    async def update_page_translations(
        self,
        document_id: str,
        translations: dict[int, str],
    ) -> None:
        """Update translated_text for document pages."""
        session = self.get_session()

        try:
            for page_number, translated_text in translations.items():
                page = (
                    session.query(DocumentPage)
                    .filter(
                        DocumentPage.document_id == document_id,
                        DocumentPage.page_number == page_number,
                    )
                    .first()
                )
                if page:
                    page.translated_text = translated_text

            session.commit()
            logger.info(
                "page_translations_updated",
                document_id=document_id,
                count=len(translations),
            )

        except Exception as e:
            session.rollback()
            logger.error(
                "update_page_translations_failed",
                document_id=document_id,
                error=str(e),
            )
            raise

        finally:
            session.close()

    async def get_document_status(self, document_id: str) -> Optional[str]:
        """Get document status."""
        session = self.get_session()

        try:
            doc = session.query(Document).filter(Document.id == document_id).first()
            return doc.status if doc else None

        finally:
            session.close()

    async def update_document_status(self, document_id: str, status: str) -> None:
        """Update document status."""
        session = self.get_session()

        try:
            doc = session.query(Document).filter(Document.id == document_id).first()
            if doc:
                doc.status = status
                session.commit()
                logger.info("document_status_updated", document_id=document_id, status=status)
            else:
                logger.warning("document_not_found", document_id=document_id)

        except Exception as e:
            session.rollback()
            logger.error("update_status_failed", document_id=document_id, error=str(e))
            raise

        finally:
            session.close()

    async def update_document_metadata(
        self,
        document_id: str,
        doc_type: Optional[str] = None,
        language: Optional[str] = None,
    ) -> None:
        """Update classified document metadata in database."""
        session = self.get_session()

        try:
            doc = session.query(Document).filter(Document.id == document_id).first()
            if not doc:
                logger.warning(
                    "document_not_found_for_metadata_update",
                    document_id=document_id,
                )
                return

            updated = False
            if doc_type:
                setattr(doc, "doc_type", doc_type)
                updated = True
            if language:
                setattr(doc, "language", language)
                updated = True

            if updated:
                session.commit()
                logger.info(
                    "document_metadata_updated",
                    document_id=document_id,
                    doc_type=doc_type,
                    language=language,
                )

        except Exception as e:
            session.rollback()
            logger.error(
                "failed_to_update_document_metadata",
                document_id=document_id,
                error=str(e),
            )
            raise

        finally:
            session.close()


db = PGVectorRepository()
chunk_persistence = db
