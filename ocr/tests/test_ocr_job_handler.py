"""Tests for the OCRJobHandler use case.

These exercise the orchestration with mocked Protocol ports (OCRClient,
OCRPersistence, MessagePublisher) so no live Mistral / MinIO / RabbitMQ is
needed. They assert the observable behavior: status updates, OCR call,
persistence, and the downstream processing publish.
"""

from datetime import datetime
from unittest.mock import AsyncMock, MagicMock, call

import pytest

from docvault_shared.models import OCRPage, OCRResult
from ocr.application.ocr_job import OCRJobHandler


PROCESSING_QUEUE = "docvault.processing.jobs"


def _message() -> dict:
    return {
        "document_id": "doc-1",
        "version_id": "ver-1",
        "storage_key": "tenant/doc-1.pdf",
        "tenant_id": "tenant-1",
        "org_id": "org-1",
        "mime_type": "application/pdf",
        "language": "en",
    }


def _result() -> OCRResult:
    return OCRResult(
        document_id="doc-1",
        version_id="ver-1",
        pages=[
            OCRPage(page_number=1, text="hello", confidence=0.99, model="mistral-ocr-2503"),
            OCRPage(page_number=2, text="world", confidence=0.95, model="mistral-ocr-2503"),
        ],
        completed_at=datetime(2026, 1, 1),
    )


def _make_handler(result: OCRResult, low_confidence: list[int]):
    ocr = MagicMock()
    ocr.process_document = AsyncMock(return_value=result)
    ocr.flag_low_confidence_pages = MagicMock(return_value=low_confidence)

    storage = MagicMock()
    storage.update_document_status = AsyncMock(return_value=None)
    storage.save_ocr_results = AsyncMock(return_value=["page-1", "page-2"])

    publisher = MagicMock()
    publisher.publish = MagicMock(return_value=None)

    handler = OCRJobHandler(ocr, storage, publisher, PROCESSING_QUEUE)
    return handler, ocr, storage, publisher


async def test_handle_runs_full_pipeline_and_publishes():
    result = _result()
    handler, ocr, storage, publisher = _make_handler(result, low_confidence=[])

    await handler.handle(_message())

    # Status moves to "processing" before and after OCR (behavior preserved).
    assert storage.update_document_status.await_args_list == [
        call("doc-1", "processing"),
        call("doc-1", "processing"),
    ]

    # OCR was invoked once with a job carrying the message fields.
    assert ocr.process_document.await_count == 1
    job = ocr.process_document.await_args.args[0]
    assert job.document_id == "doc-1"
    assert job.storage_key == "tenant/doc-1.pdf"
    assert job.mime_type == "application/pdf"

    # Results were persisted.
    storage.save_ocr_results.assert_awaited_once_with(result)

    # Downstream processing message published to the injected queue.
    publisher.publish.assert_called_once()
    pub_kwargs = publisher.publish.call_args.kwargs
    assert pub_kwargs["queue"] == PROCESSING_QUEUE
    published = pub_kwargs["message"]
    assert published["document_id"] == "doc-1"
    assert published["version_id"] == "ver-1"
    assert published["tenant_id"] == "tenant-1"
    assert published["org_id"] == "org-1"
    assert published["language"] == "en"
    assert published["retry_count"] == 0
    assert published["page_ids"] == ["page-1", "page-2"]
    assert published["low_confidence_pages"] == []
    assert [p["page_number"] for p in published["pages"]] == [1, 2]
    assert published["pages"][0]["text"] == "hello"
    assert published["pages"][0]["confidence"] == 0.99
    assert published["pages"][0]["model"] == "mistral-ocr-2503"


async def test_handle_propagates_low_confidence_pages():
    result = _result()
    handler, _ocr, _storage, publisher = _make_handler(result, low_confidence=[2])

    await handler.handle(_message())

    published = publisher.publish.call_args.kwargs["message"]
    assert published["low_confidence_pages"] == [2]


async def test_handle_defaults_missing_optional_fields():
    msg = _message()
    del msg["mime_type"]
    del msg["language"]
    result = _result()
    handler, ocr, _storage, publisher = _make_handler(result, low_confidence=[])

    await handler.handle(msg)

    job = ocr.process_document.await_args.args[0]
    assert job.mime_type == "application/octet-stream"
    assert job.language is None
    assert publisher.publish.call_args.kwargs["message"]["language"] is None


@pytest.mark.parametrize(
    "missing", ["document_id", "version_id", "storage_key", "tenant_id", "org_id"]
)
async def test_handle_rejects_missing_required_field(missing):
    msg = _message()
    del msg[missing]
    handler, ocr, storage, publisher = _make_handler(_result(), low_confidence=[])

    with pytest.raises(ValueError, match=f"Missing required field: {missing}"):
        await handler.handle(msg)

    # Nothing downstream should run when validation fails.
    ocr.process_document.assert_not_awaited()
    storage.update_document_status.assert_not_awaited()
    publisher.publish.assert_not_called()
