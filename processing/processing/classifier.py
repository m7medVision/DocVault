"""Document classification and metadata extraction service."""

import re
import structlog
from dataclasses import dataclass
from enum import Enum
from typing import Optional

logger = structlog.get_logger(__name__)


class DocumentType(str, Enum):
    """Document type classification."""

    INVOICE = "invoice"
    CONTRACT = "contract"
    IDENTITY = "identity"
    WARRANTY = "warranty"
    RECEIPT = "receipt"
    OTHER = "other"


@dataclass
class DocumentClassification:
    """Document classification result."""

    doc_type: DocumentType
    confidence: float
    language: Optional[str]
    extracted_metadata: dict


@dataclass
class MetadataField:
    """A single extracted metadata field."""

    key: str
    value: Optional[str]
    confidence: float


class DocumentClassifier:
    """Rules-first document classifier."""

    TYPE_PATTERNS = {
        DocumentType.INVOICE: [
            r"invoice\s*#?\s*[:]?\s*[\w-]+",
            r"bill\s+to",
            r"total\s+due",
            r"amount\s+payable",
            r"payment\s+terms",
            r" subtotal ",
            r" tax ",
            r" total ",
        ],
        DocumentType.CONTRACT: [
            r"agreement\s+between",
            r"contract\s+no\.?\s*[:]?\s*[\w-]+",
            r"whereas",
            r"terms\s+and\s+conditions",
            r"parties\s+agree",
            r"effective\s+date",
            r"signature",
        ],
        DocumentType.IDENTITY: [
            r"driver'?s?\s+license",
            r"passport\s+number",
            r"national\s+id",
            r"date\s+of\s+birth",
            r"issued\s+by",
            r"date\s+of\s+issue",
        ],
        DocumentType.WARRANTY: [
            r"warranty\s+period",
            r"serial\s+number",
            r"model\s+number",
            r"warranty\s+void",
            r"limited\s+warranty",
            r"date\s+of\s+purchase",
        ],
        DocumentType.RECEIPT: [
            r"receipt",
            r"transaction\s+id",
            r"paid\s+on",
            r"cashier",
            r"register",
            r"purchase\s+date",
        ],
    }

    METADATA_PATTERNS = {
        "issuer": [
            r"from\s*[:]\s*([A-Z][A-Za-z\s]+)",
            r"issued\s+by\s*[:]\s*([A-Z][A-Za-z\s]+)",
            r"company\s*[:]\s*([A-Z][A-Za-z\s]+)",
        ],
        "amount": [
            r"(?:total\s*(?:due|amount)?|subtotal|sum|amount\s*(?:due|payable)?)\s*[:.]?\s*\$?\s*(\d{1,3}(?:[,.]?\d{0,2})*\.?\d{0,2})",
            r"\$\s*(\d{1,3}(?:[,.]?\d{0,2})*\.?\d{0,2})",
        ],
        "currency": [
            r"(USD|EUR|GBP|SAR|AED|JPY|CNY)",
            r"(dollars?|euros?|pounds?|riyals?)",
        ],
        "issue_date": [
            r"(?:dated?|dated?\s+on)\s+[:.]?\s*(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})",
            r"(?:issued?\s+on|effective\s+date)\s+[:.]?\s*(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})",
            r"(\d{4}-\d{2}-\d{2})",
        ],
        "due_date": [
            r"(?:due\s*(?:date|on|by)|payment\s*due)\s*[:.]?\s*(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})",
            r"(?:due\s*(?:date|on|by)|payment\s*due)\s*[:.]?\s*([A-Z][a-z]+ \d{1,2},?\s*\d{4})",
        ],
        "expiry_date": [
            r"(?:expires?|valid\s+until|expiration\s+date)\s+[:.]?\s*(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})",
            r"(\d{4}-\d{2}-\d{2})",
        ],
        "document_number": [
            r"(?:invoice|contract|receipt|warranty)\s*(?:#|no\.?|number\s*[:.]?)\s*([\w-]+)",
            r"(?:ID|Ref|Reference)\s*[:.]?\s*([\w-]+)",
        ],
    }

    def classify(self, text: str) -> DocumentClassification:
        """Classify a document and extract metadata."""
        language = self._detect_language(text)
        doc_type, type_confidence = self._classify_type(text)
        metadata = self._extract_metadata(text)

        return DocumentClassification(
            doc_type=doc_type,
            confidence=type_confidence,
            language=language,
            extracted_metadata=metadata,
        )

    def _detect_language(self, text: str) -> str:
        """Detect the language of the text."""
        if re.search(r"[\u0600-\u06FF]", text):
            return "ar"
        return "en"

    def _classify_type(self, text: str) -> tuple[DocumentType, float]:
        """Classify the document type."""
        text_lower = text.lower()
        scores = {}

        for doc_type, patterns in self.TYPE_PATTERNS.items():
            matches = 0
            for pattern in patterns:
                if re.search(pattern, text_lower, re.IGNORECASE):
                    matches += 1
            if matches > 0:
                scores[doc_type] = min(matches / len(patterns) * 2, 1.0)

        if not scores:
            return DocumentType.OTHER, 0.5

        best_type = max(scores.items(), key=lambda x: x[1])
        return best_type[0], best_type[1]

    def _extract_metadata(self, text: str) -> dict:
        """Extract metadata fields from text."""
        metadata = {}

        for field, patterns in self.METADATA_PATTERNS.items():
            value = self._extract_first_match(text, patterns)
            if value:
                metadata[field] = value

        return metadata

    def _extract_first_match(self, text: str, patterns: list[str]) -> Optional[str]:
        """Extract the first matching group from patterns."""
        for pattern in patterns:
            match = re.search(pattern, text, re.IGNORECASE)
            if match and match.groups():
                return match.group(1).strip()
        return None


class MetadataExtractor:
    """Advanced metadata extraction using rules and patterns."""

    def __init__(self):
        """Initialize the extractor."""
        self.classifier = DocumentClassifier()

    async def extract(
        self,
        document_id: str,
        text: str,
        existing_metadata: Optional[dict] = None,
    ) -> dict:
        """Extract complete metadata for a document."""
        classification = self.classifier.classify(text)

        metadata = {
            "doc_type": classification.doc_type.value,
            "language": classification.language,
            **classification.extracted_metadata,
        }

        logger.info(
            "metadata_extracted",
            document_id=document_id,
            doc_type=classification.doc_type.value,
            has_amount="amount" in metadata,
            language=classification.language,
        )

        return metadata


classifier = DocumentClassifier()
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
