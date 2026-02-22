from __future__ import annotations

from packages.llm.base import LLMProvider
from packages.llm.providers.gemini import GeminiProvider
from packages.llm.providers.openai import OpenAIProvider


class UnsupportedProviderError(ValueError):
    pass


def get_provider(provider: str, model: str | None = None) -> LLMProvider:
    normalized = provider.strip().lower()

    if normalized == "gemini":
        return GeminiProvider(model=model or "gemini-2.5-flash")
    if normalized == "openai":
        return OpenAIProvider(model=model or "gpt-4.1-mini")

    raise UnsupportedProviderError(
        f"Unsupported provider '{provider}'. Supported values: gemini, openai."
    )
