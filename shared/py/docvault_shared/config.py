"""Configuration for DocVault Python services."""

import os
from dataclasses import dataclass
from pathlib import Path
from urllib.parse import quote, urlsplit, urlunsplit

from dotenv import load_dotenv


def _load_repo_env() -> None:
    """Load the shared repository root .env file."""
    repo_root = Path(__file__).resolve().parents[2]
    load_dotenv(repo_root / ".env", override=False)


_load_repo_env()


DEFAULT_EMBEDDING_MODEL = "mistralai/mistral-embed-2312"
APPROVED_OPENROUTER_EMBEDDING_MODELS = (DEFAULT_EMBEDDING_MODEL,)
OPENROUTER_EMBEDDING_DIMENSIONS = 1024
DEFAULT_TRANSLATION_MODEL = "mistralai/mistral-large"


@dataclass
class Config:
    """Application configuration."""

    environment: str

    rabbitmq_url: str
    rabbitmq_queue_ocr: str
    rabbitmq_queue_processing: str
    rabbitmq_queue_reminder: str

    database_url: str

    minio_endpoint: str
    minio_access_key: str
    minio_secret_key: str
    minio_bucket: str
    minio_secure: bool

    mistral_api_key: str
    mistral_endpoint: str

    openrouter_api_key: str
    embedding_model: str
    translation_model: str

    chunk_size: int
    chunk_overlap: int

    min_confidence_threshold: float

    sentry_dsn: str


def validate_openrouter_model_name(
    model_name: str,
    allowed_models: tuple[str, ...],
    setting_name: str = "EMBEDDING_MODEL",
) -> str:
    """Allow only explicitly approved OpenRouter models."""
    normalized = model_name.strip()
    if not normalized:
        raise ValueError(f"{setting_name} must not be empty")
    if normalized not in allowed_models:
        allowed = ", ".join(allowed_models)
        raise ValueError(f"{setting_name} must be one of: {allowed}. Received: {model_name!r}.")
    return normalized


def _normalize_rabbitmq_url(url: str) -> str:
    """Support the legacy `//vhost` format used in local env files."""
    parsed = urlsplit(url)
    if parsed.scheme not in {"amqp", "amqps"} or not parsed.path.startswith("//"):
        return url

    encoded_vhost = "/" + quote(parsed.path[1:], safe="")
    return urlunsplit(parsed._replace(path=encoded_vhost))


def _validate_runtime_config(loaded_config: Config) -> None:
    """Require the AI key outside local development."""
    if loaded_config.environment == "development":
        return
    if not loaded_config.openrouter_api_key:
        raise ValueError("OPENROUTER_API_KEY is required outside development")


def load() -> Config:
    """Load configuration from environment variables."""
    loaded_config = Config(
        environment=os.getenv("ENVIRONMENT", "development"),
        rabbitmq_url=_normalize_rabbitmq_url(
            os.getenv("RABBITMQ_URL", "amqp://docvault:changeme@localhost:5672/%2Fdocvault")
        ),
        rabbitmq_queue_ocr=os.getenv("RABBITMQ_QUEUE_OCR", "docvault.ocr.jobs"),
        rabbitmq_queue_processing=os.getenv(
            "RABBITMQ_QUEUE_PROCESSING", "docvault.processing.jobs"
        ),
        rabbitmq_queue_reminder=os.getenv("RABBITMQ_QUEUE_REMINDER", "docvault.reminder.jobs"),
        database_url=os.getenv(
            "DATABASE_URL", "postgresql://docvault:docvault_dev@localhost:5432/docvault"
        ),
        minio_endpoint=os.getenv("MINIO_ENDPOINT", "localhost:9000"),
        minio_access_key=os.getenv("MINIO_ACCESS_KEY", "minioadmin"),
        minio_secret_key=os.getenv("MINIO_SECRET_KEY", "minioadmin"),
        minio_bucket=os.getenv("MINIO_BUCKET", "docvault-documents"),
        minio_secure=os.getenv("MINIO_USE_SSL", os.getenv("MINIO_SECURE", "false")).lower()
        == "true",
        mistral_api_key=os.getenv("MISTRAL_API_KEY", ""),
        mistral_endpoint=os.getenv("MISTRAL_ENDPOINT", "https://api.mistral.ai/v1/ocr"),
        openrouter_api_key=os.getenv("OPENROUTER_API_KEY", ""),
        embedding_model=validate_openrouter_model_name(
            os.getenv("EMBEDDING_MODEL", DEFAULT_EMBEDDING_MODEL),
            APPROVED_OPENROUTER_EMBEDDING_MODELS,
        ),
        translation_model=os.getenv("TRANSLATION_MODEL", DEFAULT_TRANSLATION_MODEL).strip(),
        chunk_size=int(os.getenv("CHUNK_SIZE", "500")),
        chunk_overlap=int(os.getenv("CHUNK_OVERLAP", "50")),
        min_confidence_threshold=float(os.getenv("MIN_CONFIDENCE_THRESHOLD", "0.7")),
        sentry_dsn=os.getenv("SENTRY_DSN_PROCESSING", ""),
    )
    _validate_runtime_config(loaded_config)
    return loaded_config


config = load()
