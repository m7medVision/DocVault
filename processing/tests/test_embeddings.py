"""Tests for the OpenRouter embedding service batching behavior."""

from types import SimpleNamespace

import pytest

from docvault_shared.config import OPENROUTER_EMBEDDING_DIMENSIONS

from processing import embeddings


class FakeEmbeddings:
    """Records every embeddings.create() call and returns ordered vectors."""

    def __init__(self) -> None:
        self.calls: list[dict] = []

    def create(self, **kwargs):
        self.calls.append(kwargs)
        inputs = kwargs["input"]
        # The OpenAI/OpenRouter API may return data out of order; emulate that
        # by reversing, each carrying its original index so the provider can
        # re-sort.
        data = [
            SimpleNamespace(
                index=i,
                embedding=[float(i)] * OPENROUTER_EMBEDDING_DIMENSIONS,
            )
            for i in range(len(inputs))
        ]
        return SimpleNamespace(data=list(reversed(data)))


class FakeOpenAIClient:
    def __init__(self, embeddings_obj: FakeEmbeddings):
        self.embeddings = embeddings_obj


def _build_service(monkeypatch, fake_embeddings: FakeEmbeddings) -> embeddings.EmbeddingService:
    monkeypatch.setattr(
        embeddings.openai,
        "OpenAI",
        lambda **_: FakeOpenAIClient(fake_embeddings),
    )
    return embeddings.EmbeddingService()


def test_generate_embeddings_batch_sends_single_request(monkeypatch) -> None:
    fake = FakeEmbeddings()
    service = _build_service(monkeypatch, fake)

    texts = [f"chunk-{i}" for i in range(44)]
    result = service.generate_embeddings_batch(texts)

    # One HTTP call for the whole document (44 chunks <= batch size), not 44.
    assert len(fake.calls) == 1
    assert fake.calls[0]["input"] == texts
    assert fake.calls[0]["dimensions"] == OPENROUTER_EMBEDDING_DIMENSIONS

    # One 1024-dim vector per input, returned in input order despite the API
    # returning them reversed.
    assert len(result) == len(texts)
    assert all(len(vec) == OPENROUTER_EMBEDDING_DIMENSIONS for vec in result)
    assert [vec[0] for vec in result] == [float(i) for i in range(len(texts))]


@pytest.mark.asyncio
async def test_generate_chunk_embeddings_batches_one_request(monkeypatch) -> None:
    fake = FakeEmbeddings()
    service = _build_service(monkeypatch, fake)

    chunks = [SimpleNamespace(text=f"chunk-{i}") for i in range(10)]
    result = await service.generate_chunk_embeddings(chunks)

    assert len(fake.calls) == 1
    assert fake.calls[0]["input"] == [c.text for c in chunks]
    assert len(result) == len(chunks)


def test_generate_embeddings_batch_splits_oversized_input(monkeypatch) -> None:
    fake = FakeEmbeddings()
    service = _build_service(monkeypatch, fake)

    count = embeddings.EMBEDDING_REQUEST_BATCH_SIZE + 5
    texts = [f"chunk-{i}" for i in range(count)]
    result = service.generate_embeddings_batch(texts)

    # Two sub-batch requests, preserving global order.
    assert len(fake.calls) == 2
    assert (
        fake.calls[0]["input"] + fake.calls[1]["input"] == texts
    )
    assert len(result) == count


def test_generate_embeddings_batch_empty_makes_no_request(monkeypatch) -> None:
    fake = FakeEmbeddings()
    service = _build_service(monkeypatch, fake)

    assert service.generate_embeddings_batch([]) == []
    assert fake.calls == []
