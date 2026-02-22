from __future__ import annotations

import hashlib


EMBEDDING_DIM = 8
EMBEDDING_PROVIDER = "stub"
EMBEDDING_MODEL = "stub-embedding-v1"


def embed_text_deterministic(text: str, dimensions: int = EMBEDDING_DIM) -> list[float]:
    if dimensions <= 0:
        raise ValueError("dimensions must be > 0")

    digest = hashlib.sha256(text.encode("utf-8")).digest()
    values: list[float] = []

    for i in range(dimensions):
        byte = digest[i % len(digest)]
        # Map byte [0,255] to float [-1.0, 1.0]
        value = (byte / 127.5) - 1.0
        values.append(round(value, 6))

    return values


def to_pgvector_literal(vector: list[float]) -> str:
    return "[" + ",".join(f"{value:.6f}" for value in vector) + "]"
