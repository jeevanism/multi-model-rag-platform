from __future__ import annotations

from packages.evals.types import EvalCaseResult, EvalRunSummary


def summarize_results(results: list[EvalCaseResult]) -> EvalRunSummary:
    if not results:
        return EvalRunSummary(total_cases=0, passed_cases=0, avg_latency_ms=0.0)

    total_cases = len(results)
    passed_cases = sum(1 for result in results if result.passed)
    avg_latency_ms = sum(result.latency_ms for result in results) / total_cases
    return EvalRunSummary(
        total_cases=total_cases,
        passed_cases=passed_cases,
        avg_latency_ms=avg_latency_ms,
    )
