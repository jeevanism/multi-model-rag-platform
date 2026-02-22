from __future__ import annotations

from pydantic import BaseModel, Field


class ChatRequest(BaseModel):
    message: str = Field(min_length=1)
    provider: str
    model: str | None = None


class ChatResponse(BaseModel):
    answer: str
    provider: str
    model: str
    latency_ms: int
    tokens_in: int | None = None
    tokens_out: int | None = None
    cost_usd: float | None = None
