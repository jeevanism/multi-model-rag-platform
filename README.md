# Multi-Model RAG Platform

Production-oriented RAG application built end-to-end with:
- Go backend (`/chat`, `/chat/stream`, `/ingest/text`, eval APIs)
- Postgres + `pgvector` retrieval
- React UI (chat + eval dashboard)
- Eval runner + scoring + regression gate
- Observability (structured logs + request/span tracing)
- GCP deployment (Cloud Run + Cloud SQL + Secret Manager)
- Firebase Hosting frontend

## Live Demo
- Frontend (Firebase): [https://multi-model-rag-5713b.web.app](https://multi-model-rag-5713b.web.app)
- Backend API (Cloud Run): [https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app](https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app)
- API docs (Swagger): [https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app/docs](https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app/docs)
- Demo Video : https://youtu.be/iM-_bXSPgOQ

Note:
- I pause Cloud SQL outside demo/testing windows to control billing on a personal GCP account.
- If DB is paused, DB-backed features (`/ingest/text`, `/evals/*`, `rag=true`) will fail until Cloud SQL is resumed.

## Project Goal
Build a production-grade, production-shaped RAG system that demonstrates:
- model abstraction (multi-provider design)
- retrieval + grounding + citations
- evaluation and regression safety
- cloud deployment and operations
- observability and debugging discipline

## What Is Working (Current State)

### Core Product
- Chat API (`/chat`) and streaming chat (`/chat/stream` SSE)
- Text ingestion (`/ingest/text`)
- RAG retrieval via `pgvector`
- Citation-grounded responses
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

### Cloud Deployment
- Backend deployed to Cloud Run
- Cloud SQL Postgres attached to Cloud Run
- Secret Manager for runtime secrets (`DATABASE_URL`, `GEMINI_API_KEY`)
- Frontend deployed to Firebase Hosting

## Provider Status (Important)
- `Gemini`: real generation + real embeddings enabled and validated in cloud ✅
- `OpenAI`: adapter implemented, runtime validation pending (no OpenAI API key configured) ⏳
- `Qwen` / `DeepSeek`: planned next (provider expansion roadmap) ⏳

## Architecture (Implemented)
High-level flow:

1. `apps/web` sends requests to `apps/api`
2. `apps/api` handles API schemas + service orchestration
3. `packages/llm` routes provider calls (Gemini/OpenAI abstraction)
4. `packages/rag` handles chunking, embeddings, retrieval, grounding, citations
5. `packages/evals` runs/persists evals and applies scoring/gates
6. Postgres + `pgvector` stores documents/chunks/embeddings + eval history
7. `packages/observability` emits structured request/span logs

### Main Components
- `cmd/api` + `internal/*`: Go backend
- `apps/api`: legacy FastAPI reference implementation kept during migration
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
- Cloud operations/debugging workflows are documented and reproducible

## Tradeoffs / Current Limitations
- OpenAI runtime path is implemented but not validated yet (no key available)
- Additional providers (Qwen/DeepSeek) not integrated yet
- CI is implemented; CD is still manual (deploy scripts + commands)
- Tracing is currently log-based spans (not full OpenTelemetry exporter)

## Roadmap (Near-Term)
1. Provider expansion: Qwen / DeepSeek / OpenAI runtime validation
2. CI/CD automation:
   - backend deploy to Cloud Run
   - frontend deploy to Firebase Hosting
3. Real-mode baseline management (separate baselines per provider/model)
4. Further RAG quality improvements (retrieval tuning/reranking)

## Local Quickstart

### Backend
```bash
go mod tidy

make up
make migrate
go run ./cmd/api
```

Legacy Python backend reference:
```bash
uv venv --python 3.11
source .venv/bin/activate
uv pip install -e ".[dev]"
uv run uvicorn apps.api.main:app --reload --port 8000
```

### Frontend
```bash
cd apps/web
npm install
npm run dev
```

Open `http://localhost:5173`

### Checks
```bash
uv run ruff format .
uv run ruff check .
uv run mypy apps/api tests packages scripts
uv run pytest -q
GOCACHE=/tmp/go-build-cache go test ./cmd/... ./internal/...
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
