from types import SimpleNamespace
from unittest.mock import AsyncMock, Mock
from unittest.mock import patch

import pytest

from processing.application.processing_job import ProcessingJobHandler


@pytest.mark.asyncio
async def test_processing_job_loads_pages_from_storage_for_legacy_message() -> None:
    chunker = Mock()
    chunker.chunk_ocr_result.return_value = SimpleNamespace(chunks=[])
    classifier = Mock()
    classifier.extract = AsyncMock(return_value={"language": "en", "doc_type": "invoice"})
    pg_repo = Mock()
    pg_repo.delete_by_document = AsyncMock()
    pg_repo.save_chunks = AsyncMock()
    ocr_repo = Mock()
    ocr_repo.get_document_pages = AsyncMock(
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
    ocr_repo.update_document_metadata = AsyncMock()
    ocr_repo.update_document_status = AsyncMock()
    reminder_publisher = Mock()
    reminder_publisher.publish_after_processing = AsyncMock()
    translator = Mock()
    translator.translate_document = AsyncMock()

    handler = ProcessingJobHandler(
        chunker=chunker,
        embedder=AsyncMock(),
        classifier=classifier,
        pg_repo=pg_repo,
        ocr_repo=ocr_repo,
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

    ocr_repo.get_document_pages.assert_awaited_once_with("doc-1", "ver-1")
    chunk_args, chunk_kwargs = chunker.chunk_ocr_result.call_args
    assert chunk_kwargs["page_ids"] == {1: "page-1"}
    assert chunk_args[1][0].page_number == 1
    assert chunk_args[1][0].text == "hello world"


@pytest.mark.asyncio
async def test_processing_job_raises_when_legacy_message_cannot_be_hydrated() -> None:
    chunker = Mock()
    classifier = Mock()
    classifier.extract = AsyncMock()
    pg_repo = Mock()
    ocr_repo = Mock()
    ocr_repo.get_document_pages = AsyncMock(return_value=[])
    reminder_publisher = Mock()
    translator = Mock()

    handler = ProcessingJobHandler(
        chunker=chunker,
        embedder=AsyncMock(),
        classifier=classifier,
        pg_repo=pg_repo,
        ocr_repo=ocr_repo,
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
    pg_repo = Mock()
    ocr_repo = Mock()
    ocr_repo.get_document_pages = AsyncMock(return_value=[])
    reminder_publisher = Mock()
    translator = Mock()

    handler = ProcessingJobHandler(
        chunker=chunker,
        embedder=AsyncMock(),
        classifier=classifier,
        pg_repo=pg_repo,
        ocr_repo=ocr_repo,
        reminder_publisher=reminder_publisher,
        connection=Mock(),
        translator=translator,
    )

    with patch("processing.application.processing_job.QueuePublisher") as publisher_cls:
        publisher = publisher_cls.return_value

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

    ocr_repo.get_document_pages.assert_awaited_once_with("doc-1", "ver-1")
    publisher_cls.assert_called_once_with(connection=handler._connection)
    publisher.publish.assert_called_once()
    _, publish_kwargs = publisher.publish.call_args
    assert publish_kwargs["queue"] == "docvault.ocr.jobs"
    assert publish_kwargs["message"]["retry_count"] == 0
    chunker.chunk_ocr_result.assert_not_called()
