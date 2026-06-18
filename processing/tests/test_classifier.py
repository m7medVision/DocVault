from types import SimpleNamespace

import pytest

from processing import classifier


class FakeCompletions:
    def __init__(self, content: str | None = None, error: Exception | None = None):
        self.content = content
        self.error = error
        self.calls = []

    def create(self, **kwargs):
        self.calls.append(kwargs)
        if self.error:
            raise self.error
        return SimpleNamespace(
            choices=[SimpleNamespace(message=SimpleNamespace(content=self.content))]
        )


class FakeOpenAIClient:
    def __init__(self, completions: FakeCompletions):
        self.chat = SimpleNamespace(completions=completions)


def build_classifier(monkeypatch, completions: FakeCompletions, model: str = "cheap/model"):
    client = FakeOpenAIClient(completions)
    captured_kwargs = {}

    def fake_openai(**kwargs):
        captured_kwargs.update(kwargs)
        return client

    monkeypatch.setattr(classifier.config, "classification_model", model)
    monkeypatch.setattr(classifier.openai, "OpenAI", fake_openai)

    service = classifier.OpenRouterDocumentClassifier(api_key="test-key")
    return service, captured_kwargs


def test_classifier_uses_configured_model(monkeypatch) -> None:
    completions = FakeCompletions('{"doc_type":"other","language":"en"}')

    service, captured_kwargs = build_classifier(monkeypatch, completions, "google/test-cheap")

    assert service.model_name == "google/test-cheap"
    assert captured_kwargs["api_key"] == "test-key"
    assert captured_kwargs["base_url"] == "https://openrouter.ai/api/v1"


def test_classifier_rejects_empty_model(monkeypatch) -> None:
    monkeypatch.setattr(classifier.config, "classification_model", "   ")

    with pytest.raises(ValueError, match="CLASSIFICATION_MODEL must not be empty"):
        classifier.OpenRouterDocumentClassifier(api_key="test-key")


@pytest.mark.asyncio
async def test_classifier_extracts_metadata_from_json_response(monkeypatch) -> None:
    completions = FakeCompletions(
        """```json
        {
          "doc_type": "invoice",
          "language": "en",
          "issuer": "Acme Corp",
          "amount": "123.45",
          "currency": "USD",
          "issue_date": null,
          "due_date": "2026-06-01",
          "expiry_date": null,
          "document_number": "INV-123"
        }
        ```"""
    )
    service, _ = build_classifier(monkeypatch, completions)

    metadata = await service.extract_metadata("doc-1", "Invoice # INV-123 total due USD 123.45")

    assert metadata == {
        "doc_type": "invoice",
        "language": "en",
        "issuer": "Acme Corp",
        "amount": "123.45",
        "currency": "USD",
        "due_date": "2026-06-01",
        "document_number": "INV-123",
    }
    assert completions.calls[0]["model"] == "cheap/model"
    assert completions.calls[0]["temperature"] == 0
    assert completions.calls[0]["max_tokens"] == 500


@pytest.mark.asyncio
async def test_classifier_defaults_invalid_doc_type_to_other(monkeypatch) -> None:
    completions = FakeCompletions('{"doc_type":"bank_statement","language":"en"}')
    service, _ = build_classifier(monkeypatch, completions)

    metadata = await service.extract_metadata("doc-1", "Account summary")

    assert metadata == {"doc_type": "other", "language": "en"}


@pytest.mark.asyncio
async def test_classifier_falls_back_when_api_fails(monkeypatch) -> None:
    completions = FakeCompletions(error=RuntimeError("network failed"))
    service, _ = build_classifier(monkeypatch, completions)

    metadata = await service.extract_metadata("doc-1", "مرحبا")

    assert metadata == {"doc_type": "other", "language": "ar"}


@pytest.mark.parametrize(
    ("metadata", "expected"),
    [
        ({"doc_type": "invoice", "issuer": "Acme Corp"}, "Invoices/Acme Corp"),
        ({"doc_type": "receipt", "issuer": "Acme Corp"}, "Invoices/Acme Corp"),
        ({"doc_type": "invoice"}, "Invoices"),
        ({"doc_type": "contract", "issuer": "Globex"}, "Contracts/Globex"),
        ({"doc_type": "warranty", "issuer": "Globex"}, "Warranties/Globex"),
        ({"doc_type": "identity", "issuer": "Passport Office"}, "Identity"),
        ({"doc_type": "other"}, "Documents"),
        ({}, "Documents"),
        # Issuer sanitization: path separators stripped, whitespace collapsed.
        ({"doc_type": "invoice", "issuer": "  Acme / Co\\rp  "}, "Invoices/Acme Co rp"),
        ({"doc_type": "invoice", "issuer": "   "}, "Invoices"),
    ],
)
def test_suggest_folder_path(metadata, expected) -> None:
    assert classifier.suggest_folder_path(metadata) == expected


@pytest.mark.parametrize(
    ("metadata", "expected"),
    [
        ({"doc_type": "invoice", "issuer": "Acme"}, 0.9),
        ({"doc_type": "invoice"}, 0.7),
        ({"doc_type": "other"}, 0.3),
        ({}, 0.3),
        ({"doc_type": "bank_statement"}, 0.3),
    ],
)
def test_suggestion_confidence(metadata, expected) -> None:
    assert classifier.suggestion_confidence(metadata) == expected
