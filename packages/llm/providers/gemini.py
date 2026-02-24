from __future__ import annotations

import os
import time
from typing import Any

from packages.llm.base import LLMProvider
from packages.llm.types import LLMResponse


class GeminiProvider(LLMProvider):
    def __init__(self, model: str = "gemini-2.5-flash", mode: str = "stub") -> None:
        self.provider_name = "gemini"
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
            return not bool(os.getenv("GEMINI_API_KEY"))
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
        api_key = os.getenv("GEMINI_API_KEY")
        if not api_key:
            raise RuntimeError("GEMINI_API_KEY is required when LLM_PROVIDER_MODE=real.")

        try:
            from google import genai  # type: ignore[import-not-found,attr-defined]
        except ImportError as exc:
            raise RuntimeError(
                "google-genai package is required for real Gemini calls. "
                'Install with: uv pip install "google-genai>=1.0.0"'
            ) from exc

        start = time.perf_counter()
        client = genai.Client(api_key=api_key)
        response = client.models.generate_content(model=self.model, contents=prompt)
        latency_ms = int((time.perf_counter() - start) * 1000)

        answer = self._extract_text(response)
        usage = getattr(response, "usage_metadata", None)
        tokens_in = self._maybe_int(getattr(usage, "prompt_token_count", None))
        tokens_out = self._maybe_int(getattr(usage, "candidates_token_count", None))

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
        text = getattr(response, "text", None)
        if isinstance(text, str) and text:
            return text
        return str(response)

    @staticmethod
    def _maybe_int(value: Any) -> int | None:
        return value if isinstance(value, int) else None
