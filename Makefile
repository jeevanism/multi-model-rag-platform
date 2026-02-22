PYTHON ?= python3
UVICORN ?= uvicorn
PSQL ?= psql

.PHONY: up down logs api test lint format typecheck ci-check migrate db-shell db-check eval-smoke eval-gate docker-build docker-run deploy-cloud-run

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
	uv run python scripts/eval_run.py --limit 3

eval-gate:
	uv run python scripts/eval_run.py --limit 3 --output .tmp/eval_current.json
	uv run python scripts/eval_gate.py --current .tmp/eval_current.json

docker-build:
	docker build -t multi-model-rag-api:local .

docker-run:
	docker run --rm -p 8080:8080 --env-file .env.example multi-model-rag-api:local

deploy-cloud-run:
	bash infra/deploy_cloud_run.sh
