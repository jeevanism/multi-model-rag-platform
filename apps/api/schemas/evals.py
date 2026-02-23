from __future__ import annotations

from pydantic import BaseModel


class EvalRunSummaryItem(BaseModel):
    id: int
    dataset_name: str
    provider: str
    model: str | None
    total_cases: int
    passed_cases: int
    avg_latency_ms: float | None
    created_at: str


class EvalRunCaseItem(BaseModel):
    id: int
    case_id: str
    question: str
    passed: bool
    latency_ms: int
    correctness_score: float | None
    groundedness_score: float | None
    hallucination_score: float | None
    citations: list[str]
    error: str | None


class EvalRunDetail(BaseModel):
    run: EvalRunSummaryItem
    cases: list[EvalRunCaseItem]
