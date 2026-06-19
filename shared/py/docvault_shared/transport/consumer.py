"""Generic queue consumer with retry and DLQ support."""

import json
import inspect
import signal
import threading
import time
from typing import Callable, Optional

from pika.adapters.blocking_connection import BlockingChannel
from pika.spec import Basic, BasicProperties

from .connection import RabbitMQConnection
from .publisher import QueuePublisher

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
        message.get("job_id")
        or message.get("message_id")
        or message.get("document_id")
        or message.get("version_id")
        or str(hash(json.dumps(message, sort_keys=True)))
    )


def _get_retry_count(message: dict) -> int:
    """Read a persisted retry counter from the message."""
    try:
        return max(int(message.get("retry_count", 0)), 0)
    except (TypeError, ValueError):
        return 0


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
        self.retry = f"{queue_name}.retry"
        self.max_retries = MAX_RETRY_ATTEMPTS
        self._should_stop = False

    def consume(self, callback: Callable[[dict], object]) -> None:
        """Start consuming messages from the queue."""
        channel = self._connection.channel
        channel.basic_qos(prefetch_count=1)

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
                error_str = f"invalid_json: {e}"
                logger.error(
                    "invalid_json_message",
                    queue=self.queue,
                    delivery_tag=method.delivery_tag,
                    error=error_str,
                    body=body[:100],
                )
                if self._publish_to_dlq(
                    {"raw_body": body.decode("utf-8", errors="replace")},
                    error_str,
                ):
                    ch.basic_ack(delivery_tag=method.delivery_tag)
                else:
                    ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)

            except Exception as e:
                error_str = str(e)
                logger.error(
                    "message_processing_failed",
                    queue=self.queue,
                    delivery_tag=method.delivery_tag,
                    error=error_str,
                )

                failed_message = message or {"raw_body": body.decode("utf-8", errors="replace")}
                msg_id = _extract_message_id(failed_message)
                retry_count = _get_retry_count(message or {})

                if _is_permanent_error(error_str) or retry_count >= self.max_retries:
                    logger.error(
                        "sending_message_to_dlq",
                        queue=self.queue,
                        message_id=msg_id,
                        error=error_str,
                        retry_count=retry_count,
                    )
                    if self._publish_to_dlq(failed_message, error_str, retry_count):
                        ch.basic_ack(delivery_tag=method.delivery_tag)
                    else:
                        ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)
                    return

                backoff = RETRY_BACKOFF_SECONDS * (2**retry_count)
                logger.warning(
                    "retrying_message_with_backoff",
                    queue=self.queue,
                    message_id=msg_id,
                    retry_count=retry_count,
                    backoff_seconds=backoff,
                )

                retry_message = dict(message or {})
                retry_message["retry_count"] = retry_count + 1

                if self._publish_with_delay(retry_message, backoff):
                    ch.basic_ack(delivery_tag=method.delivery_tag)
                else:
                    ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)

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

    def _publish(self, queue: str, message: dict) -> bool:
        """Publish a message and report whether it succeeded."""
        try:
            QueuePublisher(connection=self._connection).publish(queue, message)
            return True
        except Exception as exc:
            logger.error(
                "failed_to_publish_control_message",
                queue=queue,
                error=str(exc),
            )
            return False

    def _publish_with_delay(self, message: dict, delay_seconds: int) -> bool:
        """Publish a retry to the delayed-retry queue with a per-message TTL.

        The retry queue has no consumer and dead-letters back to the main queue
        when the TTL expires, so the message is redelivered only after the
        backoff elapses — without blocking the single consume thread.
        """
        try:
            QueuePublisher(connection=self._connection).publish(
                self.retry, message, expiration_ms=delay_seconds * 1000
            )
            return True
        except Exception as exc:
            logger.error(
                "failed_to_publish_retry_message",
                queue=self.retry,
                error=str(exc),
            )
            return False

    def _publish_to_dlq(self, message: dict, error: str, retry_count: int = 0) -> bool:
        """Publish a failed message to the dead-letter queue."""
        dlq_message = {
            "original_queue": self.queue,
            "original_message": message,
            "error": error,
            "retry_count": retry_count,
            "failed_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        }
        return self._publish(self.dlq, dlq_message)

    def close(self) -> None:
        """Stop the consumer."""
        self._should_stop = True
