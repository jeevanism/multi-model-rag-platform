# Multi-Model RAG Platform

Production-oriented RAG application built as a showcase of how the same product can be implemented with both Go and Python/FastAPI backends while preserving shared frontend contracts, retrieval architecture, evaluation tooling, and deployment shape.

Built end-to-end with:
- Go backend (`/chat`, `/chat/stream`, `/ingest/text`, eval APIs)
- Python/FastAPI backend for the same product surface
- Postgres + `pgvector` retrieval
- React UI (chat + eval dashboard)
- Eval runner + scoring + regression gate
- Observability (structured logs + request/span tracing)
- GCP deployment (Cloud Run + Cloud SQL + Secret Manager)
- Firebase Hosting frontend

## Live Demo
- Frontend (Firebase): [https://multi-model-rag-5713b.web.app](https://multi-model-rag-5713b.web.app)
- Backend API (Cloud Run): [https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app](https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app)
- Demo Video : https://youtu.be/iM-_bXSPgOQ

Note:
- I pause Cloud SQL outside demo/testing windows to control billing on a personal GCP account.
- If DB is paused, DB-backed features (`/ingest/text`, `/evals/*`, `rag=true`) will fail until Cloud SQL is resumed.
- The local repository contains both backend implementations. Go is the current primary runtime path, and the FastAPI backend remains available for comparison, reference, and parity checks.

## Project Goal
Build a production-grade, production-shaped RAG system that demonstrates:
- model abstraction (multi-provider design)
- retrieval + grounding + citations
- evaluation and regression safety
- cloud deployment and operations
- observability and debugging discipline

## What Is Working (Current State)

### Migration Status
- Core FastAPI-to-Go migration plan: `11/11` phases complete
- Remaining work is hardening, parity review, real provider integrations, and deciding the long-term coexistence or cleanup path for the two backend implementations

### Core Product
- Go backend endpoints:
  - `GET /`
  - `GET /health`
  - `POST /chat`
  - `POST /chat/stream`
  - `POST /ingest/text`
  - `GET /evals/runs`
  - `GET /evals/runs/{id}`
  - `GET /auth/demo-status`
  - `POST /auth/demo-unlock`
  - `POST /auth/demo-lock`
- RAG ingestion and retrieval via Postgres + `pgvector`
- Citation-grounded responses
- Demo auth cookie flow
- React UI with:
  - JSON + SSE chat modes
  - provider selection
  - RAG/debug toggles
  - citations panel
  - retrieved chunks debug panel
  - eval dashboard (runs + case-level scores)

### Evaluation / Safety
- Eval dataset runner (`jsonl`)
- Judge scoring:
  - correctness
  - groundedness
  - hallucination
- Regression gating against baseline
- Eval persistence in DB (`eval_run`, `eval_run_case`)

### Runtime / Delivery
- Go server runs from `cmd/api` with packages under `internal/*`
- FastAPI backend also exists under `apps/api`
- Dockerfile builds and runs the Go backend
- Cloud SQL migrations remain SQL-first via `psql`
- Frontend remains unchanged in `apps/web`

## Provider Status (Important)
- Go backend currently implements deterministic stub behavior for:
  - `gemini`
  - `openai`
  - `grok`
- Default model names are wired for those providers, but real provider HTTP integrations in Go are still pending
- Real Gemini/OpenAI/Grok execution is post-migration follow-up work, not part of the completed core replacement

## Architecture (Implemented)
High-level flow:

1. `apps/web` sends requests to the Go API
2. `cmd/api` boots the server and loads env-based config
3. `internal/httpapi` handles routes, decoding, validation, and HTTP responses
4. `internal/service` orchestrates chat, ingest, and eval behavior
5. `internal/llm` provides provider routing and stub responses
6. `internal/rag` handles chunking, embeddings, grounding, citations, and prompt construction
7. `internal/store` accesses Postgres + `pgvector` via `database/sql` and `github.com/jackc/pgx/v5/stdlib`
8. `internal/observability` emits structured request and span logs

### Main Components
- `cmd/api` + `internal/*`: Go backend
- `apps/api`: Python/FastAPI backend for the same product surface
- `apps/web`: React + Vite frontend
- `packages/llm`: provider abstraction
- `packages/rag`: ingestion/retrieval/citations
- `packages/evals`: eval runner/scoring/gating
- `packages/observability`: logging/tracing helpers
- `infra/`: deployment scripts
- `datasets/`: eval datasets + baselines

## What This Project Proves
- End-to-end RAG pipeline works in a hosted environment (Firebase + Cloud Run + Cloud SQL)
- Retrieval and citations are visible in the UI and API
- Eval and regression tooling can catch behavior changes between stub and real providers
- The same RAG product can be implemented in both Go and Python/FastAPI without changing the frontend contract
- Cloud operations/debugging workflows are documented and reproducible

## Tradeoffs / Current Limitations
- Go backend real provider execution is not implemented yet; current Go behavior is stub-first
- Exact FastAPI validation and edge-case parity still needs a final review pass
- Maintaining both Go and FastAPI backends increases documentation and maintenance overhead
- CI is implemented; CD is still manual (deploy scripts + commands)
- Tracing is currently log-based spans (not full OpenTelemetry exporter)

## Roadmap (Near-Term)
1. Real Gemini, OpenAI, and Grok provider integrations in Go
2. Parity gap review against remaining FastAPI edge cases
3. Cutover cleanup and eventual deprecation of `apps/api`
4. CI/CD automation:
   - backend deploy to Cloud Run
   - frontend deploy to Firebase Hosting
5. Real-mode baseline management (separate baselines per provider/model)
6. Further RAG quality improvements (retrieval tuning/reranking)

## Local Quickstart

### Backend
```bash
cp .env.example .env

make up
make migrate
make api
```

Python/FastAPI backend:
```bash
uv venv --python 3.11
source .venv/bin/activate
uv pip install -e ".[dev]"
uv run uvicorn apps.api.main:app --reload --host 0.0.0.0 --port 8000
```

Backend runtime summary:
```text
Go API (primary):       http://localhost:8080
FastAPI API:            http://localhost:8000
```

The Go backend is the primary local and container runtime path:
```bash
make api
```

The FastAPI backend remains available for comparison and parity checks:
```bash
uv run uvicorn apps.api.main:app --reload --port 8000
```

### Frontend
```bash
cd apps/web
npm install
npm run dev
```

Open `http://localhost:5173`

Default local ports:
- Go API: `http://localhost:8080`
- Vite frontend: `http://localhost:5173`

### Checks
```bash
uv run ruff check .
uv run ruff format --check .
uv run mypy apps/api tests packages scripts
uv run pytest -q
GOCACHE=/tmp/go-build-cache go test ./cmd/... ./internal/...
docker build -t multi-model-rag-api:local .
```

## Cloud / Ops Runbooks
Use these docs for setup, troubleshooting, and reproducible cloud commands:

- `docs/gcp-commands.md` - reusable GCP command reference (Cloud Run, Cloud SQL, Secret Manager, logs, eval commands)
- `docs/gcp-setup-steps.md` - chronological cloud setup + deployment steps actually performed
- `docs/troubleshooting.md` - GCP/cloud issues encountered and how they were fixed
- `docs/RAG-TEST.md` - positive/negative RAG test plan and proof criteria
- `docs/database_design.md` - schema/data model notes, `pgvector` design, embedding dimension decisions

## Demo Notes (For Reviewers)
- The hosted demo is real cloud infrastructure (not localhost).
- If a live demo endpoint fails, it may be because Cloud SQL is paused to control costs.
- I can resume the DB and run a live walkthrough if needed.

## Repo History / Build Approach
This project was built iteratively (small slices with tests + proof commands + pushes), and the detailed planning artifacts remain in:
- `docs/project_plan.md`
