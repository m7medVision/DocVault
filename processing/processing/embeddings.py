"""Embedding generation service using OpenRouter."""

import structlog

import openai

from docvault_shared.config import (
    APPROVED_OPENROUTER_EMBEDDING_MODELS,
    OPENROUTER_EMBEDDING_DIMENSIONS,
    config,
    validate_openrouter_model_name,
)

logger = structlog.get_logger(__name__)


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
        results = []
        for text in texts:
            embedding = self.provider.generate_embedding(text)
            results.append(embedding)
            logger.debug(
                "batch_embedding_generated",
                text_length=len(text),
            )

        logger.info(
            "batch_embeddings_completed",
            count=len(texts),
        )

        return results


embedding_service = EmbeddingService()


async def generate_chunk_embeddings(chunks: list) -> list[list[float]]:
    texts = [chunk.text for chunk in chunks]
    return embedding_service.generate_embeddings_batch(texts)
