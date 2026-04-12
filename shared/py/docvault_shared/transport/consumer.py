"""Generic queue consumer with retry and DLQ support."""

import json
import inspect
import signal
import threading
import time
from typing import Callable, Optional

import pika
from pika.adapters.blocking_connection import BlockingChannel
from pika.spec import Basic, BasicProperties

from .connection import RabbitMQConnection

logger = __import__("structlog").get_logger(__name__)

MAX_RETRY_ATTEMPTS = 3
RETRY_BACKOFF_SECONDS = 30

_PERMANENT_ERROR_CODES = {400, 401, 403, 404, 422}


def _is_permanent_error(error_str: str) -> bool:
    """Check if an error string indicates a permanent (non-retryable) failure."""
    error_lower = error_str.lower()
    for code in _PERMANENT_ERROR_CODES:
        if f"status {code}" in error_lower or f"status_code={code}" in error_lower:
            return True
    if "invalid file format" in error_lower:
        return True
    if "invalid_json" in error_lower or "malformed" in error_lower:
        return True
    return False


def _extract_message_id(message: dict) -> str:
    """Extract a stable identifier from a message for retry tracking."""
    return (
        message.get("document_id")
        or message.get("version_id")
        or str(hash(json.dumps(message, sort_keys=True)))
    )


class QueueConsumer:
    """Generic queue consumer with retry and DLQ built-in."""

    def __init__(
        self,
        connection: RabbitMQConnection,
        queue_name: str,
        dlq_name: Optional[str] = None,
    ):
        self._connection = connection
        self.queue = queue_name
        self.dlq = dlq_name or f"{queue_name}.dlq"
        self.max_retries = MAX_RETRY_ATTEMPTS
        self._should_stop = False

    def consume(self, callback: Callable[[dict], object]) -> None:
        """Start consuming messages from the queue."""
        channel = self._connection.channel
        channel.basic_qos(prefetch_count=1)

        retry_counts: dict[str, int] = {}

        def on_message(
            ch: BlockingChannel,
            method: Basic.Deliver,
            properties: BasicProperties,
            body: bytes,
        ) -> None:
            message: Optional[dict] = None
            try:
                message = json.loads(body.decode("utf-8"))
                logger.info(
                    "message_received",
                    queue=self.queue,
                    delivery_tag=method.delivery_tag,
                )

                if inspect.iscoroutinefunction(callback):
                    import asyncio

                    asyncio.run(callback(message))
                else:
                    callback(message)

                ch.basic_ack(delivery_tag=method.delivery_tag)
                logger.info(
                    "message_acked",
                    queue=self.queue,
                    delivery_tag=method.delivery_tag,
                )

            except json.JSONDecodeError as e:
                logger.error(
                    "invalid_json_message",
                    queue=self.queue,
                    delivery_tag=method.delivery_tag,
                    error=str(e),
                    body=body[:100],
                )
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)

            except Exception as e:
                error_str = str(e)
                logger.error(
                    "message_processing_failed",
                    queue=self.queue,
                    delivery_tag=method.delivery_tag,
                    error=error_str,
                )

                dlq_message = message or json.loads(body.decode("utf-8"))
                msg_id = _extract_message_id(dlq_message)

                if _is_permanent_error(error_str):
                    logger.error(
                        "permanent_error_sending_to_dlq",
                        queue=self.queue,
                        message_id=msg_id,
                        error=error_str,
                    )
                    self._publish_to_dlq(dlq_message, error_str)
                    ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
                    return

                retry_count = retry_counts.get(msg_id, 0)
                if retry_count >= self.max_retries:
                    logger.error(
                        "max_retry_attempts_exceeded",
                        queue=self.queue,
                        message_id=msg_id,
                        retry_count=retry_count,
                    )
                    self._publish_to_dlq(dlq_message, error_str)
                    ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
                else:
                    backoff = RETRY_BACKOFF_SECONDS * (2**retry_count)
                    logger.warning(
                        "retrying_message_with_backoff",
                        queue=self.queue,
                        message_id=msg_id,
                        retry_count=retry_count,
                        backoff_seconds=backoff,
                    )
                    retry_counts[msg_id] = retry_count + 1
                    ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)
                    time.sleep(backoff)

        channel.basic_consume(queue=self.queue, on_message_callback=on_message)
        logger.info("started_consuming", queue=self.queue)

        if threading.current_thread() is threading.main_thread():
            signal.signal(signal.SIGINT, self._signal_handler)
            signal.signal(signal.SIGTERM, self._signal_handler)

        try:
            while not self._should_stop:
                self._connection._connection.process_data_events(time_limit=1)
        except Exception as e:
            logger.error("consumer_error", error=str(e))
            raise

    def _signal_handler(self, signum, frame) -> None:
        """Handle shutdown signals."""
        logger.info("shutdown_signal_received", signal=signum)
        self._should_stop = True

    def _publish_to_dlq(self, message: dict, error: str) -> None:
        """Publish a failed message to the dead-letter queue."""
        dlq_message = {
            "original_queue": self.queue,
            "original_message": message,
            "error": error,
        }
        publisher = QueuePublisher(self._connection.channel, self.queue)
        publisher.publish(self.dlq, dlq_message)

    def close(self) -> None:
        """Stop the consumer."""
        self._should_stop = True
