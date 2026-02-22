from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class EvalCase:
    id: str
    question: str
    provider: str
    rag: bool
    top_k: int
    expected_contains: list[str]
    require_citations: bool
    model: str | None = None
    debug: bool = False


@dataclass(frozen=True)
class EvalCaseResult:
    case_id: str
    question: str
    expected_contains: list[str]
    answer: str
    citations: list[str]
    rag_used: bool
    latency_ms: int
    passed: bool
    error: str | None = None


@dataclass(frozen=True)
class EvalRunSummary:
    total_cases: int
    passed_cases: int
    avg_latency_ms: float
