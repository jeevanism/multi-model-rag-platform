from __future__ import annotations

import os
import time
from typing import Any

from packages.llm.base import LLMProvider
from packages.llm.types import LLMResponse


class OpenAIProvider(LLMProvider):
    def __init__(self, model: str = "gpt-4.1-mini", mode: str = "stub") -> None:
        self.provider_name = "openai"
        self.model = model
        self.mode = mode

    def generate(self, prompt: str) -> LLMResponse:
        if self._use_stub():
            return self._generate_stub(prompt)
        return self._generate_real(prompt)

    def _use_stub(self) -> bool:
        if self.mode == "stub":
            return True
        if self.mode == "auto":
            return not bool(os.getenv("OPENAI_API_KEY"))
        if self.mode == "real":
            return False
        raise ValueError(f"Unsupported LLM_PROVIDER_MODE '{self.mode}'. Use stub, auto, or real.")

    def _generate_stub(self, prompt: str) -> LLMResponse:
        start = time.perf_counter()
        answer = f"[stub:{self.provider_name}] {prompt}"
        latency_ms = int((time.perf_counter() - start) * 1000)
        return LLMResponse(
            answer=answer,
            provider=self.provider_name,
            model=self.model,
            latency_ms=latency_ms,
            raw={"stub": True},
        )

    def _generate_real(self, prompt: str) -> LLMResponse:
        api_key = os.getenv("OPENAI_API_KEY")
        if not api_key:
            raise RuntimeError("OPENAI_API_KEY is required when LLM_PROVIDER_MODE=real.")

        try:
            from openai import OpenAI  # type: ignore[import-not-found]
        except ImportError as exc:
            raise RuntimeError(
                "openai package is required for real OpenAI calls. "
                'Install with: uv pip install "openai>=1.0.0"'
            ) from exc

        start = time.perf_counter()
        client = OpenAI(api_key=api_key)
        response = client.responses.create(model=self.model, input=prompt)
        latency_ms = int((time.perf_counter() - start) * 1000)

        answer = self._extract_text(response)
        usage = getattr(response, "usage", None)
        tokens_in = self._maybe_int(getattr(usage, "input_tokens", None))
        tokens_out = self._maybe_int(getattr(usage, "output_tokens", None))

        return LLMResponse(
            answer=answer,
            provider=self.provider_name,
            model=self.model,
            latency_ms=latency_ms,
            tokens_in=tokens_in,
            tokens_out=tokens_out,
            raw={"stub": False},
        )

    @staticmethod
    def _extract_text(response: Any) -> str:
        output_text = getattr(response, "output_text", None)
        if isinstance(output_text, str) and output_text:
            return output_text
        return str(response)

    @staticmethod
    def _maybe_int(value: Any) -> int | None:
        return value if isinstance(value, int) else None
