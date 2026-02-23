from __future__ import annotations

from typing import Any, cast

from sqlalchemy import text
from sqlalchemy.orm import Session

from apps.api.schemas.evals import EvalRunCaseItem, EvalRunDetail, EvalRunSummaryItem


def list_eval_runs(db: Session, limit: int = 20) -> list[EvalRunSummaryItem]:
    rows = db.execute(
        text(
            """
            SELECT
                id,
                dataset_name,
                provider,
                model,
                total_cases,
                passed_cases,
                avg_latency_ms,
                created_at
            FROM eval_run
            ORDER BY id DESC
            LIMIT :limit
            """
        ),
        {"limit": limit},
    )
    return [_to_eval_run_summary_item(row) for row in rows.mappings()]


def get_eval_run_detail(db: Session, eval_run_id: int) -> EvalRunDetail | None:
    run_row = (
        db.execute(
            text(
                """
            SELECT
                id,
                dataset_name,
                provider,
                model,
                total_cases,
                passed_cases,
                avg_latency_ms,
                created_at
            FROM eval_run
            WHERE id = :eval_run_id
            """
            ),
            {"eval_run_id": eval_run_id},
        )
        .mappings()
        .one_or_none()
    )

    if run_row is None:
        return None

    case_rows = db.execute(
        text(
            """
            SELECT
                id, case_id, question, passed, latency_ms,
                correctness_score, groundedness_score, hallucination_score,
                citations, error
            FROM eval_run_case
            WHERE eval_run_id = :eval_run_id
            ORDER BY id ASC
            """
        ),
        {"eval_run_id": eval_run_id},
    )

    return EvalRunDetail(
        run=_to_eval_run_summary_item(run_row),
        cases=[_to_eval_run_case_item(row) for row in case_rows.mappings()],
    )


def _to_eval_run_summary_item(row: Any) -> EvalRunSummaryItem:
    return EvalRunSummaryItem(
        id=cast(int, row["id"]),
        dataset_name=cast(str, row["dataset_name"]),
        provider=cast(str, row["provider"]),
        model=cast(str | None, row["model"]),
        total_cases=cast(int, row["total_cases"]),
        passed_cases=cast(int, row["passed_cases"]),
        avg_latency_ms=float(row["avg_latency_ms"]) if row["avg_latency_ms"] is not None else None,
        created_at=str(row["created_at"]),
    )


def _to_eval_run_case_item(row: Any) -> EvalRunCaseItem:
    citations = row["citations"] or []
    return EvalRunCaseItem(
        id=cast(int, row["id"]),
        case_id=cast(str, row["case_id"]),
        question=cast(str, row["question"]),
        passed=cast(bool, row["passed"]),
        latency_ms=cast(int, row["latency_ms"]),
        correctness_score=(
            float(row["correctness_score"]) if row["correctness_score"] is not None else None
        ),
        groundedness_score=float(row["groundedness_score"])
        if row["groundedness_score"] is not None
        else None,
        hallucination_score=float(row["hallucination_score"])
        if row["hallucination_score"] is not None
        else None,
        citations=[str(citation) for citation in citations],
        error=cast(str | None, row["error"]),
    )
