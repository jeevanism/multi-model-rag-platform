from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from apps.api.main import app


def test_evals_runs_endpoint_returns_list(monkeypatch: pytest.MonkeyPatch) -> None:
    client = TestClient(app)

    monkeypatch.setattr(
        "apps.api.main.list_eval_runs",
        lambda db: [
            {
                "id": 1,
                "dataset_name": "eval_set.jsonl",
                "provider": "gemini",
                "model": None,
                "total_cases": 3,
                "passed_cases": 3,
                "avg_latency_ms": 0.0,
                "created_at": "2026-02-23T00:00:00+00:00",
            }
        ],  # type: ignore[no-any-return]
    )

    response = client.get("/evals/runs")

    assert response.status_code == 200
    assert response.json()[0]["id"] == 1


def test_eval_run_detail_endpoint_returns_404_when_missing(monkeypatch: pytest.MonkeyPatch) -> None:
    client = TestClient(app)
    monkeypatch.setattr(
        "apps.api.main.get_eval_run_detail",
        lambda eval_run_id, db: None,  # type: ignore[no-any-return]
    )

    response = client.get("/evals/runs/999")

    assert response.status_code == 404
