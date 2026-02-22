from __future__ import annotations

from packages.rag.retrieval import RetrievedChunk


def format_citations(chunks: list[RetrievedChunk]) -> list[str]:
    return [f"[source:{chunk.title}#chunk={chunk.chunk_index}]" for chunk in chunks]
