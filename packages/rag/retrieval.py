from __future__ import annotations

from dataclasses import dataclass
from typing import cast

from sqlalchemy import text
from sqlalchemy.orm import Session

from packages.rag.embeddings import embed_text_deterministic, to_pgvector_literal


@dataclass(frozen=True)
class RetrievedChunk:
    document_id: int
    chunk_id: int
    chunk_index: int
    title: str
    content: str
    score: float


def retrieve_chunks(db: Session, query: str, top_k: int = 3) -> list[RetrievedChunk]:
    vector_literal = to_pgvector_literal(embed_text_deterministic(query))
    rows = db.execute(
        text(
            """
            SELECT
                d.id AS document_id,
                c.id AS chunk_id,
                c.chunk_index AS chunk_index,
                d.title AS title,
                c.content AS content,
                (e.embedding <-> CAST(:embedding AS vector(8))) AS distance
            FROM embeddings e
            JOIN chunks c ON c.id = e.chunk_id
            JOIN documents d ON d.id = c.document_id
            ORDER BY e.embedding <-> CAST(:embedding AS vector(8)) ASC
            LIMIT :top_k
            """
        ),
        {"embedding": vector_literal, "top_k": top_k},
    )

    results: list[RetrievedChunk] = []
    for row in rows.mappings():
        results.append(
            RetrievedChunk(
                document_id=cast(int, row["document_id"]),
                chunk_id=cast(int, row["chunk_id"]),
                chunk_index=cast(int, row["chunk_index"]),
                title=cast(str, row["title"]),
                content=cast(str, row["content"]),
                score=float(cast(float, row["distance"])),
            )
        )
    return results
