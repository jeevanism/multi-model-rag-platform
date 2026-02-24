from __future__ import annotations

from typing import cast

from sqlalchemy import text
from sqlalchemy.orm import Session

from apps.api.schemas.ingest import IngestTextRequest, IngestTextResponse
from packages.observability.tracing import trace_span
from packages.rag.chunking import chunk_text
from packages.rag.embeddings import (
    embed_text,
    to_pgvector_literal,
)


def ingest_text_document(db: Session, request: IngestTextRequest) -> IngestTextResponse:
    with trace_span("ingest.text", title=request.title):
        chunks = chunk_text(
            request.content,
            max_chars=request.chunk_size,
            overlap_chars=request.chunk_overlap,
        )
        if not chunks:
            raise ValueError("No chunks generated from content")

        try:
            document_id = _insert_document(db, title=request.title, content=request.content)
            embedding_count = 0
            first_embedding_provider = ""
            first_embedding_model = ""

            for chunk_index, chunk in enumerate(chunks):
                chunk_id = _insert_chunk(
                    db,
                    document_id=document_id,
                    chunk_index=chunk_index,
                    content=chunk,
                )
                embedding = embed_text(chunk)
                _insert_embedding(
                    db,
                    chunk_id=chunk_id,
                    vector_literal=to_pgvector_literal(embedding.vector),
                    provider=embedding.provider,
                    model=embedding.model,
                )
                if chunk_index == 0:
                    first_embedding_provider = embedding.provider
                    first_embedding_model = embedding.model
                embedding_count += 1

            db.commit()
        except Exception:
            db.rollback()
            raise

        return IngestTextResponse(
            document_id=document_id,
            chunk_count=len(chunks),
            embedding_count=embedding_count,
            embedding_provider=first_embedding_provider,
            embedding_model=first_embedding_model,
        )


def _insert_document(db: Session, title: str, content: str) -> int:
    result = db.execute(
        text(
            """
            INSERT INTO documents (title, content)
            VALUES (:title, :content)
            RETURNING id
            """
        ),
        {"title": title, "content": content},
    )
    return cast(int, result.scalar_one())


def _insert_chunk(db: Session, document_id: int, chunk_index: int, content: str) -> int:
    result = db.execute(
        text(
            """
            INSERT INTO chunks (document_id, chunk_index, content)
            VALUES (:document_id, :chunk_index, :content)
            RETURNING id
            """
        ),
        {
            "document_id": document_id,
            "chunk_index": chunk_index,
            "content": content,
        },
    )
    return cast(int, result.scalar_one())


def _insert_embedding(
    db: Session, chunk_id: int, vector_literal: str, provider: str, model: str
) -> None:
    db.execute(
        text(
            """
            INSERT INTO embeddings (chunk_id, provider, model, embedding)
            VALUES (
                :chunk_id,
                :provider,
                :model,
                CAST(:embedding AS vector(8))
            )
            """
        ),
        {
            "chunk_id": chunk_id,
            "provider": provider,
            "model": model,
            "embedding": vector_literal,
        },
    )
