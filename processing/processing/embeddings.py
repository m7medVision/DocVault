"""Embedding generation service using OpenRouter."""

import asyncio

import structlog

import openai

from docvault_shared.config import (
    APPROVED_OPENROUTER_EMBEDDING_MODELS,
    OPENROUTER_EMBEDDING_DIMENSIONS,
    config,
    validate_openrouter_model_name,
)

logger = structlog.get_logger(__name__)

# Max number of inputs to send in a single embeddings request. The OpenAI /
# OpenRouter embeddings API accepts a list input; we cap each HTTP request at
# this many texts so a very large document still issues only a handful of calls
# instead of one-call-per-chunk. Vectors are returned in input order.
EMBEDDING_REQUEST_BATCH_SIZE = 128


class OpenRouterEmbeddingProvider:
    """Embedding provider backed by OpenRouter's OpenAI-compatible API."""

    def __init__(self, api_key: str | None = None, model: str | None = None):
        self.model_name = validate_openrouter_model_name(
            model or config.embedding_model,
            APPROVED_OPENROUTER_EMBEDDING_MODELS,
        )
        self.api_key = api_key or config.openrouter_api_key
        self.client = openai.OpenAI(
            api_key=self.api_key,
            base_url="https://openrouter.ai/api/v1",
        )

    def generate_embedding(self, text: str) -> list[float]:
        response = self.client.embeddings.create(
            model=self.model_name,
            input=text,
            dimensions=OPENROUTER_EMBEDDING_DIMENSIONS,
        )
        embedding = response.data[0].embedding
        logger.debug(
            "openrouter_embedding_generated",
            text_length=len(text),
            embedding_dim=len(embedding),
        )
        if len(embedding) != OPENROUTER_EMBEDDING_DIMENSIONS:
            raise ValueError(
                f"Unexpected embedding size returned from OpenRouter: {len(embedding)}"
            )
        return embedding

    def generate_embeddings(self, texts: list[str]) -> list[list[float]]:
        """Embed many texts in a single request and return vectors in order.

        The OpenAI/OpenRouter embeddings API accepts a list ``input`` and
        returns one entry per input. ``response.data`` may arrive out of order,
        so we re-sort by ``index`` to guarantee the output order matches the
        input order.
        """
        if not texts:
            return []

        response = self.client.embeddings.create(
            model=self.model_name,
            input=texts,
            dimensions=OPENROUTER_EMBEDDING_DIMENSIONS,
        )
        ordered = sorted(response.data, key=lambda item: item.index)
        embeddings = [item.embedding for item in ordered]

        if len(embeddings) != len(texts):
            raise ValueError(
                "Embedding count mismatch from OpenRouter: "
                f"expected {len(texts)}, got {len(embeddings)}"
            )
        for embedding in embeddings:
            if len(embedding) != OPENROUTER_EMBEDDING_DIMENSIONS:
                raise ValueError(
                    f"Unexpected embedding size returned from OpenRouter: {len(embedding)}"
                )
        logger.debug(
            "openrouter_embeddings_generated",
            count=len(embeddings),
        )
        return embeddings

    def count_tokens(self, text: str) -> int:
        return len(text) // 4

    def get_name(self) -> str:
        return "openrouter"


class EmbeddingService:
    """Service for generating embeddings with OpenRouter."""

    def __init__(self):
        self.provider = OpenRouterEmbeddingProvider()
        logger.info(
            "embedding_service_initialized",
            model=config.embedding_model,
        )

    def generate_embedding(self, text: str) -> list[float]:
        return self.provider.generate_embedding(text)

    def generate_embeddings_batch(self, texts: list[str]) -> list[list[float]]:
        """Embed every text, sending the list to the API in batched requests.

        Texts are sent in a single request when they fit within
        ``EMBEDDING_REQUEST_BATCH_SIZE``; larger inputs are split into
        successive requests. The returned vectors preserve input order.
        """
        results: list[list[float]] = []
        for start in range(0, len(texts), EMBEDDING_REQUEST_BATCH_SIZE):
            sub_batch = texts[start : start + EMBEDDING_REQUEST_BATCH_SIZE]
            results.extend(self.provider.generate_embeddings(sub_batch))

        logger.info(
            "batch_embeddings_completed",
            count=len(texts),
        )

        return results

    async def generate_chunk_embeddings(self, chunks: list) -> list[list[float]]:
        """Embed a list of chunks (the EmbedderPort entry point).

        Offloaded to a worker thread so the blocking HTTP call does not stall
        the event loop.
        """
        texts = [chunk.text for chunk in chunks]
        return await asyncio.to_thread(self.generate_embeddings_batch, texts)
