from __future__ import annotations

import time

from packages.llm.base import LLMProvider
from packages.llm.types import LLMResponse


class GeminiProvider(LLMProvider):
    def __init__(self, model: str = "gemini-2.5-flash") -> None:
        self.provider_name = "gemini"
        self.model = model

    def generate(self, prompt: str) -> LLMResponse:
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

