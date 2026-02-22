from __future__ import annotations

import argparse

from packages.llm.router import get_provider


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Local LLM provider smoke test (stub providers).")
    parser.add_argument("--provider", choices=["gemini", "openai"], required=True)
    parser.add_argument("--prompt", required=True)
    parser.add_argument("--model", default=None)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    provider = get_provider(args.provider, model=args.model)
    response = provider.generate(args.prompt)
    print(
        {
            "answer": response.answer,
            "provider": response.provider,
            "model": response.model,
            "latency_ms": response.latency_ms,
            "stub": response.raw,
        }
    )


if __name__ == "__main__":
    main()

