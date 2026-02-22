from __future__ import annotations

from packages.evals.judge import score_case
from packages.evals.types import EvalCase


def test_score_case_rewards_expected_content_and_citations_for_rag() -> None:
    case = EvalCase(
        id="c1",
        question="What is the capital of France?",
        provider="gemini",
        rag=True,
        top_k=2,
        expected_contains=["France", "capital"],
        require_citations=True,
    )

    scores = score_case(
        case,
        answer="France is the capital. [source:RAG Doc#chunk=0]",
        citations=["[source:RAG Doc#chunk=0]"],
        rag_used=True,
    )

    assert scores.correctness_score == 1.0
    assert scores.groundedness_score == 1.0
    assert scores.hallucination_score == 0.0


def test_score_case_penalizes_missing_citations_for_rag_case() -> None:
    case = EvalCase(
        id="c2",
        question="q",
        provider="gemini",
        rag=True,
        top_k=2,
        expected_contains=["France"],
        require_citations=True,
    )

    scores = score_case(case, answer="France", citations=[], rag_used=True)

    assert scores.correctness_score == 1.0
    assert scores.groundedness_score == 0.0
    assert scores.hallucination_score == 1.0


def test_score_case_non_rag_defaults_groundedness_to_one() -> None:
    case = EvalCase(
        id="c3",
        question="hello",
        provider="openai",
        rag=False,
        top_k=3,
        expected_contains=["hello"],
        require_citations=False,
    )

    scores = score_case(case, answer="hello there", citations=[], rag_used=False)

    assert scores.correctness_score == 1.0
    assert scores.groundedness_score == 1.0
    assert scores.hallucination_score == 0.0
