"""Tests for MinIOClient offloading blocking boto3 calls off the event loop.

boto3 is synchronous; the async methods now wrap each call in
asyncio.to_thread. We build a client without running __init__, inject a
MagicMock boto3 client, and assert the calls are made with the expected
arguments and the return values are unchanged. asyncio.to_thread over a mocked
sync method works because the mock is just an ordinary callable.
"""

from unittest.mock import MagicMock

from botocore.exceptions import ClientError

from ocr.storage import MinIOClient


def _make_client():
    client = object.__new__(MinIOClient)
    client.bucket = "docvault"
    client.client = MagicMock()
    return client


async def test_get_object_reads_body_off_the_loop():
    client = _make_client()
    body = MagicMock()
    body.read.return_value = b"file-bytes"
    client.client.get_object.return_value = {"Body": body}

    data = await client.get_object("tenant/doc-1.pdf")

    assert data == b"file-bytes"
    client.client.get_object.assert_called_once_with(
        Bucket="docvault", Key="tenant/doc-1.pdf"
    )
    body.read.assert_called_once_with()


async def test_put_object_offloads_call():
    client = _make_client()

    await client.put_object("tenant/doc-1.pdf", b"data", "application/pdf")

    client.client.put_object.assert_called_once_with(
        Bucket="docvault",
        Key="tenant/doc-1.pdf",
        Body=b"data",
        ContentType="application/pdf",
    )


async def test_generate_presigned_url_returns_value():
    client = _make_client()
    client.client.generate_presigned_url.return_value = "https://signed/url"

    url = await client.generate_presigned_url("tenant/doc-1.pdf", expires_in=60)

    assert url == "https://signed/url"
    client.client.generate_presigned_url.assert_called_once_with(
        "get_object",
        Params={"Bucket": "docvault", "Key": "tenant/doc-1.pdf"},
        ExpiresIn=60,
    )


async def test_delete_object_offloads_call():
    client = _make_client()

    await client.delete_object("tenant/doc-1.pdf")

    client.client.delete_object.assert_called_once_with(
        Bucket="docvault", Key="tenant/doc-1.pdf"
    )


async def test_object_exists_true_then_false():
    client = _make_client()

    assert await client.object_exists("tenant/doc-1.pdf") is True
    client.client.head_object.assert_called_once_with(
        Bucket="docvault", Key="tenant/doc-1.pdf"
    )

    client.client.head_object.side_effect = ClientError(
        {"Error": {"Code": "404"}}, "HeadObject"
    )
    assert await client.object_exists("missing.pdf") is False
