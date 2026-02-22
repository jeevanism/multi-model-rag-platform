from __future__ import annotations

import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    database_url: str = os.getenv(
        "DATABASE_URL",
        "postgresql+psycopg://postgres:postgres@localhost:5432/multimodel_rag",
    )
    log_level: str = os.getenv("LOG_LEVEL", "info")


settings = Settings()
