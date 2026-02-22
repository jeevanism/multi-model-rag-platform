from __future__ import annotations

from pathlib import Path

import pytest

from packages.evals.aggregate import summarize_results
from packages.evals.dataset import load_eval_dataset
from packages.evals.types import EvalCaseResult


def test_load_eval_dataset_reads_cases(tmp_path: Path) -> None:
    dataset = tmp_path / "eval.jsonl"
    dataset.write_text(
        "\n".join(
            [
                '{"id":"c1","question":"q1","provider":"gemini","expected_contains":["x"]}',
                '{"id":"c2","question":"q2","provider":"openai","expected_contains":["y"],"rag":true}',
            ]
        ),
        encoding="utf-8",
    )

    cases = load_eval_dataset(dataset)

    assert len(cases) == 2
    assert cases[0].id == "c1"
    assert cases[1].rag is True


def test_load_eval_dataset_rejects_missing_fields(tmp_path: Path) -> None:
    dataset = tmp_path / "eval.jsonl"
    dataset.write_text(
        '{"id":"c1","provider":"gemini","expected_contains":["x"]}\n',
        encoding="utf-8",
    )

    with pytest.raises(ValueError, match="Missing required field"):
        load_eval_dataset(dataset)


def test_summarize_results_computes_pass_rate_and_latency() -> None:
    results = [
        EvalCaseResult(
            case_id="c1",
            question="q1",
            expected_contains=["x"],
            answer="x",
            citations=[],
            rag_used=False,
            latency_ms=100,
            passed=True,
        ),
        EvalCaseResult(
            case_id="c2",
            question="q2",
            expected_contains=["y"],
            answer="",
            citations=[],
            rag_used=False,
            latency_ms=200,
            passed=False,
            error="bad",
        ),
    ]

    summary = summarize_results(results)

    assert summary.total_cases == 2
    assert summary.passed_cases == 1
    assert summary.avg_latency_ms == 150.0
