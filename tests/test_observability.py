from __future__ import annotations

import json
import logging
from collections.abc import Generator

import pytest
from fastapi.testclient import TestClient

from apps.api.main import app


@pytest.fixture
def log_capture() -> Generator[list[str], None, None]:
    logger = logging.getLogger("multi_model_rag")
    records: list[str] = []

    class _ListHandler(logging.Handler):
        def emit(self, record: logging.LogRecord) -> None:
            records.append(record.getMessage())

    handler = _ListHandler()
    logger.addHandler(handler)
    try:
        yield records
    finally:
        logger.removeHandler(handler)


def test_request_id_header_is_added_and_propagated_to_health() -> None:
    client = TestClient(app)

    response = client.get("/health")

    assert response.status_code == 200
    assert "x-request-id" in response.headers
    assert response.json()["request_id"] == response.headers["x-request-id"]


def test_request_logs_are_emitted(log_capture: list[str]) -> None:
    client = TestClient(app)

    response = client.post("/chat", json={"message": "hello", "provider": "gemini"})

    assert response.status_code == 200
    events = [json.loads(message)["event"] for message in log_capture if message.startswith("{")]
    assert "request.start" in events
    assert "request.end" in events
    assert "span.start" in events
    assert "span.end" in events
