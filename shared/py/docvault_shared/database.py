"""Database access for DocVault services."""

import structlog
import uuid
from datetime import datetime
from typing import Optional

from sqlalchemy import Column, DateTime, Float, ForeignKey, Integer, String, Text, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import Session, sessionmaker

from .config import config

logger = structlog.get_logger(__name__)

Base = declarative_base()


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
    processing_stage = Column(String)
    processing_error = Column(Text)
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


class OCRPersistence:
    """Handles persistence of OCR results to database."""

    def __init__(self, db_url: Optional[str] = None):
        """Initialize persistence layer."""
        self.db_url = db_url or config.database_url
        self.engine = create_engine(
            self.db_url,
            pool_pre_ping=True,
            pool_size=5,
            max_overflow=10,
        )
        self.SessionLocal = sessionmaker(bind=self.engine)

    def get_session(self) -> Session:
        """Get a new database session."""
        return self.SessionLocal()

    async def save_ocr_results(self, result) -> list[str]:
        """Save OCR results to database."""
        session = self.get_session()
        page_ids = []

        try:
            for page in result.pages:
                page_id = str(
                    uuid.uuid5(
                        uuid.NAMESPACE_URL,
                        f"{result.document_id}:{result.version_id}:{page.page_number}",
                    )
                )

                doc_page = DocumentPage(
                    id=page_id,
                    document_id=result.document_id,
                    version_id=result.version_id,
                    page_number=page.page_number,
                    ocr_text=page.text,
                    confidence=page.confidence,
                    ocr_model=page.model,
                    created_at=datetime.utcnow(),
                )

                session.merge(doc_page)
                page_ids.append(page_id)

            session.commit()
            logger.info(
                "ocr_results_saved",
                document_id=result.document_id,
                page_count=len(page_ids),
            )

            return page_ids

        except Exception as e:
            session.rollback()
            logger.error(
                "failed_to_save_ocr_results",
                document_id=result.document_id,
                error=str(e),
            )
            raise

        finally:
            session.close()

    async def update_document_status(
        self,
        document_id: str,
        status: str,
    ) -> None:
        """Update document status in database."""
        session = self.get_session()

        try:
            doc = session.query(Document).filter(Document.id == document_id).first()
            if doc:
                setattr(doc, "status", status)
                session.commit()
                logger.info(
                    "document_status_updated",
                    document_id=document_id,
                    status=status,
                )
            else:
                logger.warning(
                    "document_not_found_for_status_update",
                    document_id=document_id,
                )

        except Exception as e:
            session.rollback()
            logger.error(
                "failed_to_update_document_status",
                document_id=document_id,
                error=str(e),
            )
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


def get_ocr_persistence() -> OCRPersistence:
    """Get an OCR persistence instance."""
    return OCRPersistence()
