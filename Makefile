PYTHON ?= python3
UVICORN ?= uvicorn

.PHONY: up down logs api test lint format typecheck ci-check

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f postgres

api:
	$(UVICORN) apps.api.main:app --reload --host 0.0.0.0 --port 8000

test:
	pytest -q

lint:
	ruff check .

format:
	ruff format .

typecheck:
	mypy apps/api tests

ci-check: lint typecheck test

