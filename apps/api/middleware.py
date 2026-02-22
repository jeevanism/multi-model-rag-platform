from __future__ import annotations

import time
import uuid

from fastapi import Request, Response

from packages.observability.logging import log_event
from packages.observability.request_context import set_request_id


async def request_observability_middleware(request: Request, call_next) -> Response:  # type: ignore[no-untyped-def]
    request_id = request.headers.get("x-request-id") or str(uuid.uuid4())
    set_request_id(request_id)
    start = time.perf_counter()
    log_event(
        "request.start",
        request_id=request_id,
        method=request.method,
        path=request.url.path,
    )
    try:
        response: Response = await call_next(request)
    except Exception:
        elapsed_ms = round((time.perf_counter() - start) * 1000, 2)
        log_event(
            "request.error",
            request_id=request_id,
            method=request.method,
            path=request.url.path,
            duration_ms=elapsed_ms,
        )
        raise

    elapsed_ms = round((time.perf_counter() - start) * 1000, 2)
    response.headers["x-request-id"] = request_id
    log_event(
        "request.end",
        request_id=request_id,
        method=request.method,
        path=request.url.path,
        status_code=response.status_code,
        duration_ms=elapsed_ms,
    )
    return response
