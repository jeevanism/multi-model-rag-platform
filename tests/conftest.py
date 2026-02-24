from __future__ import annotations

import pytest


@pytest.fixture(autouse=True)
def default_stub_modes(monkeypatch: pytest.MonkeyPatch) -> None:
    """Keep tests deterministic unless a test explicitly overrides modes."""
    monkeypatch.setenv("LLM_PROVIDER_MODE", "stub")
    monkeypatch.setenv("EMBEDDING_PROVIDER_MODE", "stub")
