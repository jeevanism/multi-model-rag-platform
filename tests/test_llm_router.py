from __future__ import annotations

import pytest

from packages.llm.router import UnsupportedProviderError, get_provider


@pytest.mark.parametrize("provider_name", ["gemini", "openai"])
def test_get_provider_returns_provider_with_unified_response(provider_name: str) -> None:
    provider = get_provider(provider_name)

    response = provider.generate("hello")

    assert response.provider == provider_name
    assert response.answer == f"[stub:{provider_name}] hello"
    assert isinstance(response.latency_ms, int)
    assert response.latency_ms >= 0
    assert response.raw == {"stub": True}


def test_get_provider_raises_for_unsupported_provider() -> None:
    with pytest.raises(UnsupportedProviderError, match="Unsupported provider"):
        get_provider("anthropic")


def test_get_provider_allows_model_override() -> None:
    provider = get_provider("gemini", model="gemini-test-model")

    response = provider.generate("hi")

    assert response.model == "gemini-test-model"
