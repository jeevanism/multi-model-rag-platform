from __future__ import annotations

from apps.api.schemas.chat import ChatRequest, ChatResponse
from packages.llm.router import get_provider


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
