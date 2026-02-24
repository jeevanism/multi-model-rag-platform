from __future__ import annotations

import os

from packages.llm.base import LLMProvider
from packages.llm.providers.gemini import GeminiProvider
from packages.llm.providers.openai import OpenAIProvider


class UnsupportedProviderError(ValueError):
    pass


def _provider_mode() -> str:
    return os.getenv("LLM_PROVIDER_MODE", "stub").strip().lower()


def get_provider(provider: str, model: str | None = None) -> LLMProvider:
    normalized = provider.strip().lower()
    mode = _provider_mode()

    if normalized == "gemini":
        return GeminiProvider(model=model or "gemini-2.5-flash", mode=mode)
    if normalized == "openai":
        return OpenAIProvider(model=model or "gpt-4.1-mini", mode=mode)

    raise UnsupportedProviderError(
        f"Unsupported provider '{provider}'. Supported values: gemini, openai."
    )
