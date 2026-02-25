## Project Goal

We are building a production-oriented multi-model RAG platform that demonstrates:
- chat and streaming APIs
- model abstraction across providers
- retrieval with Postgres + `pgvector`
- grounded responses with citations
- evaluation and regression checks
- observability and operational debugging
- cloud deployment on GCP
- a user-facing frontend for runtime testing and eval review

## What Success Looks Like

A successful version of this project should allow us to:
- ingest documents
- retrieve relevant chunks with vector search
- answer questions with citations
- compare model behavior with evals
- debug failures using logs/runbooks
- deploy and demonstrate the system in the cloud

## Architecture Direction

High-level components:
- `apps/api` — FastAPI backend (chat, ingest, eval APIs)
- `apps/web` — React frontend (chat + eval dashboard)
- `packages/llm` — provider abstraction layer
- `packages/rag` — chunking, embeddings, retrieval, prompting, citations
- `packages/evals` — dataset runner, scoring, persistence, gating
- `packages/observability` — structured logging + request/span tracing helpers
- Postgres + `pgvector` — documents, chunks, embeddings, eval history
- Cloud Run + Cloud SQL + Secret Manager + Firebase Hosting — deployment stack

## Delivery Approach 

We build in small, verifiable iterations:
1. Implement the smallest useful slice
2. Add/update tests
3. Prove behavior with `curl` / logs / UI checks
4. Commit and push
5. Move to the next dependency-safe iteration

This keeps delivery fast while preserving production-style discipline.

## Phase Plan  

### Phase 1: Foundation
Objective:
- Create a reliable local and CI-ready development base

Scope:
- repo scaffold
- FastAPI baseline app (`/health`)
- Docker Compose (Postgres + pgvector)
- Python tooling (`uv`) and quality gates
- CI pipeline (lint/typecheck/tests)

Outcome:
- We can build and test backend slices safely and repeatedly.

### Phase 2: Model Access
Objective:
- Introduce a provider abstraction that keeps API/UI stable while internals evolve

Scope:
- `LLMProvider` interface
- provider router/factory
- Gemini adapter
- OpenAI adapter
- unified response schema
- `/chat` and `/chat/stream` endpoints

Outcome:
- We can swap providers without changing API contracts.

### Phase 3: RAG Core
Objective:
- Implement a real retrieval pipeline with grounded prompting and citations

Scope:
- RAG schema (`documents`, `chunks`, `embeddings`)
- text ingestion pipeline
- chunking
- embedding generation and vector storage
- pgvector retrieval
- grounded prompt construction
- citation formatting

Outcome:
- We can ingest content and answer with retrieved context + citations.

### Phase 4: Evaluation and Regression Safety
Objective:
- Add measurable quality checks to compare changes and catch regressions

Scope:
- eval dataset (`jsonl`)
- eval runner
- judge/scoring heuristics (correctness, groundedness, hallucination)
- eval persistence in DB
- regression gating against baseline

Outcome:
- We can compare behavior across versions/providers and make controlled changes.

### Phase 5: Observability and Deployment
Objective:
- Make the system inspectable and deployable in a production-like environment

Scope:
- structured JSON logs
- request IDs and span timing logs
- Dockerfile for backend
- Cloud Run deployment path
- Cloud SQL integration
- Secret Manager integration

Outcome:
- We can operate, debug, and demonstrate the app in cloud infrastructure.

### Phase 6: Product Surface and Demo Readiness
Objective:
- Provide a usable UI and reviewer-friendly proof of functionality

Scope:
- React UI (chat + streaming)
- citations and retrieved chunk debug panel
- eval dashboard UI
- Firebase Hosting deployment
- recruiter/demo documentation and runbooks

Outcome:
- We can share a live demo and show end-to-end functionality clearly.

## Iteration Sequence (Dependency-First)

This is the practical order we follow:

1. Repo scaffold + `/health` + CI
2. DB connectivity + pgvector extension migration
3. LLM provider abstraction (initial stubs)
4. `/chat` endpoint (non-streaming)
5. `/chat/stream` SSE
6. RAG schema + ingestion + chunking
7. Retrieval + grounded prompt + citations
8. Eval runner + persistence
9. Judge scoring
10. Regression gating
11. Observability
12. Cloud deployment path
13. React UI + eval dashboard
14. Real provider/embedding rollout (Gemini first)

## Current Achievement Summary 

We have already achieved:
- hosted frontend (Firebase) + hosted backend (Cloud Run)
- Cloud SQL-backed RAG and eval storage
- real Gemini generation in cloud
- real Gemini embeddings in cloud
- RAG citations and debug retrieval display in UI
- eval dashboard showing persisted runs and case-level scores
- GCP runbooks and troubleshooting documentation

## Current Provider Status

- Gemini: real generation + real embeddings enabled and validated 
- OpenAI: adapter implemented; runtime validation pending (no API key available) 
- Qwen / DeepSeek: planned as next provider expansion 

## Roadmap (Next Phases)

### Near-Term
- Add/validate more providers (OpenAI, Qwen, DeepSeek)
- Create real-mode eval baselines (separate from stub-era baseline)
- Finish final demo assets (video link + concise proofs in docs)

### Platform Improvements
- CI/CD automation:
  - backend deploy to Cloud Run
  - frontend deploy to Firebase Hosting
- stronger secret rotation/process docs
- optional telemetry upgrade (OpenTelemetry exporter)

### Retrieval Quality Improvements
- higher-dimensional embeddings migration
- re-index/re-ingest corpus after embedding model upgrades
- retrieval tuning (thresholds / reranking)

## Related Docs (Operational + Proof)

Use these documents for setup, troubleshooting, and reproducible demos:
- `README.md` — recruiter-facing project overview + live links
- `docs/gcp-commands.md` — reusable GCP command reference
- `docs/gcp-setup-steps.md` — chronological cloud setup/deploy runbook
- `docs/troubleshooting.md` — cloud issues and fixes
- `docs/RAG-TEST.md` — positive/negative RAG test cases and proof criteria
- `docs/database_design.md` — schema and vector design notes
