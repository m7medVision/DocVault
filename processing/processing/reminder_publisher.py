"""Reminder job publisher for the worker layer."""

import structlog
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Optional

from docvault_shared.config import config
from docvault_shared.transport.connection import RabbitMQConnection
from docvault_shared.transport.publisher import QueuePublisher

logger = structlog.get_logger(__name__)


@dataclass
class ReminderJob:
    """Reminder job message for the worker layer."""

    document_id: str
    tenant_id: str
    org_id: str
    source_text: str
    document_type: Optional[str] = None
    expiry_date: Optional[str] = None
    issuer: Optional[str] = None
    priority: str = "medium"


_connection: Optional[RabbitMQConnection] = None


def _get_connection() -> RabbitMQConnection:
    global _connection
    if _connection is None:
        _connection = RabbitMQConnection()
    return _connection


class ReminderPublisher:
    """Publishes reminder jobs to RabbitMQ for the worker layer."""

    def __init__(self):
        """Initialize publisher."""
        self.queue = config.rabbitmq_queue_reminder

    def publish_reminder_job(
        self,
        document_id: str,
        tenant_id: str,
        org_id: str,
        source_text: str,
        document_type: Optional[str] = None,
        expiry_date: Optional[str] = None,
        issuer: Optional[str] = None,
        priority: str = "medium",
    ) -> str:
        """Publish a reminder job to the queue."""
        job_id = f"reminder-{document_id}-{datetime.utcnow().timestamp()}"

        job = ReminderJob(
            document_id=document_id,
            tenant_id=tenant_id,
            org_id=org_id,
            source_text=source_text,
            document_type=document_type,
            expiry_date=expiry_date,
            issuer=issuer,
            priority=priority,
        )

        message = {
            "job_id": job_id,
            "document_id": document_id,
            "tenant_id": tenant_id,
            "org_id": org_id,
            "source_text": source_text,
            "document_type": document_type,
            "expiry_date": expiry_date,
            "issuer": issuer,
            "priority": priority,
            "created_at": datetime.now(timezone.utc).isoformat(),
        }

        try:
            conn = _get_connection()
            publisher = QueuePublisher(connection=conn)
            publisher.publish(queue=self.queue, message=message)
            logger.info(
                "reminder_job_published",
                job_id=job_id,
                document_id=document_id,
                queue=self.queue,
            )
            return job_id

        except Exception as e:
            logger.error(
                "failed_to_publish_reminder_job",
                document_id=document_id,
                error=str(e),
            )
            raise

    async def publish_after_processing(
        self,
        document_id: str,
        tenant_id: str,
        org_id: str,
        pages: list,
        metadata: dict,
    ) -> Optional[str]:
        """Publish reminder job after document processing completes."""
        source_text = "\n\n".join(f"[Page {p.page_number}]\n{p.text}" for p in pages)

        doc_type = metadata.get("doc_type")
        priority = self._determine_priority(doc_type)

        expiry_date = metadata.get("expiry_date")

        if doc_type in ("identity",):
            logger.info(
                "reminder_skipped_for_doc_type",
                document_id=document_id,
                doc_type=doc_type,
            )
            return None

        job_id = self.publish_reminder_job(
            document_id=document_id,
            tenant_id=tenant_id,
            org_id=org_id,
            source_text=source_text,
            document_type=doc_type,
            expiry_date=expiry_date,
            issuer=metadata.get("issuer"),
            priority=priority,
        )

        return job_id

    def _determine_priority(self, doc_type: Optional[str]) -> str:
        """Determine reminder job priority based on document type."""
        high_priority = {"contract", "warranty", "identity"}
        medium_priority = {"invoice", "receipt"}

        if doc_type in high_priority:
            return "high"
        elif doc_type in medium_priority:
            return "medium"
        else:
            return "low"


reminder_publisher = ReminderPublisher()
