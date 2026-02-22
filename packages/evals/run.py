from __future__ import annotations

import argparse
import json
from pathlib import Path

import httpx

from apps.api.db import SessionLocal
from packages.evals.aggregate import summarize_results
from packages.evals.dataset import load_eval_dataset
from packages.evals.judge import score_case
from packages.evals.persistence import create_eval_run, insert_eval_run_cases
from packages.evals.types import EvalCase, EvalCaseResult


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run eval dataset against local /chat API.")
    parser.add_argument("--dataset", default="datasets/eval_set.jsonl")
    parser.add_argument("--api-base-url", default="http://localhost:8000")
    parser.add_argument("--limit", type=int, default=None)
    parser.add_argument("--persist", action="store_true")
    return parser.parse_args()


def run_eval(
    *,
    dataset_path: str,
    api_base_url: str,
    limit: int | None = None,
    persist: bool = False,
) -> dict[str, object]:
    cases = load_eval_dataset(dataset_path)
    if limit is not None:
        cases = cases[:limit]

    results = _execute_cases(cases, api_base_url=api_base_url)
    summary = summarize_results(results)

    payload: dict[str, object] = {
        "dataset": str(dataset_path),
        "api_base_url": api_base_url,
        "summary": {
            "total_cases": summary.total_cases,
            "passed_cases": summary.passed_cases,
            "avg_latency_ms": round(summary.avg_latency_ms, 2),
        },
        "results": [result.__dict__ for result in results],
    }

    if persist:
        db = SessionLocal()
        try:
            provider = cases[0].provider if cases else "mixed"
            model = cases[0].model if cases and cases[0].model else None
            eval_run_id = create_eval_run(
                db,
                dataset_name=Path(dataset_path).name,
                provider=provider,
                model=model,
                api_base_url=api_base_url,
                summary=summary,
            )
            insert_eval_run_cases(db, eval_run_id, results)
            db.commit()
            payload["eval_run_id"] = eval_run_id
        except Exception:
            db.rollback()
            raise
        finally:
            db.close()

    return payload


def _execute_cases(cases: list[EvalCase], *, api_base_url: str) -> list[EvalCaseResult]:
    results: list[EvalCaseResult] = []
    with httpx.Client(base_url=api_base_url, timeout=10.0) as client:
        for case in cases:
            results.append(_execute_case(client, case))
    return results


def _execute_case(client: httpx.Client, case: EvalCase) -> EvalCaseResult:
    request_payload = {
        "message": case.question,
        "provider": case.provider,
        "model": case.model,
        "rag": case.rag,
        "top_k": case.top_k,
        "debug": case.debug,
    }
    try:
        response = client.post("/chat", json=request_payload)
        response.raise_for_status()
        body = response.json()
        answer = str(body.get("answer", ""))
        citations = [str(item) for item in body.get("citations", [])]
        rag_used = bool(body.get("rag_used", False))
        latency_ms = int(body.get("latency_ms", 0) or 0)
        passed = _evaluate_case(case, answer=answer, citations=citations)
        scores = score_case(case, answer=answer, citations=citations, rag_used=rag_used)
        return EvalCaseResult(
            case_id=case.id,
            question=case.question,
            expected_contains=case.expected_contains,
            answer=answer,
            citations=citations,
            rag_used=rag_used,
            latency_ms=latency_ms,
            passed=passed,
            correctness_score=scores.correctness_score,
            groundedness_score=scores.groundedness_score,
            hallucination_score=scores.hallucination_score,
        )
    except (httpx.HTTPError, json.JSONDecodeError, ValueError) as exc:
        return EvalCaseResult(
            case_id=case.id,
            question=case.question,
            expected_contains=case.expected_contains,
            answer="",
            citations=[],
            rag_used=False,
            latency_ms=0,
            passed=False,
            correctness_score=0.0,
            groundedness_score=0.0,
            hallucination_score=1.0,
            error=str(exc),
        )


def _evaluate_case(case: EvalCase, *, answer: str, citations: list[str]) -> bool:
    answer_lower = answer.lower()
    has_expected = all(expected.lower() in answer_lower for expected in case.expected_contains)
    has_citations = (not case.require_citations) or bool(citations)
    return has_expected and has_citations


def main() -> None:
    args = parse_args()
    payload = run_eval(
        dataset_path=args.dataset,
        api_base_url=args.api_base_url,
        limit=args.limit,
        persist=args.persist,
    )
    print(json.dumps(payload["summary"], indent=2))
    if "eval_run_id" in payload:
        print(json.dumps({"eval_run_id": payload["eval_run_id"]}))


if __name__ == "__main__":
    main()
