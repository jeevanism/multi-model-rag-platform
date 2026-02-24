from __future__ import annotations

import hashlib
import os
from dataclasses import dataclass
from typing import Any


EMBEDDING_DIM = 8
STUB_EMBEDDING_PROVIDER = "stub"
STUB_EMBEDDING_MODEL = "stub-embedding-v1"


@dataclass(frozen=True)
class EmbeddingResult:
    vector: list[float]
    provider: str
    model: str
    raw: dict[str, Any] | None = None


def _embedding_mode() -> str:
    return os.getenv("EMBEDDING_PROVIDER_MODE", "stub").strip().lower()


def _embedding_provider_name() -> str:
    return os.getenv("EMBEDDING_PROVIDER", "gemini").strip().lower()


def embed_text(text: str, dimensions: int = EMBEDDING_DIM) -> EmbeddingResult:
    mode = _embedding_mode()

    if mode == "stub":
        return _embed_stub(text, dimensions)
    if mode == "auto":
        if _embedding_provider_name() == "gemini" and os.getenv("GEMINI_API_KEY"):
            return _embed_gemini(text, dimensions)
        if _embedding_provider_name() == "openai" and os.getenv("OPENAI_API_KEY"):
            return _embed_openai(text, dimensions)
        return _embed_stub(text, dimensions)
    if mode == "real":
        provider = _embedding_provider_name()
        if provider == "gemini":
            return _embed_gemini(text, dimensions)
        if provider == "openai":
            return _embed_openai(text, dimensions)
        raise ValueError(f"Unsupported EMBEDDING_PROVIDER '{provider}'. Use gemini or openai.")

    raise ValueError(f"Unsupported EMBEDDING_PROVIDER_MODE '{mode}'. Use stub, auto, or real.")


def _embed_stub(text: str, dimensions: int) -> EmbeddingResult:
    return EmbeddingResult(
        vector=embed_text_deterministic(text, dimensions),
        provider=STUB_EMBEDDING_PROVIDER,
        model=STUB_EMBEDDING_MODEL,
        raw={"stub": True},
    )


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


def _embed_gemini(text: str, dimensions: int) -> EmbeddingResult:
    api_key = os.getenv("GEMINI_API_KEY")
    if not api_key:
        raise RuntimeError("GEMINI_API_KEY is required for real Gemini embeddings.")

    try:
        from google import genai  # type: ignore[import-not-found,attr-defined]
    except ImportError as exc:
        raise RuntimeError(
            "google-genai package is required for real Gemini embeddings. "
            'Install with: uv pip install "google-genai>=1.0.0"'
        ) from exc

    model = os.getenv("GEMINI_EMBEDDING_MODEL", "gemini-embedding-001")
    client = genai.Client(api_key=api_key)

    response = client.models.embed_content(
        model=model,
        contents=text,
        config={"output_dimensionality": dimensions},
    )

    embeddings = getattr(response, "embeddings", None)
    if not embeddings:
        raise RuntimeError("Gemini embedding response did not include embeddings.")

    first = embeddings[0]
    values = getattr(first, "values", None)
    if not isinstance(values, list) or not values:
        raise RuntimeError("Gemini embedding response did not include embedding values.")

    vector = [float(v) for v in values][:dimensions]
    return EmbeddingResult(vector=vector, provider="gemini", model=model, raw={"stub": False})


def _embed_openai(text: str, dimensions: int) -> EmbeddingResult:
    api_key = os.getenv("OPENAI_API_KEY")
    if not api_key:
        raise RuntimeError("OPENAI_API_KEY is required for real OpenAI embeddings.")

    try:
        from openai import OpenAI  # type: ignore[import-not-found]
    except ImportError as exc:
        raise RuntimeError(
            "openai package is required for real OpenAI embeddings. "
            'Install with: uv pip install "openai>=1.0.0"'
        ) from exc

    model = os.getenv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small")
    client = OpenAI(api_key=api_key)
    response = client.embeddings.create(model=model, input=text, dimensions=dimensions)
    data = getattr(response, "data", None)
    if not data:
        raise RuntimeError("OpenAI embedding response did not include data.")

    embedding = getattr(data[0], "embedding", None)
    if not isinstance(embedding, list) or not embedding:
        raise RuntimeError("OpenAI embedding response did not include embedding values.")

    vector = [float(v) for v in embedding][:dimensions]
    return EmbeddingResult(vector=vector, provider="openai", model=model, raw={"stub": False})


def to_pgvector_literal(vector: list[float]) -> str:
    return "[" + ",".join(f"{value:.6f}" for value in vector) + "]"
