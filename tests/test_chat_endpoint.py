from __future__ import annotations

from typing import Any

import pytest
from fastapi.testclient import TestClient

from apps.api.main import app
from packages.rag.retrieval import RetrievedChunk


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
    assert data["rag_used"] is False
    assert data["citations"] == []


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


def test_chat_rag_returns_citations_and_debug_chunks(monkeypatch: pytest.MonkeyPatch) -> None:
    client = TestClient(app)

    def fake_retrieve_chunks(db: Any, query: str, top_k: int = 3) -> list[RetrievedChunk]:
        assert query == "what is in the doc?"
        assert top_k == 2
        return [
            RetrievedChunk(
                document_id=1,
                chunk_id=10,
                chunk_index=0,
                title="Test Doc",
                content="Important context.",
                score=0.123,
            )
        ]

    monkeypatch.setattr("apps.api.services.chat.retrieve_chunks", fake_retrieve_chunks)

    response = client.post(
        "/chat",
        json={
            "message": "what is in the doc?",
            "provider": "gemini",
            "rag": True,
            "top_k": 2,
            "debug": True,
        },
    )

    assert response.status_code == 200
    data = response.json()
    assert data["rag_used"] is True
    assert data["citations"] == ["[source:Test Doc#chunk=0]"]
    assert data["retrieved_chunks"][0]["title"] == "Test Doc"
    assert "[source:Test Doc#chunk=0]" in data["answer"]


def test_chat_stream_returns_sse_events() -> None:
    client = TestClient(app)

    response = client.post(
        "/chat/stream",
        json={"message": "hello world", "provider": "gemini"},
    )

    assert response.status_code == 200
    assert response.headers["content-type"].startswith("text/event-stream")
    assert "event: start" in response.text
    assert "event: token" in response.text
    assert '"text": "hello"' in response.text
    assert '"text": "world"' in response.text
    assert "event: end" in response.text


def test_chat_stream_returns_400_for_unsupported_provider() -> None:
    client = TestClient(app)

    response = client.post(
        "/chat/stream",
        json={"message": "hello", "provider": "anthropic"},
    )

    assert response.status_code == 400
