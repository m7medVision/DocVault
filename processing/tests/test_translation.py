from types import SimpleNamespace

import pytest

from processing import translation


def test_translation_service_uses_configured_model(monkeypatch) -> None:
    monkeypatch.setattr(translation.config, "translation_model", "mistralai/mistral-large")

    service = translation.OpenRouterTranslationService(api_key="test-key")

    assert service.model_name == "mistralai/mistral-large"


def test_translation_service_rejects_empty_model(monkeypatch) -> None:
    monkeypatch.setattr(translation.config, "translation_model", "   ")

    try:
        translation.OpenRouterTranslationService(api_key="test-key")
    except ValueError as exc:
        assert str(exc) == "TRANSLATION_MODEL must not be empty"
    else:
        raise AssertionError("expected empty translation model to raise ValueError")


@pytest.mark.asyncio
async def test_translate_document_preserves_page_order(monkeypatch) -> None:
    """Pages translate concurrently but the result map keeps page numbers."""
    monkeypatch.setattr(translation.config, "translation_model", "mistralai/mistral-large")

    translator = translation.DocumentTranslator()

    async def fake_translate_text(text, source_language, target_language="en"):
        return translation.TranslationResult(
            original_text=text,
            translated_text=f"EN:{text}",
            source_language=source_language,
            target_language=target_language,
            is_translation=True,
        )

    monkeypatch.setattr(
        translator.translation_service, "translate_text", fake_translate_text
    )

    pages = [
        SimpleNamespace(page_number=1, text="alpha"),
        SimpleNamespace(page_number=2, text="beta"),
        SimpleNamespace(page_number=3, text="gamma"),
    ]

    result = await translator.translate_document("doc-1", pages, "ar")

    assert result["is_translation"] is True
    assert result["pages"][1]["translated_text"] == "EN:alpha"
    assert result["pages"][2]["translated_text"] == "EN:beta"
    assert result["pages"][3]["translated_text"] == "EN:gamma"
    assert list(result["pages"].keys()) == [1, 2, 3]
