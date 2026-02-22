from __future__ import annotations

import json
import logging
import sys
from datetime import UTC, datetime
from typing import Any


def configure_logging(level: str = "INFO") -> None:
    logger = logging.getLogger("multi_model_rag")
    if logger.handlers:
        logger.setLevel(level.upper())
        return

    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(logging.Formatter("%(message)s"))
    logger.addHandler(handler)
    logger.setLevel(level.upper())
    logger.propagate = False


def get_logger() -> logging.Logger:
    return logging.getLogger("multi_model_rag")


def log_event(event: str, **fields: Any) -> None:
    payload: dict[str, Any] = {
        "timestamp": datetime.now(UTC).isoformat(),
        "event": event,
        **fields,
    }
    get_logger().info(json.dumps(payload, default=str))
