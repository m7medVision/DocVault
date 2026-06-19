"""Tests for MistralOCRClient response parsing and low-confidence flagging.

The pure methods (_parse_ocr_response, flag_low_confidence_pages) don't need
the live Mistral SDK, so we build an instance without running __init__ and set
just the attributes those methods read. This keeps the tests hermetic.
"""

from datetime import datetime
from types import SimpleNamespace

from docvault_shared.config import config
from docvault_shared.models import OCRPage, OCRResult
from ocr.ocr import MistralOCRClient


def _client() -> MistralOCRClient:
    """A MistralOCRClient with only the attributes the pure methods use."""
    client = object.__new__(MistralOCRClient)
    client.model = "mistral-ocr-2503"
    return client


def test_parse_response_dict_pages_uses_index_for_page_number():
    response = SimpleNamespace(
        pages=[
            {"index": 0, "text": "first", "confidence": 0.9},
            {"index": 1, "markdown": "second", "confidence": 0.4},
        ]
    )

    pages = _client()._parse_ocr_response(response, "doc-1", "ver-1")

    assert len(pages) == 2
    assert pages[0].page_number == 1
    assert pages[0].text == "first"
    assert pages[0].confidence == 0.9
    assert pages[0].model == "mistral-ocr-2503"
    # falls back to "markdown" when "text" is absent
    assert pages[1].page_number == 2
    assert pages[1].text == "second"
    assert pages[1].confidence == 0.4


def test_parse_response_object_pages():
    page = SimpleNamespace(index=4, text="hello", confidence=0.81)
    response = SimpleNamespace(pages=[page])

    pages = _client()._parse_ocr_response(response, "doc-1", "ver-1")

    assert len(pages) == 1
    assert pages[0].page_number == 5
    assert pages[0].text == "hello"
    assert pages[0].confidence == 0.81


def test_parse_response_without_pages_attr_falls_back_to_single_page():
    # A bare object with no "pages" attribute -> one degraded page.
    class NoPages:
        def __str__(self):
            return "raw-text"

    pages = _client()._parse_ocr_response(NoPages(), "doc-1", "ver-1")

    assert len(pages) == 1
    assert pages[0].page_number == 1
    assert pages[0].text == "raw-text"
    assert pages[0].confidence == 0.5


def test_flag_low_confidence_pages_returns_pages_below_threshold():
    threshold = config.min_confidence_threshold
    result = OCRResult(
        document_id="doc-1",
        version_id="ver-1",
        pages=[
            OCRPage(page_number=1, text="a", confidence=threshold + 0.1),
            OCRPage(page_number=2, text="b", confidence=threshold - 0.1),
            OCRPage(page_number=3, text="c", confidence=threshold - 0.5),
        ],
        completed_at=datetime(2026, 1, 1),
    )

    low = _client().flag_low_confidence_pages(result)

    assert low == [2, 3]


def test_flag_low_confidence_pages_empty_when_all_above_threshold():
    threshold = config.min_confidence_threshold
    result = OCRResult(
        document_id="doc-1",
        version_id="ver-1",
        pages=[
            OCRPage(page_number=1, text="a", confidence=threshold + 0.01),
            OCRPage(page_number=2, text="b", confidence=1.0),
        ],
        completed_at=datetime(2026, 1, 1),
    )

    assert _client().flag_low_confidence_pages(result) == []
