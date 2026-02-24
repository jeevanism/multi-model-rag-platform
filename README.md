# Multi-Model RAG (Planning + Execution Guide)

## Overview
This repository is a planning and execution workspace for building a production-oriented multi-model RAG application.

Target system capabilities:
- FastAPI backend with chat + streaming
- Multi-model provider abstraction (Gemini + OpenAI)
- RAG with Postgres + pgvector
- Evaluation harness (correctness, groundedness, hallucination, latency, cost)
- Observability (structured logs + tracing)
- React + TypeScript UI
- GCP deployment (Cloud Run + Cloud SQL + Secret Manager)

This `README.md` consolidates the content from:
- `project_plan.md` (roadmap + setup)
- `iteration_plan.md` (build/test/prove/commit workflow)

## Working Style (Required)
Build in small, verifiable increments.

For each iteration:
1. Implement the smallest usable slice.
2. Add or update tests.
3. Prove it works (CLI command / logs / curl).
4. Commit with a clear message.
5. Push.

If something fails:
1. Reproduce with the smallest command.
2. Add logging around the failing boundary.
3. Fix it.
4. Add a regression test.
5. Commit as `fix(...)`.

## Python Tooling Standard (`uv`)
Use `uv` for Python environment and package management across local development and CI.

Preferred workflow:
1. `uv venv --python 3.11`
2. `source .venv/bin/activate`
3. `uv pip install -e ".[dev]"`

Notes:
- Do not use raw `pip install ...` commands in project docs/scripts unless there is a specific reason.
- `pyproject.toml` remains the source of dependency definitions.
- Recommended Python version for this project: `3.11` (best compatibility with current provider SDKs and SSL stack).
- Python `3.14` may work for some paths, but we hit SDK/SSL runtime issues (`google-genai` / `aiohttp` / `ssl`) during real provider/embedding integration.
- Cloud Run runtime is also aligned to Python `3.11`, so using `3.11` locally reduces environment mismatch.

## Engineering Standards
Use `python-practice.md` as the Python implementation standard for this project (coding style, architecture boundaries, testing expectations, and quality gates).

Before opening a PR, verify your changes follow:
- `python-practice.md`
- the local checks in this README (`ruff`, `mypy`, `pytest`)

## Operational Docs
Use these docs for cloud setup, troubleshooting, and RAG proof/testing workflows:

- `gcp-commands.md` - reusable GCP command reference (Cloud Run, Cloud SQL, secrets, logs)
- `gcp-setup-steps.md` - chronological record of the cloud setup steps performed
- `troubleshooting.md` - GCP-specific errors encountered and how we fixed them
- `RAG-TEST.md` - positive/negative RAG test cases with proof criteria

## Quality Gates
Run before every push (once repo code exists):
- `ruff check .`
- `ruff format .` (or `black .`)
- `mypy apps/api`
- `pytest -q`

CI should eventually run:
- lint
- format check
- type check
- unit tests
- integration tests (with Docker Compose / pgvector)
- eval smoke tests (later)

## Project Architecture (Planned)
- `apps/api`: FastAPI backend
- `apps/web`: React + TypeScript frontend
- `packages/llm`: provider abstraction (Gemini/OpenAI)
- `packages/rag`: ingestion, chunking, retrieval
- `packages/evals`: eval runner, judge, aggregation, gating
- `packages/observability`: logging + tracing helpers
- `datasets/`: eval datasets
- `infra/`: deployment configs/scripts
- `docs/`: architecture and screenshots

## Architecture (Implemented)
High-level request flow:

1. `apps/web` (React UI) sends chat/eval requests to `apps/api`
2. `apps/api` routes requests through service-layer functions
3. `packages/llm` handles provider abstraction (Gemini real mode working; OpenAI adapter implemented)
4. `packages/rag` handles chunking, embeddings, retrieval, prompt grounding, citations (Gemini embeddings real mode working)
5. `packages/evals` runs eval datasets, scoring, aggregation, persistence, and regression gating
6. Postgres + `pgvector` stores documents/chunks/embeddings and eval run history
7. `packages/observability` emits request-level structured logs and span timing events

### Runtime Components
- API: FastAPI (`/health`, `/chat`, `/chat/stream`, `/ingest/text`, `/evals/runs`, `/evals/runs/{id}`)
- DB: Postgres + `pgvector`
- UI: Vite React app with Chat + Evals dashboard tabs
- Local infra: Docker Compose for Postgres, Dockerfile for API container

### Data Model (Current)
- RAG:
  - `documents`
  - `chunks`
  - `embeddings`
- Evals:
  - `eval_run`
  - `eval_run_case` (including judge score columns)

## Tradeoffs (Current State)
This project is intentionally portfolio-friendly and iteration-driven, so several components are production-shaped but stub-backed.

### What is production-shaped
- API boundaries and schemas
- provider abstraction layer
- pgvector retrieval path
- eval runner + persistence + gating
- structured logging and request tracing
- deployment path (Dockerfile + Cloud Run deploy script)

### What is intentionally stubbed / partially integrated (current)
- OpenAI provider runtime validation is pending (no OpenAI API key available yet)
- Additional providers (Qwen, DeepSeek) are planned but not integrated yet
- tracing is log-based spans (not full OpenTelemetry exporter yet)

### Why this tradeoff was chosen
- Enables end-to-end architecture, testing, and deployment workflow first
- Keeps iteration speed high
- Makes it easy to replace internals later without changing surface APIs/UI

### Provider Roadmap (Near-Term)
- Gemini: real generation + real embeddings are enabled and validated in the hosted cloud demo
- OpenAI: adapter is implemented, but runtime validation is pending because no OpenAI API key is currently configured
- Planned next providers: Qwen and DeepSeek (after demo/polish phase)

## Suggested Build Sequence (Dependency-First)
### Phase 1: Foundation
Status: `Complete`
- Repo scaffold
- Docker Compose with Postgres + pgvector
- FastAPI `/health`
- DB connectivity + migration setup
- CI pipeline + local quality gates

### Phase 2: Model Access
Status: `In Progress` (Gemini real generation working; OpenAI adapter implemented but runtime validation pending)
- LLM provider abstraction (Gemini + OpenAI)
- Unified response types + retries/timeouts
- `/chat` endpoint (non-streaming first)
- `/chat/stream` via SSE

### Phase 3: RAG Core
Status: `Complete` (retrieval + citations working; real Gemini embeddings enabled in cloud)
- RAG schema (`documents`, `chunks`, `embeddings`)
- Ingestion pipeline (text/markdown first)
- Chunking + embeddings storage
- Retrieval via pgvector
- Grounded prompting + citations

### Phase 4: Evaluation and Regression Safety
Status: `Complete` (dataset, runner, persistence, judge scoring, and regression gating complete)
- Eval dataset (`jsonl`)
- Eval runner + result persistence
- Judge prompts/scoring (groundedness/correctness/hallucination)
- Regression gating against baseline

### Phase 5: Observability and Deployment
- Status: `Complete` (deployment path + manual cloud deploy complete; CI/CD automation pending)
- Structured logging + request correlation
- OpenTelemetry spans (request/retrieval/llm)
- Dockerfile + Cloud Run deployment path

### Phase 6: Product Surface and Polish
- Status: `In Progress`
- React UI (chat streaming, model selector, citations, latency/cost)
- Eval dashboard
- Final docs, screenshots, architecture/tradeoffs
- CI/CD automation (GitHub Actions deploy to Cloud Run + Firebase Hosting)

## Iteration Plan (Execution Sequence)
Use this order for implementation:

1. Iteration 0: Repo + tooling baseline
   - API skeleton + `/health`
   - Docker Compose for pgvector
   - CI pipeline

2. Iteration 1: DB connectivity + migrations
   - SQLAlchemy engine/session
   - init migration enabling pgvector

3. Iteration 2: LLM provider abstraction
   - `LLMProvider` interface
   - Gemini/OpenAI adapters
   - unified response type
   - timeout/retry

4. Iteration 3: `/chat` API (non-streaming)
   - provider selection
   - unified response payload + metrics

5. Iteration 4: `/chat/stream` SSE
   - streaming response endpoint

6. Iteration 5: RAG schema + ingestion
   - tables + chunking + embeddings

7. Iteration 6: Retrieval + grounded prompt + citations
   - top-k retrieval + cited outputs

8. Iteration 7: Eval harness v1
   - dataset + runner + persisted results

9. Iteration 8: Judge prompts + scoring
   - groundedness / fact match / hallucination

10. Iteration 9: Regression gating
   - baseline + fail thresholds in CI

11. Iteration 10: Observability
   - structured logs + tracing

12. Iteration 11: Cloud deployment
   - Cloud Run + Cloud SQL + secrets

13. Iteration 12: React UI
   - chat + citations + model selector + telemetry display

14. Iteration 13: Dashboard + polish
   - eval UI + final docs/screenshots

## Definition of Done (Per Iteration)
An iteration is done only when:
- tests added or updated
- proof command succeeds locally
- CI passes (when configured)
- changes committed and pushed

## Quickstart (Local)
### 1. Start dependencies
From repo root:

```bash
make up
make migrate
```

### 2. Run backend
From repo root:

```bash
uv venv
source .venv/bin/activate
uv pip install -e ".[dev]"
uv run uvicorn apps.api.main:app --reload --port 8000
```

### 3. Run frontend
In a second terminal:

```bash
cd apps/web
npm install
npm run dev
```

Open `http://localhost:5173`

### 4. Run checks
Backend (repo root):

```bash
uv run ruff format .
uv run ruff check .
uv run mypy apps/api tests packages scripts
uv run pytest -q
```

Frontend (`apps/web`):

```bash
npm run build
```

## Eval Workflow (Local)
With API running on `http://localhost:8000`:

```bash
make eval-smoke
make eval-gate
```

Persist eval results to DB:

```bash
uv run python scripts/eval_run.py --limit 3 --persist
```

Inspect recent evals:

```bash
curl -s http://localhost:8000/evals/runs | jq
curl -s http://localhost:8000/evals/runs/1 | jq
```

## Commit Message Style
Examples:
- `chore: scaffold repo with api + pgvector + ci`
- `feat(db): add postgres connection and init migration`
- `feat(llm): add gemini and openai providers with unified interface`
- `feat(api): add SSE streaming chat endpoint`
- `fix(rag): correct pgvector similarity query ordering`

## What Not to Build (Yet)
Avoid overbuilding:
- Kubernetes / GKE
- Terraform (optional later)
- More than 2 model providers initially
- Full auth platform
- Multi-agent orchestration engine
- Fine-tuning pipelines
- Complex queues/caching unless needed

## Environment Variables (Planned `.env.example`)
- `DATABASE_URL=`
- `GEMINI_API_KEY=`
- `OPENAI_API_KEY=`
- `DEFAULT_ROUTING_MODE=manual`
- `DEFAULT_PROVIDER=gemini`
- `LOG_LEVEL=info`

## Actionable Checklist (Current Repo)
Use this checklist as the live progress tracker. Update it after each verified iteration / push.

### Phase 0: Bootstrap the repository
- [x] Create project folders:
  - `apps/api`
  - `apps/web`
  - `packages/llm`
  - `packages/rag`
  - `packages/evals`
  - `packages/observability`
  - `infra`
  - `datasets`
  - `docs`
- [x] Add `.gitignore`
- [x] Add `.env.example`
- [x] Add `docker-compose.yml` (Postgres + pgvector)
- [x] Add `pyproject.toml` (ruff/mypy/pytest config)
- [x] Add `Makefile`
- [x] Add `.github/workflows/ci.yml`

### Phase 1: Prove backend baseline
- [x] Create FastAPI app in `apps/api/main.py`
- [x] Add `GET /health`
- [x] Add `tests/test_health.py`
- [x] Run local proof: `curl http://localhost:8000/health`
- [x] Commit baseline scaffold

### Phase 2: Database readiness
- [x] Add DB connection module
- [x] Add migration `migrations/001_init.sql` with pgvector extension
- [x] Add integration test to verify `vector` extension exists
- [x] Commit DB setup

### Phase 3: LLM integration baseline
- [x] Add `LLMProvider` interface
- [x] Add Gemini provider adapter (stub)
- [x] Add OpenAI provider adapter (stub)
- [x] Add unified response schema
- [x] Add smoke script for provider calls
- [x] Commit provider abstraction

### Phase 4: Chat + Streaming + RAG
- [x] Add `/chat`
- [x] Add `/chat/stream` (SSE)
- [x] Add ingestion pipeline + schema (text ingestion + chunking + stub embeddings)
- [x] Add retrieval + citations (pgvector retrieval + citation-grounded chat)
- [x] Add tests and proof commands for each completed step

### Phase 5: Evals + Observability + Deploy + UI
- [x] Eval dataset + runner + judges + gating (including regression gating baseline checks)
- [x] Structured logging + tracing (request IDs, JSON logs, span timing events)
- [x] Cloud Run deployment path (Dockerfile + deploy script + local container `/health` proof complete)
- [x] React UI + eval dashboard (chat, streaming, citations, retrieved chunks, eval runs/case table)
- [ ] Final screenshots + architecture docs

## Screenshots Checklist (Portfolio Polish)
- [ ] Chat UI (JSON mode with citations + debug retrieved chunks)
- [ ] Chat UI (SSE mode message)
- [ ] Evals dashboard (runs list + case score table)
- [ ] API observability log sample (request + spans)
- [ ] Local container `/health` proof
- [ ] Optional Cloud Run `/health` + `/chat` proof

## Immediate Next Steps (Practical)
1. Finalize portfolio polish: screenshots + README architecture/tradeoffs/docs.
2. Add CI/CD automation for backend (Cloud Run) and frontend (Firebase Hosting) deploys.
3. Keep updating this checklist after each verified push.
3. Do not start UI work until retrieval and citations are working and tested.

## Cloud Run Deployment (Iteration 11)
Current status:
- `Dockerfile` is present and builds the FastAPI API image
- `infra/deploy_cloud_run.sh` builds and deploys to Cloud Run
- Script supports Cloud SQL attachment (`--add-cloudsql-instances`) and Secret Manager env injection (`--set-secrets`)

### Required GCP setup (minimum)
- Cloud Run service
- Cloud SQL Postgres instance (with `pgvector` available in DB)
- Secret Manager secrets for API keys / DB URL (recommended)
- IAM permissions for Cloud Build + Cloud Run deploy

### Deployment Modes
1. Direct `DATABASE_URL` (quickest path, external DB or already-routable DB)
2. Cloud SQL attachment + secret-managed `DATABASE_URL` (recommended for GCP)

### Example: Deploy with Secret Manager + Cloud SQL (Recommended)
Assumes:
- `DATABASE_URL` secret contains a SQLAlchemy URL (for Unix socket Cloud SQL, typically using `/cloudsql/<INSTANCE_CONNECTION_NAME>`)
- `GEMINI_API_KEY` and/or `OPENAI_API_KEY` secrets exist

```bash
export GCP_PROJECT_ID="your-project-id"
export GCP_REGION="us-central1"
export CLOUD_RUN_SERVICE="multi-model-rag-api"
export CLOUDSQL_INSTANCE="your-project-id:us-central1:rag-db"

# Optional runtime settings
export CLOUD_RUN_MEMORY="1Gi"
export CLOUD_RUN_CPU="1"
export CLOUD_RUN_MAX_INSTANCES="3"
export CLOUD_RUN_INGRESS="all"

# Secret Manager env injection for Cloud Run
export SECRET_ENV_VARS="DATABASE_URL=DATABASE_URL:latest,GEMINI_API_KEY=GEMINI_API_KEY:latest,OPENAI_API_KEY=OPENAI_API_KEY:latest"

# Non-secret env vars
export LOG_LEVEL="info"
export ENABLE_TRACING="true"
export DEFAULT_PROVIDER="gemini"
export DEFAULT_ROUTING_MODE="manual"

make deploy-cloud-run
```

### Example: Deploy with plain env `DATABASE_URL` (Non-secret / quick smoke only)
```bash
export GCP_PROJECT_ID="your-project-id"
export GCP_REGION="us-central1"
export DATABASE_URL="postgresql+psycopg://USER:PASSWORD@HOST:5432/DBNAME"

make deploy-cloud-run
```

### Post-Deploy Proof Commands
The deploy script prints the service URL and runs `/health` automatically (if `curl` is installed).

Manual checks:
```bash
curl -s "https://YOUR_SERVICE_URL/health" | jq

curl -s -X POST "https://YOUR_SERVICE_URL/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"hello","provider":"gemini"}' | jq
```

### Notes / Constraints
- The script deploys the API image only (UI is a later iteration).
- For private services (`ALLOW_UNAUTHENTICATED=false`), post-deploy curl will require authenticated requests.
- Cloud SQL connectivity requires the DB itself to have the schema/migrations applied and `vector` extension enabled.
