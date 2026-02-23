FROM ghcr.io/astral-sh/uv:python3.11-bookworm-slim AS builder

WORKDIR /app

# Copy dependency metadata first for better layer caching.
COPY pyproject.toml uv.lock ./

# Copy source packages required by the API package installation.
COPY apps ./apps
COPY packages ./packages

# Create a production virtualenv from the lockfile.
RUN uv sync --frozen --no-dev

FROM python:3.11-slim AS runtime

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PATH="/app/.venv/bin:${PATH}" \
    PORT=8080

COPY --from=builder /app/.venv /app/.venv
COPY apps ./apps
COPY packages ./packages
COPY migrations ./migrations
COPY scripts ./scripts

EXPOSE 8080

CMD ["uvicorn", "apps.api.main:app", "--host", "0.0.0.0", "--port", "8080"]

