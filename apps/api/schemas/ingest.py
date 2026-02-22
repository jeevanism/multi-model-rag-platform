from __future__ import annotations

from pydantic import BaseModel, Field


class IngestTextRequest(BaseModel):
    title: str = Field(min_length=1)
    content: str = Field(min_length=1)
    chunk_size: int = Field(default=400, gt=0, le=4000)
    chunk_overlap: int = Field(default=40, ge=0, le=1000)


class IngestTextResponse(BaseModel):
    document_id: int
    chunk_count: int
    embedding_count: int
    embedding_provider: str
    embedding_model: str
