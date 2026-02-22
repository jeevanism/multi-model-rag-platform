from __future__ import annotations

from fastapi.testclient import TestClient

from apps.api.main import app


def test_chat_returns_required_fields_for_stub_provider() -> None:
    client = TestClient(app)

    response = client.post(
        "/chat",
        json={"message": "hello", "provider": "gemini"},
    )

    assert response.status_code == 200
    data = response.json()
    assert data["answer"] == "[stub:gemini] hello"
    assert data["provider"] == "gemini"
    assert data["model"] == "gemini-2.5-flash"
    assert isinstance(data["latency_ms"], int)
    assert "tokens_in" in data
    assert "tokens_out" in data
    assert "cost_usd" in data


def test_chat_returns_400_for_unsupported_provider() -> None:
    client = TestClient(app)

    response = client.post(
        "/chat",
        json={"message": "hello", "provider": "anthropic"},
    )

    assert response.status_code == 400
    assert "Unsupported provider" in response.json()["detail"]


def test_chat_validates_missing_message() -> None:
    client = TestClient(app)

    response = client.post("/chat", json={"provider": "gemini"})

    assert response.status_code == 422
