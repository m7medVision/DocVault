"""Main entry point for the OCR Service."""

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
    """Main entry point for OCR Service (composition root)."""
    from docvault_shared.config import config
    from docvault_shared import telemetry
    from docvault_shared.transport.connection import RabbitMQConnection
    from docvault_shared.transport.consumer import QueueConsumer
    from docvault_shared.transport.publisher import QueuePublisher
    from docvault_shared.database import get_ocr_persistence

    from .application.ocr_job import OCRJobHandler
    from .ocr import MistralOCRClient
    from .storage import MinIOClient

    telemetry.init_telemetry("docvault-ocr")

    logger.info(
        "starting_ocr_service",
        environment=config.environment,
    )

    connections: list[RabbitMQConnection] = []

    try:
        conn = RabbitMQConnection(
            queues=[config.rabbitmq_queue_ocr, config.rabbitmq_queue_processing]
        )
        connections.append(conn)

        storage = MinIOClient()
        ocr_client = MistralOCRClient(storage)
        publisher = QueuePublisher(connection=conn)
        handler = OCRJobHandler(
            ocr_client,
            get_ocr_persistence(),
            publisher,
            config.rabbitmq_queue_processing,
        )

        consumer = QueueConsumer(
            conn, config.rabbitmq_queue_ocr, f"{config.rabbitmq_queue_ocr}.dlq"
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
