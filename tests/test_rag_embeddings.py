from __future__ import annotations

import pytest

from packages.rag.embeddings import embed_text


def test_embed_text_defaults_to_stub() -> None:
    result = embed_text("hello")

    assert result.provider == "stub"
    assert result.model == "stub-embedding-v1"
    assert result.raw == {"stub": True}
    assert len(result.vector) == 8


def test_embed_text_auto_mode_falls_back_to_stub_without_key(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.setenv("EMBEDDING_PROVIDER_MODE", "auto")
    monkeypatch.setenv("EMBEDDING_PROVIDER", "gemini")
    monkeypatch.delenv("GEMINI_API_KEY", raising=False)

    result = embed_text("hello")

    assert result.provider == "stub"
    assert result.raw == {"stub": True}


def test_embed_text_real_mode_requires_key(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("EMBEDDING_PROVIDER_MODE", "real")
    monkeypatch.setenv("EMBEDDING_PROVIDER", "openai")
    monkeypatch.delenv("OPENAI_API_KEY", raising=False)

    with pytest.raises(RuntimeError, match="OPENAI_API_KEY"):
        embed_text("hello")
