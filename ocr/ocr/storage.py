"""MinIO/S3 storage client for document binary access."""

import io
import structlog
from typing import Optional

import boto3
from botocore.config import Config as BotoConfig
from botocore.exceptions import ClientError

from docvault_shared.config import config

logger = structlog.get_logger(__name__)


class MinIOClient:
    """Client for MinIO object storage operations."""

    def __init__(self):
        """Initialize MinIO client."""
        endpoint = config.minio_endpoint.removeprefix("https://").removeprefix("http://")
        scheme = "https" if config.minio_secure else "http"
        self.client = boto3.client(
            "s3",
            endpoint_url=f"{scheme}://{endpoint}",
            aws_access_key_id=config.minio_access_key,
            aws_secret_access_key=config.minio_secret_key,
            config=BotoConfig(
                signature_version="s3v4",
                retries={"max_attempts": 3, "mode": "standard"},
            ),
        )
        self.bucket = config.minio_bucket

    async def get_object(self, storage_key: str) -> bytes:
        """
        Download an object from MinIO.

        Args:
            storage_key: The key of the object to download.

        Returns:
            The object content as bytes.

        Raises:
            ClientError: If the object cannot be downloaded.
        """
        try:
            response = self.client.get_object(Bucket=self.bucket, Key=storage_key)
            return response["Body"].read()
        except ClientError as e:
            logger.error("failed_to_download_object", storage_key=storage_key, error=str(e))
            raise

    async def put_object(self, storage_key: str, content: bytes, content_type: str) -> None:
        """
        Upload an object to MinIO.

        Args:
            storage_key: The key to store the object under.
            content: The content to upload.
            content_type: The MIME type of the content.
        """
        try:
            self.client.put_object(
                Bucket=self.bucket,
                Key=storage_key,
                Body=content,
                ContentType=content_type,
            )
            logger.info("object_uploaded", storage_key=storage_key)
        except ClientError as e:
            logger.error("failed_to_upload_object", storage_key=storage_key, error=str(e))
            raise

    async def generate_presigned_url(self, storage_key: str, expires_in: int = 3600) -> str:
        """
        Generate a presigned URL for temporary access.

        Args:
            storage_key: The key of the object.
            expires_in: URL expiration time in seconds.

        Returns:
            Presigned URL string.
        """
        try:
            url = self.client.generate_presigned_url(
                "get_object",
                Params={"Bucket": self.bucket, "Key": storage_key},
                ExpiresIn=expires_in,
            )
            return url
        except ClientError as e:
            logger.error(
                "failed_to_generate_presigned_url",
                storage_key=storage_key,
                error=str(e),
            )
            raise

    async def delete_object(self, storage_key: str) -> None:
        """
        Delete an object from MinIO.

        Args:
            storage_key: The key of the object to delete.
        """
        try:
            self.client.delete_object(Bucket=self.bucket, Key=storage_key)
            logger.info("object_deleted", storage_key=storage_key)
        except ClientError as e:
            logger.error("failed_to_delete_object", storage_key=storage_key, error=str(e))
            raise

    async def object_exists(self, storage_key: str) -> bool:
        """
        Check if an object exists in MinIO.

        Args:
            storage_key: The key to check.

        Returns:
            True if the object exists, False otherwise.
        """
        try:
            self.client.head_object(Bucket=self.bucket, Key=storage_key)
            return True
        except ClientError:
            return False


minio_client = MinIOClient()
