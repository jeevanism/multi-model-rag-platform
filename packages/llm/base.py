from __future__ import annotations

from abc import ABC, abstractmethod

from packages.llm.types import LLMResponse


class LLMProvider(ABC):
    provider_name: str
    model: str

    @abstractmethod
    def generate(self, prompt: str) -> LLMResponse:
        """Generate a response from the provider."""

