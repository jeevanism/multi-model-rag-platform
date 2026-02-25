# RAG Test Cases (Positive + Negative Proofs)

This file documents practical RAG test cases for this project, including:
- what we are proving
- how to test
- expected result
- proof interpretation ("hence proved")

This project currently uses:
- stub LLM providers (`gemini`, `openai`)
- stub embeddings

So these tests prove the **RAG pipeline mechanics** (retrieval, grounding, citations, cloud DB wiring), not final model intelligence quality.

## Test Scope

What these tests prove now:
- UI can call backend (local or cloud)
- `rag=true` triggers retrieval
- retrieved chunks are included in grounded prompt construction
- citations are returned
- cloud-ingested docs are retrieved from Cloud SQL

What these tests do not prove yet:
- real semantic retrieval quality (stub embeddings)
- real LLM reasoning quality (stub providers)
- production answer correctness using real model APIs

## Test Environments

### A. Local UI -> Cloud Backend (recommended for browser proof)

- Frontend: local Vite app (`http://localhost:5173`)
- Backend: Cloud Run (`https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app`)
- DB: Cloud SQL (Postgres + pgvector)

Set frontend API base URL:

```bash
cat > apps/web/.env.local <<'EOF'
VITE_API_BASE_URL=https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app
EOF
```

Run frontend:

```bash
cd apps/web
npm run dev
```

Open:
- `http://localhost:5173`

### B. API-Only (curl)

Use:

```bash
export CLOUD_RUN_URL="https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app"
```

## Precondition (Cloud Doc Must Exist)

Ingest at least one document into the cloud backend:

```bash
curl -s -X POST "$CLOUD_RUN_URL/ingest/text" \
  -H "Content-Type: application/json" \
  -d '{"title":"Cloud RAG Doc","content":"Paris is the capital of France. Berlin is the capital of Germany."}'
```

Expected:
- `200 OK`
- response with `document_id`, `chunk_count`, `embedding_count`

## Test Case 1 (Positive): Known Fact Present in Ingested Document

### What we are proving

When `rag=true`, the backend retrieves relevant chunk(s) from the cloud DB and returns citations/debug chunks.

### How to prove (UI)

In the UI:
- `API Base URL`: Cloud Run URL
- `Mode`: `/chat (JSON)` (or SSE also fine)
- `RAG`: ON
- `Debug`: ON
- ask: `What is the capital of France?`

### How to prove (curl)

```bash
curl -s -X POST "$CLOUD_RUN_URL/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"What is the capital of France?","provider":"gemini","rag":true,"top_k":2,"debug":true}'
```

### Expected result

- HTTP `200`
- `"rag_used": true`
- non-empty `"citations"`
- non-empty `"retrieved_chunks"` (because `debug=true`)
- retrieved chunk contains `Cloud RAG Doc`
- citation contains `[source:Cloud RAG Doc#chunk=0]`

### Hence proved

- RAG retrieval path is active
- cloud document retrieval works
- citation generation works
- grounded prompt construction is being used

## Test Case 2 (Positive): Another Fact Present in Same Document

### What we are proving

RAG can reuse the same ingested document to answer another in-context query and still cite the same source.

### How to prove

Ask:
- `What is the capital of Germany?`

UI settings:
- `RAG`: ON
- `Debug`: ON

curl:

```bash
curl -s -X POST "$CLOUD_RUN_URL/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"What is the capital of Germany?","provider":"gemini","rag":true,"top_k":2,"debug":true}'
```

### Expected result

- HTTP `200`
- `"rag_used": true`
- citations include `Cloud RAG Doc`
- retrieved chunks include text with `Berlin is the capital of Germany`

### Hence proved

- Same indexed content is retrievable for multiple queries
- retrieval/citation path is stable across requests

## Test Case 3 (Negative): Irrelevant Question Still Pulls Available Document (Stub Limitation)

### What we are proving

This test intentionally demonstrates current limitations of **stub embeddings + stub LLM** and proves pipeline behavior is mechanical, not semantic.

### How to prove

Ask:
- `What is 2+2?`

UI settings:
- `RAG`: ON
- `Debug`: ON

curl:

```bash
curl -s -X POST "$CLOUD_RUN_URL/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"What is 2+2?","provider":"gemini","rag":true,"top_k":2,"debug":true}'
```

### Expected result

- HTTP `200`
- `"rag_used": true`
- citations still reference `Cloud RAG Doc`
- retrieved chunks show the capitals document
- answer is stub-style and not a real mathematical answer

### Hence proved

- RAG pipeline is functioning (retrieval + citations + debug output)
- current retrieval quality is limited by stub embeddings
- current answer quality is limited by stub LLM provider

## Test Case 4 (Control): Same Question With `rag=false`

### What we are proving

Turning off RAG disables retrieval and citations, showing the difference between grounded and non-grounded execution paths.

### How to prove

Ask:
- `What is the capital of France?`

UI settings:
- `RAG`: OFF
- `Debug`: OFF (or ON; debug output should still be absent/null when RAG is off)

curl:

```bash
curl -s -X POST "$CLOUD_RUN_URL/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"What is the capital of France?","provider":"gemini","rag":false}'
```

### Expected result

- HTTP `200`
- `"rag_used": false`
- `"citations": []`
- `"retrieved_chunks": null`

### Hence proved

- RAG toggle correctly controls retrieval behavior
- endpoint supports both grounded and non-grounded paths

## Test Case 5 (API Health + Evals Readiness): DB-Backed Cloud Endpoints No Longer 500

### What we are proving

Cloud Run is correctly connected to Cloud SQL and DB-backed routes are operational.

### How to prove

```bash
curl -s "$CLOUD_RUN_URL/health"
curl -s "$CLOUD_RUN_URL/evals/runs"
```

### Expected result

- `/health` returns `200` JSON
- `/evals/runs` returns `[]` or a JSON list (not `Internal Server Error`)

### Hence proved

- Cloud Run + Secret Manager + Cloud SQL wiring is correct
- migrations were applied successfully

## Optional Test Case 6 (SSE): Streaming Still Works With Cloud Backend

### What we are proving

Cloud deployment preserves streaming endpoint behavior.

### How to prove

```bash
curl -N -X POST "$CLOUD_RUN_URL/chat/stream" \
  -H "Content-Type: application/json" \
  -d '{"message":"hello stream","provider":"gemini"}'
```

### Expected result

- SSE events appear in order:
  - `start`
  - one or more `token`
  - `end`

### Hence proved

- Cloud Run deployment supports SSE for this endpoint

## Evidence Checklist (For README / Portfolio)

Capture and keep screenshots/outputs for:

1. UI positive RAG query (`France`) showing citation + retrieved chunk
2. UI negative RAG query (`2+2`) showing irrelevant retrieval (stub limitation proof)
3. UI control query with `rag=false` showing no citations
4. `curl` output for `/evals/runs` on Cloud Run (non-500)
5. `curl` output for `/chat/stream` SSE on Cloud Run

## Notes for Future Upgrade (Real Models/Embeddings)

When real providers and embeddings are added, repeat these same tests and update expected results:
- Negative test (`2+2`) should no longer retrieve unrelated capital docs if embeddings are semantic and retrieval thresholds exist.
- Positive tests should remain stable and become semantically stronger.
