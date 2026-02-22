from __future__ import annotations

from pydantic import BaseModel, Field


class ChatRequest(BaseModel):
    message: str = Field(min_length=1)
    provider: str
    model: str | None = None
    rag: bool = False
    top_k: int = Field(default=3, gt=0, le=10)
    debug: bool = False


class RetrievedChunkPreview(BaseModel):
    document_id: int
    chunk_id: int
    chunk_index: int
    title: str
    content: str
    score: float


class ChatResponse(BaseModel):
    answer: str
    provider: str
    model: str
    latency_ms: int
    tokens_in: int | None = None
    tokens_out: int | None = None
    cost_usd: float | None = None
    citations: list[str] = Field(default_factory=list)
    rag_used: bool = False
    retrieved_chunks: list[RetrievedChunkPreview] | None = None
