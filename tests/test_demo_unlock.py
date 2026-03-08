from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from apps.api.main import app


def test_demo_status_reports_locked_by_default() -> None:
    client = TestClient(app)

    response = client.get("/auth/demo-status")

    assert response.status_code == 200
    data = response.json()
    assert data["unlocked"] is False
    assert "unlock_enabled" in data


def test_demo_unlock_returns_401_for_invalid_password(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    client = TestClient(app)
    monkeypatch.setattr("apps.api.main.validate_demo_password", lambda password: False)

    response = client.post("/auth/demo-unlock", json={"password": "wrong"})

    assert response.status_code == 401


def test_demo_unlock_sets_cookie_on_success(monkeypatch: pytest.MonkeyPatch) -> None:
    client = TestClient(app)
    monkeypatch.setattr("apps.api.main.validate_demo_password", lambda password: True)

    response = client.post("/auth/demo-unlock", json={"password": "ok"})

    assert response.status_code == 200
    assert "set-cookie" in response.headers


def test_chat_forces_stub_mode_for_public_requests_even_if_global_mode_is_real(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    client = TestClient(app)
    monkeypatch.setenv("LLM_PROVIDER_MODE", "real")
    monkeypatch.delenv("GEMINI_API_KEY", raising=False)
    monkeypatch.setattr("apps.api.main.is_demo_unlocked", lambda request: False)

    response = client.post("/chat", json={"message": "hello", "provider": "gemini"})

    assert response.status_code == 200
    assert response.json()["answer"] == "[stub:gemini] hello"
