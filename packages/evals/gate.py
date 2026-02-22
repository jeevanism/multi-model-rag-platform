from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any, cast


@dataclass(frozen=True)
class GateThresholds:
    min_pass_rate: float = 1.0
    min_correctness_avg: float = 0.9
    min_groundedness_avg: float = 0.9
    max_hallucination_avg: float = 0.2
    max_avg_latency_ms: float | None = None


@dataclass(frozen=True)
class GateDecision:
    passed: bool
    failures: list[str]


def load_json(path: str | Path) -> dict[str, Any]:
    payload = json.loads(Path(path).read_text(encoding="utf-8"))
    if not isinstance(payload, dict):
        raise ValueError("JSON payload must be an object")
    return cast(dict[str, Any], payload)


def evaluate_gate(
    current_summary: dict[str, Any],
    *,
    baseline_summary: dict[str, Any] | None = None,
    thresholds: GateThresholds | None = None,
) -> GateDecision:
    t = thresholds or GateThresholds()
    failures: list[str] = []

    total_cases = int(current_summary.get("total_cases", 0))
    passed_cases = int(current_summary.get("passed_cases", 0))
    pass_rate = (passed_cases / total_cases) if total_cases else 0.0

    correctness_avg = float(current_summary.get("correctness_avg", 0.0))
    groundedness_avg = float(current_summary.get("groundedness_avg", 0.0))
    hallucination_avg = float(current_summary.get("hallucination_avg", 1.0))
    avg_latency_ms = float(current_summary.get("avg_latency_ms", 0.0))

    if pass_rate < t.min_pass_rate:
        failures.append(f"pass_rate {pass_rate:.3f} < min_pass_rate {t.min_pass_rate:.3f}")
    if correctness_avg < t.min_correctness_avg:
        failures.append(
            "correctness_avg "
            f"{correctness_avg:.3f} < min_correctness_avg {t.min_correctness_avg:.3f}"
        )
    if groundedness_avg < t.min_groundedness_avg:
        failures.append(
            "groundedness_avg "
            f"{groundedness_avg:.3f} < min_groundedness_avg {t.min_groundedness_avg:.3f}"
        )
    if hallucination_avg > t.max_hallucination_avg:
        failures.append(
            "hallucination_avg "
            f"{hallucination_avg:.3f} > max_hallucination_avg {t.max_hallucination_avg:.3f}"
        )
    if t.max_avg_latency_ms is not None and avg_latency_ms > t.max_avg_latency_ms:
        failures.append(
            f"avg_latency_ms {avg_latency_ms:.2f} > max_avg_latency_ms {t.max_avg_latency_ms:.2f}"
        )

    if baseline_summary is not None:
        baseline_total = int(baseline_summary.get("total_cases", 0))
        baseline_passed = int(baseline_summary.get("passed_cases", 0))
        baseline_pass_rate = (baseline_passed / baseline_total) if baseline_total else 0.0
        if pass_rate + 1e-9 < baseline_pass_rate:
            failures.append(
                f"pass_rate {pass_rate:.3f} regressed below baseline {baseline_pass_rate:.3f}"
            )

        for key in ("correctness_avg", "groundedness_avg"):
            if float(current_summary.get(key, 0.0)) + 1e-9 < float(baseline_summary.get(key, 0.0)):
                failures.append(
                    f"{key} {float(current_summary.get(key, 0.0)):.3f} regressed below baseline "
                    f"{float(baseline_summary.get(key, 0.0)):.3f}"
                )

        if float(current_summary.get("hallucination_avg", 1.0)) - 1e-9 > float(
            baseline_summary.get("hallucination_avg", 1.0)
        ):
            failures.append(
                "hallucination_avg "
                f"{float(current_summary.get('hallucination_avg', 1.0)):.3f} "
                "regressed above baseline "
                f"{float(baseline_summary.get('hallucination_avg', 1.0)):.3f}"
            )

    return GateDecision(passed=not failures, failures=failures)


def gate_from_files(
    current_eval_path: str | Path,
    *,
    baseline_path: str | Path | None = None,
    thresholds: GateThresholds | None = None,
) -> GateDecision:
    current_payload = load_json(current_eval_path)
    current_summary = _extract_summary(current_payload, "current eval")

    baseline_summary: dict[str, Any] | None = None
    if baseline_path is not None:
        baseline_payload = load_json(baseline_path)
        baseline_summary = _extract_summary(baseline_payload, "baseline")

    return evaluate_gate(
        current_summary,
        baseline_summary=baseline_summary,
        thresholds=thresholds,
    )


def _extract_summary(payload: dict[str, Any], label: str) -> dict[str, Any]:
    summary = payload.get("summary")
    if not isinstance(summary, dict):
        raise ValueError(f"{label} missing 'summary' object")
    return summary
