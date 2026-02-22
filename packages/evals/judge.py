from __future__ import annotations

from dataclasses import dataclass

from packages.evals.types import EvalCase


@dataclass(frozen=True)
class JudgeScores:
    correctness_score: float
    groundedness_score: float
    hallucination_score: float


def score_case(case: EvalCase, *, answer: str, citations: list[str], rag_used: bool) -> JudgeScores:
    answer_lower = answer.lower()
    expected_hits = sum(1 for item in case.expected_contains if item.lower() in answer_lower)
    expected_total = max(1, len(case.expected_contains))
    correctness = expected_hits / expected_total

    if not case.rag:
        groundedness = 1.0
    else:
        groundedness = 1.0 if (rag_used and (not case.require_citations or citations)) else 0.0

    # Heuristic v1: if expected content is present and (when required)
    # citations exist, hallucination is low.
    hallucination = 1.0 - min(correctness, groundedness)

    return JudgeScores(
        correctness_score=round(correctness, 4),
        groundedness_score=round(groundedness, 4),
        hallucination_score=round(hallucination, 4),
    )
