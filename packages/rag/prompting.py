from __future__ import annotations

from packages.rag.citations import format_citations
from packages.rag.retrieval import RetrievedChunk


def build_grounded_prompt(question: str, chunks: list[RetrievedChunk]) -> str:
    if not chunks:
        return question

    context_lines: list[str] = []
    citations = format_citations(chunks)
    for chunk, citation in zip(chunks, citations):
        context_lines.append(f"{citation} {chunk.content}")

    context = "\n".join(context_lines)
    return (
        "Answer the question using only the provided context. "
        "Include citations using the provided source tags.\n\n"
        f"Context:\n{context}\n\nQuestion:\n{question}"
    )
