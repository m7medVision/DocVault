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
