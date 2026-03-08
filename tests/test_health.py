from fastapi.testclient import TestClient

from apps.api.main import app


def test_root_returns_200_and_expected_payload() -> None:
    client = TestClient(app)

    response = client.get("/")

    assert response.status_code == 200
    data = response.json()
    assert data["service"] == "Multi-Model RAG API"
    assert data["status"] == "ok"
    assert "/health" in data["endpoints"]
    assert isinstance(data["request_id"], str)


def test_health_returns_200_and_expected_payload() -> None:
    client = TestClient(app)

    response = client.get("/health")

    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ok"
    assert data["service"] == "api"
    assert isinstance(data["request_id"], str)
    assert data["request_id"]
