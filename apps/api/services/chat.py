from __future__ import annotations

import json
from collections.abc import Iterator
from typing import Mapping

from apps.api.schemas.chat import ChatRequest, ChatResponse, RetrievedChunkPreview
from sqlalchemy.orm import Session

from packages.llm.router import get_provider
from packages.llm.types import LLMResponse
from packages.rag.citations import format_citations
from packages.rag.prompting import build_grounded_prompt
from packages.rag.retrieval import retrieve_chunks
from packages.observability.tracing import trace_span


def generate_chat_response(request: ChatRequest, db: Session | None = None) -> ChatResponse:
    with trace_span("chat.generate", provider=request.provider, rag=request.rag):
        provider = get_provider(request.provider, model=request.model)
        citations: list[str] = []
        retrieved_chunks_payload = None
        rag_used = False
        prompt = request.message

        if request.rag:
            if db is None:
                raise ValueError("Database session is required when rag=true")
            with trace_span("rag.retrieve", top_k=request.top_k):
                retrieved_chunks = retrieve_chunks(db, request.message, top_k=request.top_k)
            citations = format_citations(retrieved_chunks)
            prompt = build_grounded_prompt(request.message, retrieved_chunks)
            rag_used = True
            if request.debug:
                retrieved_chunks_payload = [
                    RetrievedChunkPreview(
                        document_id=chunk.document_id,
                        chunk_id=chunk.chunk_id,
                        chunk_index=chunk.chunk_index,
                        title=chunk.title,
                        content=chunk.content,
                        score=chunk.score,
                    )
                    for chunk in retrieved_chunks
                ]

        with trace_span("llm.generate", provider=provider.provider_name, model=provider.model):
            result = provider.generate(prompt)
        answer = result.answer
        if citations:
            answer = f"{answer}\n\nCitations: {' '.join(citations)}"

        return ChatResponse(
            answer=answer,
            provider=result.provider,
            model=result.model,
            latency_ms=result.latency_ms,
            tokens_in=result.tokens_in,
            tokens_out=result.tokens_out,
            cost_usd=result.cost_usd,
            citations=citations,
            rag_used=rag_used,
            retrieved_chunks=retrieved_chunks_payload,
        )


def stream_chat_events(request: ChatRequest) -> Iterator[str]:
    with trace_span("chat.stream.generate", provider=request.provider):
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
