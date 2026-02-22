from __future__ import annotations

from typing import Any

import pytest
from fastapi.testclient import TestClient
from sqlalchemy.orm import Session

from apps.api.main import app
from apps.api.schemas.ingest import IngestTextRequest


def test_ingest_text_returns_summary(monkeypatch: pytest.MonkeyPatch) -> None:
    client = TestClient(app)

    def fake_ingest_text_document(db: Session, request: IngestTextRequest) -> dict[str, Any]:
        return {
            "document_id": 1,
            "chunk_count": 2,
            "embedding_count": 2,
            "embedding_provider": "stub",
            "embedding_model": "stub-embedding-v1",
        }

    monkeypatch.setattr("apps.api.main.ingest_text_document", fake_ingest_text_document)

    response = client.post(
        "/ingest/text",
        json={"title": "Doc", "content": "hello world"},
    )

    assert response.status_code == 200
    assert response.json()["chunk_count"] == 2
    assert response.json()["embedding_count"] == 2


def test_ingest_text_validates_required_fields() -> None:
    client = TestClient(app)

    response = client.post("/ingest/text", json={"title": "Doc"})

    assert response.status_code == 422
