from __future__ import annotations

import json
from pathlib import Path

import pytest

from packages.evals.gate import GateThresholds, evaluate_gate, gate_from_files


def test_evaluate_gate_passes_for_good_summary() -> None:
    decision = evaluate_gate(
        {
            "total_cases": 3,
            "passed_cases": 3,
            "avg_latency_ms": 10.0,
            "correctness_avg": 1.0,
            "groundedness_avg": 1.0,
            "hallucination_avg": 0.0,
        },
        baseline_summary={
            "total_cases": 3,
            "passed_cases": 3,
            "correctness_avg": 1.0,
            "groundedness_avg": 1.0,
            "hallucination_avg": 0.0,
        },
        thresholds=GateThresholds(max_avg_latency_ms=50.0),
    )

    assert decision.passed is True
    assert decision.failures == []


def test_evaluate_gate_fails_on_regression() -> None:
    decision = evaluate_gate(
        {
            "total_cases": 3,
            "passed_cases": 2,
            "avg_latency_ms": 5.0,
            "correctness_avg": 0.7,
            "groundedness_avg": 0.8,
            "hallucination_avg": 0.4,
        },
        baseline_summary={
            "total_cases": 3,
            "passed_cases": 3,
            "correctness_avg": 1.0,
            "groundedness_avg": 1.0,
            "hallucination_avg": 0.0,
        },
        thresholds=GateThresholds(),
    )

    assert decision.passed is False
    assert any("pass_rate" in failure for failure in decision.failures)
    assert any("correctness_avg" in failure for failure in decision.failures)
    assert any("groundedness_avg" in failure for failure in decision.failures)
    assert any("hallucination_avg" in failure for failure in decision.failures)


def test_gate_from_files_reads_payloads(tmp_path: Path) -> None:
    current = tmp_path / "current.json"
    baseline = tmp_path / "baseline.json"

    current.write_text(
        json.dumps(
            {
                "summary": {
                    "total_cases": 1,
                    "passed_cases": 1,
                    "avg_latency_ms": 0.0,
                    "correctness_avg": 1.0,
                    "groundedness_avg": 1.0,
                    "hallucination_avg": 0.0,
                }
            }
        ),
        encoding="utf-8",
    )
    baseline.write_text(current.read_text(encoding="utf-8"), encoding="utf-8")

    decision = gate_from_files(current, baseline_path=baseline)

    assert decision.passed is True


def test_gate_from_files_rejects_missing_summary(tmp_path: Path) -> None:
    current = tmp_path / "current.json"
    current.write_text("{}", encoding="utf-8")

    with pytest.raises(ValueError, match="missing 'summary'"):
        gate_from_files(current)
