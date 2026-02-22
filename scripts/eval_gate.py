from __future__ import annotations

import argparse
import json
import sys

from packages.evals.gate import GateThresholds, gate_from_files


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Eval regression gate checker.")
    parser.add_argument("--current", required=True, help="Path to current eval JSON output")
    parser.add_argument("--baseline", default="datasets/eval_baseline.json")
    parser.add_argument("--min-pass-rate", type=float, default=1.0)
    parser.add_argument("--min-correctness-avg", type=float, default=0.9)
    parser.add_argument("--min-groundedness-avg", type=float, default=0.9)
    parser.add_argument("--max-hallucination-avg", type=float, default=0.2)
    parser.add_argument("--max-avg-latency-ms", type=float, default=None)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    thresholds = GateThresholds(
        min_pass_rate=args.min_pass_rate,
        min_correctness_avg=args.min_correctness_avg,
        min_groundedness_avg=args.min_groundedness_avg,
        max_hallucination_avg=args.max_hallucination_avg,
        max_avg_latency_ms=args.max_avg_latency_ms,
    )
    decision = gate_from_files(
        args.current,
        baseline_path=args.baseline if args.baseline else None,
        thresholds=thresholds,
    )
    if decision.passed:
        print(json.dumps({"passed": True, "failures": []}))
        return

    print(json.dumps({"passed": False, "failures": decision.failures}, indent=2))
    sys.exit(1)


if __name__ == "__main__":
    main()
