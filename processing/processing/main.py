"""Main entry point for the Processing Pipeline Service."""

import sys

import structlog
import structlog.stdlib

structlog.configure(
    processors=[
        structlog.stdlib.filter_by_level,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.add_log_level,
        structlog.stdlib.PositionalArgumentsFormatter(),
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.UnicodeDecoder(),
        structlog.processors.JSONRenderer(),
    ],
    wrapper_class=structlog.stdlib.BoundLogger,
    context_class=dict,
    logger_factory=structlog.stdlib.LoggerFactory(),
    cache_logger_on_first_use=True,
)

logger = structlog.get_logger(__name__)


def main() -> None:
    """Main entry point for Processing Pipeline Service.

    Composition root: construct the external clients (LLM/embeddings, the
    pgvector repository, the reminder publisher) once here and inject them into
    the handler, rather than relying on import-time module singletons.
    """
    from docvault_shared.config import config
    from docvault_shared import telemetry
    from docvault_shared.transport.connection import RabbitMQConnection
    from docvault_shared.transport.consumer import QueueConsumer
    from .application.processing_job import ProcessingJobHandler
    from .chunker import ProcessingChunker
    from .classifier import MetadataExtractor
    from .database import PGVectorRepository
    from .embeddings import EmbeddingService
    from .reminder_publisher import ReminderPublisher
    from .translation import DocumentTranslator

    telemetry.init_telemetry("docvault-processing")

    logger.info(
        "starting_processing_pipeline_service",
        environment=config.environment,
    )

    connections: list[RabbitMQConnection] = []

    # Build external clients once (the composition root) and inject them.
    chunker = ProcessingChunker()
    embedding_service = EmbeddingService()
    metadata_extractor = MetadataExtractor()
    repository = PGVectorRepository()
    reminder_publisher = ReminderPublisher()
    translator = DocumentTranslator()

    try:
        conn = RabbitMQConnection(
            queues=[config.rabbitmq_queue_processing, config.rabbitmq_queue_reminder]
        )
        connections.append(conn)
        reminder_publisher.set_connection(conn)
        handler = ProcessingJobHandler(
            chunker=chunker,
            embedder=embedding_service.generate_chunk_embeddings,
            classifier=metadata_extractor,
            repo=repository,
            reminder_publisher=reminder_publisher,
            connection=conn,
            translator=translator,
        )
        consumer = QueueConsumer(
            conn, config.rabbitmq_queue_processing, f"{config.rabbitmq_queue_processing}.dlq"
        )
        consumer.consume(handler.handle)

    except KeyboardInterrupt:
        logger.info("shutting_down")
        sys.exit(0)
    except Exception as e:
        logger.error("fatal_error", error=str(e))
        sys.exit(1)
    finally:
        for conn in connections:
            try:
                conn.close()
            except Exception as exc:
                logger.warning("failed_to_close_connection", error=str(exc))
        telemetry.shutdown_telemetry()


if __name__ == "__main__":
    main()
