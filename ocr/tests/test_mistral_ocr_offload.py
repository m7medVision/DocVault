"""Tests for MistralOCRClient.process_document offloading blocking SDK calls.

The Mistral SDK is synchronous; process_document now wraps each call in
asyncio.to_thread so the event loop is not blocked. These tests build a client
without running __init__, inject a MagicMock sync SDK and an AsyncMock blob
store, and assert the observable behavior is preserved: the bytes are
downloaded, uploaded, OCR'd via the signed URL, the temporary upload is
deleted in a finally block (even on failure), and the parsed result is
returned unchanged. asyncio.to_thread over a mocked sync method works because
the mock is just an ordinary callable.
"""

from datetime import datetime
from types import SimpleNamespace
from unittest.mock import AsyncMock, MagicMock

import pytest

from docvault_shared.models import OCRJob
from ocr.ocr import MistralOCRClient


def _job() -> OCRJob:
    return OCRJob(
        document_id="doc-1",
        version_id="ver-1",
        storage_key="tenant/doc-1.pdf",
        mime_type="application/pdf",
        tenant_id="tenant-1",
        org_id="org-1",
        language="en",
    )


def _make_client():
    """A MistralOCRClient with a mocked sync SDK and async blob store."""
    client = object.__new__(MistralOCRClient)
    client.model = "mistral-ocr-2503"

    sdk = MagicMock()
    sdk.files.upload = MagicMock(return_value=SimpleNamespace(id="file-123"))
    sdk.files.get_signed_url = MagicMock(
        return_value=SimpleNamespace(url="https://signed.example/doc")
    )
    sdk.ocr.process = MagicMock(
        return_value=SimpleNamespace(
            pages=[
                {"index": 0, "text": "first", "confidence": 0.9},
                {"index": 1, "text": "second", "confidence": 0.8},
            ]
        )
    )
    sdk.files.delete = MagicMock(return_value=None)
    client.client = sdk

    storage = MagicMock()
    storage.get_object = AsyncMock(return_value=b"pdf-bytes")
    client.storage = storage

    return client, sdk, storage


async def test_process_document_offloads_and_preserves_behavior():
    client, sdk, storage = _make_client()

    result = await client.process_document(_job())

    # Blob store read happened once with the job's key.
    storage.get_object.assert_awaited_once_with("tenant/doc-1.pdf")

    # Upload carried the downloaded bytes, file name, mime type, purpose.
    sdk.files.upload.assert_called_once()
    upload_file = sdk.files.upload.call_args.kwargs["file"]
    assert upload_file["file_name"] == "doc-1.pdf"
    assert upload_file["content"] == b"pdf-bytes"
    assert upload_file["content_type"] == "application/pdf"
    assert sdk.files.upload.call_args.kwargs["purpose"] == "ocr"

    # Signed URL fetched for the uploaded file, then OCR ran against that URL.
    sdk.files.get_signed_url.assert_called_once_with(file_id="file-123")
    sdk.ocr.process.assert_called_once()
    ocr_kwargs = sdk.ocr.process.call_args.kwargs
    assert ocr_kwargs["model"] == "mistral-ocr-2503"
    assert ocr_kwargs["document"]["document_url"] == "https://signed.example/doc"
    assert ocr_kwargs["id"] == "ver-1"

    # Temporary upload cleaned up.
    sdk.files.delete.assert_called_once_with(file_id="file-123")

    # Parsed result is returned unchanged.
    assert result.document_id == "doc-1"
    assert result.version_id == "ver-1"
    assert [p.page_number for p in result.pages] == [1, 2]
    assert [p.text for p in result.pages] == ["first", "second"]
    assert isinstance(result.completed_at, datetime)


async def test_process_document_deletes_upload_even_when_ocr_fails():
    client, sdk, storage = _make_client()
    sdk.ocr.process = MagicMock(side_effect=RuntimeError("ocr boom"))

    with pytest.raises(RuntimeError, match="ocr boom"):
        await client.process_document(_job())

    # Cleanup still runs in the finally block.
    sdk.files.delete.assert_called_once_with(file_id="file-123")


async def test_process_document_swallows_cleanup_failure():
    client, sdk, _storage = _make_client()
    sdk.files.delete = MagicMock(side_effect=RuntimeError("delete boom"))

    # A failing best-effort delete must not mask a successful OCR result.
    result = await client.process_document(_job())

    assert [p.page_number for p in result.pages] == [1, 2]
    sdk.files.delete.assert_called_once_with(file_id="file-123")
