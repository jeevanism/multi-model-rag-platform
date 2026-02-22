PYTHON ?= python3
UVICORN ?= uvicorn
PSQL ?= psql

.PHONY: up down logs api test lint format typecheck ci-check migrate db-shell db-check eval-smoke

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f postgres

api:
	$(UVICORN) apps.api.main:app --reload --host 0.0.0.0 --port 8000

migrate:
	for f in migrations/*.sql; do \
		echo "Applying $$f"; \
		$(PSQL) "$${PSQL_DATABASE_URL:-postgresql://postgres:postgres@localhost:5432/multimodel_rag}" -f $$f; \
	done

db-shell:
	$(PSQL) "$${PSQL_DATABASE_URL:-postgresql://postgres:postgres@localhost:5432/multimodel_rag}"

db-check:
	$(PSQL) "$${PSQL_DATABASE_URL:-postgresql://postgres:postgres@localhost:5432/multimodel_rag}" -c "SELECT extname FROM pg_extension WHERE extname = 'vector';"

test:
	pytest -q

lint:
	ruff check .

format:
	ruff format .

typecheck:
	mypy apps/api tests

ci-check: lint typecheck test

eval-smoke:
	$(PYTHON) scripts/eval_run.py --limit 3
