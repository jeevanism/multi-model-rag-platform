import os

import pytest
from sqlalchemy import create_engine, text
from sqlalchemy.exc import SQLAlchemyError


def test_pgvector_extension_exists() -> None:
    database_url = os.getenv(
        "DATABASE_URL",
        "postgresql+psycopg://postgres:postgres@localhost:5432/multimodel_rag",
    )
    engine = create_engine(database_url, pool_pre_ping=True)

    try:
        with engine.connect() as connection:
            result = connection.execute(
                text("SELECT extname FROM pg_extension WHERE extname = 'vector'")
            ).scalar_one_or_none()
    except SQLAlchemyError as exc:
        pytest.skip(f"Database not reachable for integration test: {exc}")
    finally:
        engine.dispose()

    assert result == "vector"

