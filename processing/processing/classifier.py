"""Document classification and metadata extraction service."""

import json
import re
from typing import Any, Optional

import openai
import structlog

from docvault_shared.config import config

logger = structlog.get_logger(__name__)

ALLOWED_DOC_TYPES = {"invoice", "contract", "identity", "warranty", "receipt", "other"}
METADATA_FIELDS = (
    "issuer",
    "amount",
    "currency",
    "issue_date",
    "due_date",
    "expiry_date",
    "document_number",
)
MAX_CLASSIFICATION_CHARS = 12_000

CLASSIFICATION_SYSTEM_PROMPT = """You classify OCR text from uploaded documents.
Return strict JSON only, with no Markdown or explanation.

Allowed doc_type values:
- invoice
- contract
- identity
- warranty
- receipt
- other

Rules:
- Pick exactly one doc_type. Use other when uncertain.
- Detect language as a BCP-47 style short code such as en or ar.
- Extract metadata only when explicitly present in the text.
- Use null for unknown metadata values.
- Preserve dates and amounts exactly as written.

Return this JSON shape:
{
  "doc_type": "invoice|contract|identity|warranty|receipt|other",
  "language": "en",
  "issuer": null,
  "amount": null,
  "currency": null,
  "issue_date": null,
  "due_date": null,
  "expiry_date": null,
  "document_number": null
}
"""


class OpenRouterDocumentClassifier:
    """Prompt-based document classifier backed by OpenRouter."""

    def __init__(self, api_key: str | None = None, model: str | None = None):
        """Initialize OpenRouter document classifier."""
        self.api_key = api_key or config.openrouter_api_key
        self.model_name = (model or config.classification_model).strip()
        if not self.model_name:
            raise ValueError("CLASSIFICATION_MODEL must not be empty")
        self.client = openai.OpenAI(
            api_key=self.api_key,
            base_url="https://openrouter.ai/api/v1",
        )

    async def extract_metadata(self, document_id: str, text: str) -> dict:
        """Classify a document and extract metadata."""
        fallback = self._fallback_metadata(text)
        if not text.strip():
            return fallback

        try:
            response = self.client.chat.completions.create(
                model=self.model_name,
                messages=[
                    {"role": "system", "content": CLASSIFICATION_SYSTEM_PROMPT},
                    {"role": "user", "content": self._build_user_prompt(text)},
                ],
                temperature=0,
                max_tokens=500,
            )
            content = response.choices[0].message.content
            metadata = self._normalize_response(content, fallback)
            logger.info(
                "metadata_extracted",
                document_id=document_id,
                doc_type=metadata["doc_type"],
                language=metadata["language"],
                model=self.model_name,
            )
            return metadata

        except Exception as e:
            logger.warning(
                "classification_failed_using_fallback",
                document_id=document_id,
                error=str(e),
                model=self.model_name,
            )
            return fallback

    def _build_user_prompt(self, text: str) -> str:
        sample = text.strip()[:MAX_CLASSIFICATION_CHARS]
        return f"Classify and extract metadata from this OCR text:\n\n<document>\n{sample}\n</document>"

    def _normalize_response(self, content: str | None, fallback: dict) -> dict:
        if not content:
            return fallback

        parsed = self._parse_json(content)
        if not isinstance(parsed, dict):
            return fallback

        doc_type = str(parsed.get("doc_type") or "").strip().lower()
        if doc_type not in ALLOWED_DOC_TYPES:
            doc_type = fallback["doc_type"]

        language = str(parsed.get("language") or "").strip().lower()
        if not language:
            language = fallback["language"]

        metadata = {
            "doc_type": doc_type,
            "language": language,
        }
        for field in METADATA_FIELDS:
            value = parsed.get(field)
            if value is not None and str(value).strip():
                metadata[field] = str(value).strip()

        return metadata

    def _parse_json(self, content: str) -> Any:
        cleaned = content.strip()
        if cleaned.startswith("```"):
            cleaned = re.sub(r"^```(?:json)?\s*", "", cleaned, flags=re.IGNORECASE)
            cleaned = re.sub(r"\s*```$", "", cleaned)
        return json.loads(cleaned)

    def _fallback_metadata(self, text: str) -> dict:
        return {
            "doc_type": "other",
            "language": self._detect_language(text),
        }

    def _detect_language(self, text: str) -> str:
        if re.search(r"[\u0600-\u06FF]", text):
            return "ar"
        return "en"


class MetadataExtractor:
    """Document metadata extraction facade used by the processing job."""

    def __init__(self, classifier: OpenRouterDocumentClassifier | None = None):
        """Initialize the extractor."""
        self.classifier = classifier or OpenRouterDocumentClassifier()

    async def extract(
        self,
        document_id: str,
        text: str,
        existing_metadata: Optional[dict] = None,
    ) -> dict:
        """Extract complete metadata for a document."""
        metadata = await self.classifier.extract_metadata(document_id=document_id, text=text)
        if existing_metadata:
            metadata = {**existing_metadata, **metadata}
        return metadata


metadata_extractor = MetadataExtractor()


TYPE_LABELS = {
    "invoice": "Invoice",
    "contract": "Contract",
    "identity": "ID",
    "warranty": "Warranty",
    "receipt": "Receipt",
    "other": "Document",
}


def generate_display_title(metadata: dict, original_title: str) -> str:
    _, _, ext = original_title.rpartition(".")
    if ext and "." in ext:
        ext = ""

    doc_type = metadata.get("doc_type", "other")
    parts = [TYPE_LABELS.get(doc_type, "Document")]

    doc_num = metadata.get("document_number")
    if doc_num:
        parts.append(doc_num)

    issuer = metadata.get("issuer")
    if issuer:
        parts.append(issuer)

    if len(parts) > 1:
        name = " ".join(parts) + ("." + ext if ext else "")
        return name

    return original_title
