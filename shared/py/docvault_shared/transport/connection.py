"""RabbitMQ connection management."""

from typing import Optional

import pika
from pika.adapters.blocking_connection import BlockingChannel
from pika.exceptions import ChannelClosedByBroker

from .config import config

logger = pika.adapters.blocking_connection.BlockingConnection.__module__


class RabbitMQConnection:
    """Manages RabbitMQ connection and channel lifecycle."""

    def __init__(self, url: Optional[str] = None):
        self.url = url or config.rabbitmq_url
        self._connection: Optional[pika.BlockingConnection] = None
        self._channel: Optional[BlockingChannel] = None

    def connect(self) -> BlockingChannel:
        """Establish connection to RabbitMQ and return channel."""
        try:
            params = pika.URLParameters(self.url)
            self._connection = pika.BlockingConnection(params)
            self._channel = self._connection.channel()
            self._setup_queues()
            return self._channel
        except Exception as e:
            raise ConnectionError(f"Failed to connect to RabbitMQ: {e}") from e

    def _setup_queues(self) -> None:
        """Declare required queues with DLQ support."""
        queues = [
            config.rabbitmq_queue_ocr,
            config.rabbitmq_queue_processing,
            config.rabbitmq_queue_reminder,
        ]
        for queue in queues:
            self._declare_queue(queue)

    def _declare_queue(self, queue: str) -> None:
        """Declare a queue, creating it with DLQ args only when absent."""
        channel = self._require_channel()

        try:
            channel.queue_declare(queue=queue, passive=True)
        except ChannelClosedByBroker as exc:
            if exc.reply_code != 404:
                raise

            connection = self._require_connection()
            self._channel = connection.channel()
            channel = self._require_channel()
            channel.queue_declare(
                queue=queue,
                durable=True,
                arguments={
                    "x-dead-letter-exchange": "",
                    "x-dead-letter-routing-key": f"{queue}.dlq",
                },
            )

        channel.queue_declare(queue=f"{queue}.dlq", durable=True)

    def _require_connection(self) -> pika.BlockingConnection:
        if self._connection is None:
            raise RuntimeError("RabbitMQ connection not initialized")
        return self._connection

    def _require_channel(self) -> BlockingChannel:
        if self._channel is None:
            raise RuntimeError("RabbitMQ channel not initialized")
        return self._channel

    @property
    def channel(self) -> BlockingChannel:
        """Return the channel, connecting if necessary."""
        if self._channel is None:
            return self.connect()
        return self._channel

    def close(self) -> None:
        """Close the RabbitMQ connection."""
        if self._connection and not self._connection.is_closed:
            self._connection.close()
