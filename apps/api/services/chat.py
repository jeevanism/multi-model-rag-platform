from __future__ import annotations

import json
from collections.abc import Iterator
from typing import Mapping

from apps.api.schemas.chat import ChatRequest, ChatResponse
from packages.llm.router import get_provider
from packages.llm.types import LLMResponse


def generate_chat_response(request: ChatRequest) -> ChatResponse:
    provider = get_provider(request.provider, model=request.model)
    result = provider.generate(request.message)
    return ChatResponse(
        answer=result.answer,
        provider=result.provider,
        model=result.model,
        latency_ms=result.latency_ms,
        tokens_in=result.tokens_in,
        tokens_out=result.tokens_out,
        cost_usd=result.cost_usd,
    )


def stream_chat_events(request: ChatRequest) -> Iterator[str]:
    provider = get_provider(request.provider, model=request.model)
    result = provider.generate(request.message)
    return _event_iterator(result)


def _event_iterator(result: LLMResponse) -> Iterator[str]:
    start_payload = {"provider": result.provider, "model": result.model}
    yield _sse_event("start", start_payload)

    for token in result.answer.split():
        yield _sse_event("token", {"text": token})

    end_payload = {
        "answer": result.answer,
        "latency_ms": result.latency_ms,
        "tokens_in": result.tokens_in,
        "tokens_out": result.tokens_out,
        "cost_usd": result.cost_usd,
    }
    yield _sse_event("end", end_payload)


def _sse_event(event: str, data: Mapping[str, object | None]) -> str:
    return f"event: {event}\ndata: {json.dumps(data)}\n\n"
