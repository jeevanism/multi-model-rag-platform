from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from packages.evals.types import EvalCase


def load_eval_dataset(path: str | Path) -> list[EvalCase]:
    dataset_path = Path(path)
    cases: list[EvalCase] = []

    lines = dataset_path.read_text(encoding="utf-8").splitlines()
    for line_number, line in enumerate(lines, start=1):
        line = line.strip()
        if not line:
            continue
        try:
            payload = json.loads(line)
        except json.JSONDecodeError as exc:
            raise ValueError(f"Invalid JSON on line {line_number}: {exc}") from exc

        cases.append(_parse_case(payload, line_number))

    if not cases:
        raise ValueError("Eval dataset is empty")
    return cases


def _parse_case(payload: dict[str, Any], line_number: int) -> EvalCase:
    required_fields = ["id", "question", "provider", "expected_contains"]
    missing = [field for field in required_fields if field not in payload]
    if missing:
        raise ValueError(f"Missing required field(s) {missing} on line {line_number}")

    expected_contains = payload["expected_contains"]
    if not isinstance(expected_contains, list) or not all(
        isinstance(item, str) for item in expected_contains
    ):
        raise ValueError(f"'expected_contains' must be a list[str] on line {line_number}")

    return EvalCase(
        id=str(payload["id"]),
        question=str(payload["question"]),
        provider=str(payload["provider"]),
        rag=bool(payload.get("rag", False)),
        top_k=int(payload.get("top_k", 3)),
        expected_contains=expected_contains,
        require_citations=bool(payload.get("require_citations", False)),
        model=str(payload["model"]) if payload.get("model") is not None else None,
        debug=bool(payload.get("debug", False)),
    )
