from __future__ import annotations

from contextlib import contextmanager
from contextvars import ContextVar
import os
from collections.abc import Iterator

from packages.llm.base import LLMProvider
from packages.llm.providers.gemini import GeminiProvider
from packages.llm.providers.grok import GrokProvider
from packages.llm.providers.openai import OpenAIProvider


class UnsupportedProviderError(ValueError):
    pass


_PROVIDER_MODE_OVERRIDE: ContextVar[str | None] = ContextVar("llm_provider_mode_override", default=None)


def _provider_mode() -> str:
    override = _PROVIDER_MODE_OVERRIDE.get()
    if override is not None:
        return override
    return os.getenv("LLM_PROVIDER_MODE", "stub").strip().lower()


@contextmanager
def override_provider_mode(mode: str | None) -> Iterator[None]:
    token = _PROVIDER_MODE_OVERRIDE.set(mode.strip().lower() if mode else None)
    try:
        yield
    finally:
        _PROVIDER_MODE_OVERRIDE.reset(token)


def get_provider(provider: str, model: str | None = None) -> LLMProvider:
    normalized = provider.strip().lower()
    mode = _provider_mode()

    if normalized == "gemini":
        return GeminiProvider(model=model or "gemini-2.5-flash", mode=mode)
    if normalized == "openai":
        return OpenAIProvider(model=model or "gpt-4.1-mini", mode=mode)
    if normalized == "grok":
        return GrokProvider(model=model or "grok-3-mini", mode=mode)

    raise UnsupportedProviderError(
        f"Unsupported provider '{provider}'. Supported values: gemini, openai, grok."
    )
