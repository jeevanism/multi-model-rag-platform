from __future__ import annotations

import time
from contextlib import contextmanager
from collections.abc import Iterator
from typing import Any

from packages.observability.logging import log_event
from packages.observability.request_context import get_request_id


@contextmanager
def trace_span(name: str, **fields: Any) -> Iterator[None]:
    start = time.perf_counter()
    log_event("span.start", span=name, request_id=get_request_id(), **fields)
    try:
        yield
    finally:
        elapsed_ms = round((time.perf_counter() - start) * 1000, 2)
        log_event("span.end", span=name, request_id=get_request_id(), duration_ms=elapsed_ms)
