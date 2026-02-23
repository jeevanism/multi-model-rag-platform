from __future__ import annotations

import os
from dataclasses import dataclass, field


def _parse_csv_env(name: str, default: list[str]) -> list[str]:
    raw = os.getenv(name)
    if not raw:
        return default
    values = [item.strip() for item in raw.split(",")]
    return [item for item in values if item]


@dataclass(frozen=True)
class Settings:
    database_url: str = os.getenv(
        "DATABASE_URL",
        "postgresql+psycopg://postgres:postgres@localhost:5432/multimodel_rag",
    )
    log_level: str = os.getenv("LOG_LEVEL", "info")
    enable_tracing: bool = os.getenv("ENABLE_TRACING", "true").lower() == "true"
    cors_allow_origins: list[str] = field(
        default_factory=lambda: _parse_csv_env(
            "CORS_ALLOW_ORIGINS",
            [
                "http://localhost:5173",
                "http://127.0.0.1:5173",
            ],
        )
    )


settings = Settings()
