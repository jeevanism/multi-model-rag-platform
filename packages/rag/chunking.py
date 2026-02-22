from __future__ import annotations


def chunk_text(text: str, max_chars: int = 400, overlap_chars: int = 40) -> list[str]:
    normalized = " ".join(text.split())
    if not normalized:
        return []
    if max_chars <= 0:
        raise ValueError("max_chars must be > 0")
    if overlap_chars < 0:
        raise ValueError("overlap_chars must be >= 0")
    if overlap_chars >= max_chars:
        raise ValueError("overlap_chars must be smaller than max_chars")

    chunks: list[str] = []
    start = 0
    length = len(normalized)

    while start < length:
        end = min(start + max_chars, length)
        if end < length:
            split_at = normalized.rfind(" ", start, end)
            if split_at > start:
                end = split_at
        chunk = normalized[start:end].strip()
        if chunk:
            chunks.append(chunk)
        if end >= length:
            break
        start = max(0, end - overlap_chars)
        while start < length and normalized[start] == " ":
            start += 1

    return chunks
