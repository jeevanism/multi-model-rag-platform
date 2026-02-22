from __future__ import annotations

import pytest

from packages.rag.chunking import chunk_text
from packages.rag.embeddings import embed_text_deterministic, to_pgvector_literal


def test_chunk_text_respects_max_chars() -> None:
    text = "word " * 80

    chunks = chunk_text(text, max_chars=50, overlap_chars=10)

    assert chunks
    assert all(len(chunk) <= 50 for chunk in chunks)


def test_chunk_text_returns_empty_for_blank_input() -> None:
    assert chunk_text("   \n\t  ") == []


def test_chunk_text_rejects_invalid_overlap() -> None:
    with pytest.raises(ValueError):
        chunk_text("hello", max_chars=10, overlap_chars=10)


def test_deterministic_embedding_is_stable_and_pgvector_formatted() -> None:
    vector1 = embed_text_deterministic("hello")
    vector2 = embed_text_deterministic("hello")
    vector3 = embed_text_deterministic("different")

    assert vector1 == vector2
    assert vector1 != vector3
    assert len(vector1) == 8
    assert to_pgvector_literal(vector1).startswith("[")
