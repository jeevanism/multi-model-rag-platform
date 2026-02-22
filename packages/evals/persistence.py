from __future__ import annotations

from typing import cast

from sqlalchemy import text
from sqlalchemy.orm import Session

from packages.evals.types import EvalCaseResult, EvalRunSummary


def create_eval_run(
    db: Session,
    *,
    dataset_name: str,
    provider: str,
    model: str | None,
    api_base_url: str,
    summary: EvalRunSummary,
) -> int:
    result = db.execute(
        text(
            """
            INSERT INTO eval_run (
                dataset_name, provider, model, api_base_url,
                total_cases, passed_cases, avg_latency_ms
            )
            VALUES (
                :dataset_name, :provider, :model, :api_base_url,
                :total_cases, :passed_cases, :avg_latency_ms
            )
            RETURNING id
            """
        ),
        {
            "dataset_name": dataset_name,
            "provider": provider,
            "model": model,
            "api_base_url": api_base_url,
            "total_cases": summary.total_cases,
            "passed_cases": summary.passed_cases,
            "avg_latency_ms": summary.avg_latency_ms,
        },
    )
    return cast(int, result.scalar_one())


def insert_eval_run_cases(db: Session, eval_run_id: int, results: list[EvalCaseResult]) -> None:
    for result in results:
        db.execute(
            text(
                """
                INSERT INTO eval_run_case (
                    eval_run_id, case_id, question, expected_contains, answer, citations,
                    rag_used, latency_ms, passed, error
                )
                VALUES (
                    :eval_run_id, :case_id, :question, :expected_contains, :answer, :citations,
                    :rag_used, :latency_ms, :passed, :error
                )
                """
            ),
            {
                "eval_run_id": eval_run_id,
                "case_id": result.case_id,
                "question": result.question,
                "expected_contains": result.expected_contains,
                "answer": result.answer,
                "citations": result.citations,
                "rag_used": result.rag_used,
                "latency_ms": result.latency_ms,
                "passed": result.passed,
                "error": result.error,
            },
        )
