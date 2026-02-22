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
1. `uv venv`
2. `source .venv/bin/activate`
3. `uv pip install -e ".[dev]"`

Notes:
- Do not use raw `pip install ...` commands in project docs/scripts unless there is a specific reason.
- `pyproject.toml` remains the source of dependency definitions.

## Engineering Standards
Use `python-practice.md` as the Python implementation standard for this project (coding style, architecture boundaries, testing expectations, and quality gates).

Before opening a PR, verify your changes follow:
- `python-practice.md`
- the local checks in this README (`ruff`, `mypy`, `pytest`)

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

## Suggested Build Sequence (Dependency-First)
### Phase 1: Foundation
Status: `Complete`
- Repo scaffold
- Docker Compose with Postgres + pgvector
- FastAPI `/health`
- DB connectivity + migration setup
- CI pipeline + local quality gates

### Phase 2: Model Access
Status: `Complete` (stub adapters; real SDK integration pending)
- LLM provider abstraction (Gemini + OpenAI)
- Unified response types + retries/timeouts
- `/chat` endpoint (non-streaming first)
- `/chat/stream` via SSE

### Phase 3: RAG Core
Status: `Complete` (retrieval + citations working; embeddings/provider generation still stub-based)
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
- Structured logging + request correlation
- OpenTelemetry spans (request/retrieval/llm)
- Dockerfile + Cloud Run deployment path

### Phase 6: Product Surface and Polish
- React UI (chat streaming, model selector, citations, latency/cost)
- Eval dashboard
- Final docs, screenshots, architecture/tradeoffs

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
- [ ] Structured logging + tracing
- [ ] Cloud Run deployment
- [ ] React UI + eval dashboard
- [ ] Final screenshots + architecture docs

## Immediate Next Steps (Practical)
1. Implement Iteration 10: observability (structured logging + tracing).
2. Keep updating this checklist after each verified push.
3. Do not start UI work until retrieval and citations are working and tested.
