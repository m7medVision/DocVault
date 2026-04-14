"""Main entry point for the Processing Pipeline Service."""

import sys
import threading
from typing import Callable

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


class ConsumerThread(threading.Thread):
    """Run a blocking consumer in a background thread."""

    def __init__(self, name: str, starter: Callable[[], None]) -> None:
        super().__init__(name=f"{name}-consumer", daemon=True)
        self._starter = starter
        self.error: Exception | None = None

    def run(self) -> None:
        try:
            self._starter()
        except Exception as exc:
            self.error = exc


def main() -> None:
    """Main entry point for Processing Pipeline Service."""
    from docvault_shared.config import config
    from docvault_shared import telemetry
    from docvault_shared.transport.connection import RabbitMQConnection
    from docvault_shared.transport.consumer import QueueConsumer
    from .application.processing_job import ProcessingJobHandler
    from .chunker import processing_chunker
    from .embeddings import generate_chunk_embeddings
    from .classifier import metadata_extractor
    from .database import chunk_persistence
    from .reminder_publisher import reminder_publisher

    telemetry.init_telemetry("docvault-processing")

    logger.info(
        "starting_processing_pipeline_service",
        environment=config.environment,
    )

    connections: list[RabbitMQConnection] = []

    try:
        conn = RabbitMQConnection(
            queues=[config.rabbitmq_queue_processing, config.rabbitmq_queue_reminder]
        )
        connections.append(conn)
        reminder_publisher.set_connection(conn)
        handler = ProcessingJobHandler(
            processing_chunker,
            generate_chunk_embeddings,
            metadata_extractor,
            chunk_persistence,
            chunk_persistence,
            reminder_publisher,
            conn,
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
