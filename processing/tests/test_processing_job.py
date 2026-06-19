from types import SimpleNamespace
from unittest.mock import AsyncMock, Mock

import pytest

from processing.application.processing_job import ProcessingJobHandler


def _make_repo() -> Mock:
    """A single repository mock implementing every method the handler calls."""
    repo = Mock()
    repo.update_processing_stage = AsyncMock()
    repo.update_page_translations = AsyncMock()
    repo.delete_by_document = AsyncMock()
    repo.save_chunks = AsyncMock()
    repo.upsert_metadata_rows = AsyncMock()
    repo.update_document_metadata = AsyncMock()
    repo.update_document_status = AsyncMock()
    repo.suggest_organization = AsyncMock()
    repo.get_document_title = AsyncMock(return_value="test.pdf")
    return repo


@pytest.mark.asyncio
async def test_processing_job_loads_pages_from_storage_for_legacy_message() -> None:
    chunker = Mock()
    chunker.chunk_ocr_result.return_value = SimpleNamespace(chunks=[])
    classifier = Mock()
    classifier.extract = AsyncMock(return_value={"language": "en", "doc_type": "invoice"})
    repo = _make_repo()
    repo.get_document_pages = AsyncMock(
        return_value=[
            {
                "id": "page-1",
                "page_number": 1,
                "text": "hello world",
                "confidence": 0.99,
                "model": "mistral-ocr-2503",
            }
        ]
    )
    reminder_publisher = Mock()
    reminder_publisher.publish_after_processing = AsyncMock()
    translator = Mock()
    translator.translate_document = AsyncMock()

    handler = ProcessingJobHandler(
        chunker=chunker,
        embedder=AsyncMock(),
        classifier=classifier,
        repo=repo,
        reminder_publisher=reminder_publisher,
        connection=None,
        translator=translator,
    )

    await handler.handle(
        {
            "document_id": "doc-1",
            "version_id": "ver-1",
            "tenant_id": "tenant-1",
            "org_id": "org-1",
        }
    )

    repo.get_document_pages.assert_awaited_once_with("doc-1", "ver-1")
    chunk_args, chunk_kwargs = chunker.chunk_ocr_result.call_args
    assert chunk_kwargs["page_ids"] == {1: "page-1"}
    assert chunk_args[1][0].page_number == 1
    assert chunk_args[1][0].text == "hello world"

    repo.suggest_organization.assert_awaited_once()
    _, suggest_kwargs = repo.suggest_organization.call_args
    assert suggest_kwargs["document_id"] == "doc-1"
    assert suggest_kwargs["suggested_folder_name"] == "Invoices"
    assert suggest_kwargs["suggested_filename"] == "test.pdf"
    assert suggest_kwargs["suggestion_confidence"] == 0.7


@pytest.mark.asyncio
async def test_processing_job_raises_when_legacy_message_cannot_be_hydrated() -> None:
    chunker = Mock()
    classifier = Mock()
    classifier.extract = AsyncMock()
    repo = _make_repo()
    repo.get_document_pages = AsyncMock(return_value=[])
    reminder_publisher = Mock()
    translator = Mock()

    handler = ProcessingJobHandler(
        chunker=chunker,
        embedder=AsyncMock(),
        classifier=classifier,
        repo=repo,
        reminder_publisher=reminder_publisher,
        connection=None,
        translator=translator,
    )

    with pytest.raises(ValueError, match="Missing required field: pages"):
        await handler.handle(
            {
                "document_id": "doc-1",
                "version_id": "ver-1",
                "tenant_id": "tenant-1",
                "org_id": "org-1",
            }
        )


@pytest.mark.asyncio
async def test_processing_job_reroutes_misrouted_ocr_job_to_ocr_queue() -> None:
    chunker = Mock()
    classifier = Mock()
    classifier.extract = AsyncMock()
    repo = _make_repo()
    repo.get_document_pages = AsyncMock(return_value=[])
    reminder_publisher = Mock()
    translator = Mock()

    publisher = Mock()
    publisher_factory = Mock(return_value=publisher)
    connection = Mock()

    handler = ProcessingJobHandler(
        chunker=chunker,
        embedder=AsyncMock(),
        classifier=classifier,
        repo=repo,
        reminder_publisher=reminder_publisher,
        connection=connection,
        translator=translator,
        publisher_factory=publisher_factory,
    )

    await handler.handle(
        {
            "document_id": "doc-1",
            "version_id": "ver-1",
            "tenant_id": "tenant-1",
            "org_id": "org-1",
            "storage_key": "tenant/org/doc/ver/file.pdf",
            "mime_type": "application/pdf",
            "retry_count": 2,
        }
    )

    repo.get_document_pages.assert_awaited_once_with("doc-1", "ver-1")
    publisher_factory.assert_called_once_with(connection)
    publisher.publish.assert_called_once()
    _, publish_kwargs = publisher.publish.call_args
    assert publish_kwargs["queue"] == "docvault.ocr.jobs"
    assert publish_kwargs["message"]["retry_count"] == 0
    chunker.chunk_ocr_result.assert_not_called()


@pytest.mark.asyncio
async def test_processing_job_happy_path_runs_full_pipeline_in_order() -> None:
    """Inline-pages job: classify -> chunk/embed -> persist -> suggest -> reminder."""
    chunk = SimpleNamespace(text="hello world", page_id="page-1", chunk_index=0)
    chunker = Mock()
    chunker.chunk_ocr_result.return_value = SimpleNamespace(chunks=[chunk])
    classifier = Mock()
    classifier.extract = AsyncMock(
        return_value={"language": "en", "doc_type": "invoice", "issuer": "Acme"}
    )
    embedder = AsyncMock(return_value=[[0.1] * 1024])
    repo = _make_repo()
    repo.get_document_title = AsyncMock(return_value="scan.pdf")
    reminder_publisher = Mock()
    reminder_publisher.publish_after_processing = AsyncMock()
    translator = Mock()
    translator.translate_document = AsyncMock()

    handler = ProcessingJobHandler(
        chunker=chunker,
        embedder=embedder,
        classifier=classifier,
        repo=repo,
        reminder_publisher=reminder_publisher,
        connection=None,
        translator=translator,
    )

    await handler.handle(
        {
            "document_id": "doc-1",
            "version_id": "ver-1",
            "tenant_id": "tenant-1",
            "org_id": "org-1",
            "pages": [{"id": "page-1", "page_number": 1, "text": "hello world"}],
        }
    )

    # English doc: no translation API call, no re-classify (single extract).
    translator.translate_document.assert_not_awaited()
    classifier.extract.assert_awaited_once()

    # Embedding ran over the produced chunks.
    embedder.assert_awaited_once_with([chunk])
    repo.save_chunks.assert_awaited_once()
    _, save_kwargs = repo.save_chunks.call_args
    assert save_kwargs["chunks"] == [chunk]
    assert save_kwargs["embeddings"] == [[0.1] * 1024]

    # Stage transitions happen in order.
    stages = [c.args[1] for c in repo.update_processing_stage.call_args_list]
    assert stages == ["processing_running", "indexing", "completed"]

    # Suggestion + reminder + terminal status.
    repo.suggest_organization.assert_awaited_once()
    reminder_publisher.publish_after_processing.assert_awaited_once()
    repo.update_document_status.assert_awaited_once_with("doc-1", "processed")
