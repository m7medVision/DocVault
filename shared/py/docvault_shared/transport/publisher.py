"""RabbitMQ publisher with confirmation."""

import json
from typing import Optional

import pika
from pika.adapters.blocking_connection import BlockingChannel

from .connection import RabbitMQConnection

logger = __import__("structlog").get_logger(__name__)


class QueuePublisher:
    """RabbitMQ publisher with confirmation."""

    def __init__(
        self,
        channel: Optional[BlockingChannel] = None,
        connection: Optional[RabbitMQConnection] = None,
    ):
        if channel is not None:
            self._channel = channel
        elif connection is not None:
            self._channel = connection.channel
        else:
            raise ValueError("Either channel or connection must be provided")

    def publish(
        self, queue: str, message: dict, expiration_ms: Optional[int] = None
    ) -> None:
        """Publish a message to a queue.

        When expiration_ms is set it becomes the message's per-message TTL,
        used to drive delayed retries via a dead-lettered retry queue.
        """
        try:
            self._channel.basic_publish(
                exchange="",
                routing_key=queue,
                body=json.dumps(message),
                properties=pika.BasicProperties(
                    delivery_mode=2,
                    content_type="application/json",
                    expiration=(
                        str(expiration_ms) if expiration_ms is not None else None
                    ),
                ),
            )
            logger.info("message_published", queue=queue)
        except Exception as e:
            logger.error("failed_to_publish_message", queue=queue, error=str(e))
            raise
