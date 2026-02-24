# GCP Commands Reference

This file is a working command reference for deploying and operating the `multi-model-rag-platform` project on Google Cloud.

Use it as a copy/paste checklist. Update it whenever we add a new GCP step.

## 1. Project Setup

Set the active project:

```bash
gcloud config set project multi-model-rag-platform
```

Export common variables (bash):

```bash
export PROJECT_ID="multi-model-rag-platform"
export REGION="us-central1"
export CLOUD_RUN_SERVICE="multi-model-rag-api"
export SQL_INSTANCE="multi-model-rag-pg"
export DB_NAME="multimodel_rag"
export DB_USER="postgres"
export DB_PASSWORD="REPLACE_WITH_A_REAL_STRONG_PASSWORD"
export DB_URL_SECRET="multi-model-rag-database-url"
```

## 2. Enable Required APIs

```bash
gcloud services enable \
  run.googleapis.com \
  sqladmin.googleapis.com \
  secretmanager.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  --project="$PROJECT_ID"
```

## 3. Cloud Run Deploy (Backend)

Local deploy env file flow (recommended):

```bash
cat > .env.deploy.local <<'EOF'
export GCP_PROJECT_ID="multi-model-rag-platform"
export GCP_REGION="us-central1"
export CLOUD_RUN_SERVICE="multi-model-rag-api"

export LOG_LEVEL="info"
export ENABLE_TRACING="true"
export DEFAULT_PROVIDER="gemini"
export DEFAULT_ROUTING_MODE="manual"
export CORS_ALLOW_ORIGINS="http://localhost:5173,http://127.0.0.1:5173"
EOF
```

Deploy using repo script:

```bash
source .env.deploy.local
make deploy-cloud-run
```

Fish shell:

```fish
bash -lc 'cd /home/jeevanism/Documents/Projects/AI-Engineering/multi-model-RAG && source .env.deploy.local && make deploy-cloud-run'
```

## 4. Cloud Run Verification

Use deployed backend URL:

```bash
export CLOUD_RUN_URL="https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app"
```

Health:

```bash
curl -s "$CLOUD_RUN_URL/health"
```

Chat (non-RAG):

```bash
curl -s -X POST "$CLOUD_RUN_URL/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"hello from cloud","provider":"gemini"}'
```

Chat stream (SSE):

```bash
curl -N -X POST "$CLOUD_RUN_URL/chat/stream" \
  -H "Content-Type: application/json" \
  -d '{"message":"hello stream","provider":"gemini"}'
```

Eval endpoints (will fail until DB is configured in cloud):

```bash
curl -s "$CLOUD_RUN_URL/evals/runs"
curl -s "$CLOUD_RUN_URL/evals/runs/2"
```

RAG chat (will fail until DB is configured in cloud):

```bash
curl -s -X POST "$CLOUD_RUN_URL/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"What is the capital of France?","provider":"gemini","rag":true,"top_k":2,"debug":true}'
```

## 5. Cloud Run Logs

Read recent logs:

```bash
gcloud run services logs read "$CLOUD_RUN_SERVICE" \
  --region="$REGION" \
  --project="$PROJECT_ID" \
  --limit=100
```

Tip:
- If DB-backed endpoints return `500`, check Cloud Run logs first for `sqlalchemy` / `psycopg` connection errors.

Describe service:

```bash
gcloud run services describe "$CLOUD_RUN_SERVICE" \
  --region="$REGION" \
  --project="$PROJECT_ID"
```

## 6. Cloud SQL (Postgres) - Create Instance

Current known working minimum for this project/org policy:
- `ENTERPRISE`
- `1 vCPU`
- `3840MB` RAM minimum

Create instance:

```bash
gcloud sql instances create "$SQL_INSTANCE" \
  --project="$PROJECT_ID" \
  --database-version=POSTGRES_16 \
  --region="$REGION" \
  --edition=ENTERPRISE \
  --cpu=1 \
  --memory=3840MB \
  --storage-size=10GB
```

Cloud SQL create behavior (normal):
- This can take several minutes (often 3 to 10+ minutes).
- A spinner for 10 seconds is normal.

Check create progress in another terminal:

```bash
gcloud sql instances list --project="$PROJECT_ID"
```

Look at `STATE`:
- `PENDING_CREATE` (or similar) = still creating
- `RUNNABLE` = ready

Inspect instance details:

```bash
gcloud sql instances describe "$SQL_INSTANCE" --project="$PROJECT_ID"
```

See Cloud SQL operations:

```bash
gcloud sql operations list \
  --instance="$SQL_INSTANCE" \
  --project="$PROJECT_ID"
```

Inspect a specific Cloud SQL operation:

```bash
gcloud sql operations describe OPERATION_ID --project="$PROJECT_ID"
```

Cloud Logging (Cloud SQL):

```bash
gcloud logging read \
  'resource.type="cloudsql_database" AND resource.labels.database_id:"'"$PROJECT_ID:$SQL_INSTANCE"'"' \
  --project="$PROJECT_ID" \
  --limit=20 \
  --format="value(timestamp, severity, textPayload)"
```

Set `postgres` user password:

```bash
gcloud sql users set-password "$DB_USER" \
  --instance="$SQL_INSTANCE" \
  --project="$PROJECT_ID" \
  --password="$DB_PASSWORD"
```

Create app database:

```bash
gcloud sql databases create "$DB_NAME" \
  --instance="$SQL_INSTANCE" \
  --project="$PROJECT_ID"
```

Get Cloud SQL connection name:

```bash
export INSTANCE_CONN_NAME="$(gcloud sql instances describe "$SQL_INSTANCE" \
  --project="$PROJECT_ID" \
  --format='value(connectionName)')"

echo "$INSTANCE_CONN_NAME"
```

## 7. Secret Manager (DATABASE_URL)

Build Cloud SQL socket-based SQLAlchemy URL:

```bash
export DATABASE_URL="postgresql+psycopg://${DB_USER}:${DB_PASSWORD}@/${DB_NAME}?host=/cloudsql/${INSTANCE_CONN_NAME}"
echo "$DATABASE_URL"
```

Create secret (first time):

```bash
printf '%s' "$DATABASE_URL" | gcloud secrets create "$DB_URL_SECRET" \
  --project="$PROJECT_ID" \
  --replication-policy=automatic \
  --data-file=-
```

Add secret version (subsequent updates):

```bash
printf '%s' "$DATABASE_URL" | gcloud secrets versions add "$DB_URL_SECRET" \
  --project="$PROJECT_ID" \
  --data-file=-
```

## 8. Grant Cloud Run Access to Secret

Find Cloud Run service account:

```bash
export RUN_SA="$(gcloud run services describe "$CLOUD_RUN_SERVICE" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --format='value(spec.template.spec.serviceAccountName)')"
echo "$RUN_SA"
```

Fallback to default compute service account (if needed):

```bash
export PROJECT_NUMBER="$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')"
export RUN_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"
echo "$RUN_SA"
```

Grant secret access:

```bash
gcloud secrets add-iam-policy-binding "$DB_URL_SECRET" \
  --project="$PROJECT_ID" \
  --member="serviceAccount:${RUN_SA}" \
  --role="roles/secretmanager.secretAccessor"
```

Grant Cloud Run access to Gemini API key secret (for real Gemini mode):

```bash
gcloud secrets add-iam-policy-binding GEMINI_API_KEY \
  --project="$PROJECT_ID" \
  --member="serviceAccount:${RUN_SA}" \
  --role="roles/secretmanager.secretAccessor"
```

## 9. Attach Cloud SQL + Secret to Cloud Run

Update service with Cloud SQL mount + `DATABASE_URL` secret + runtime env vars:

```bash
gcloud run services update "$CLOUD_RUN_SERVICE" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --add-cloudsql-instances="$INSTANCE_CONN_NAME" \
  --set-secrets="DATABASE_URL=${DB_URL_SECRET}:latest" \
  --update-env-vars="CORS_ALLOW_ORIGINS=http://localhost:5173,http://127.0.0.1:5173,LOG_LEVEL=info,ENABLE_TRACING=true,DEFAULT_PROVIDER=gemini,DEFAULT_ROUTING_MODE=manual"
```

## 10. Run Migrations on Cloud SQL

Connect with `psql`:

```bash
gcloud sql connect "$SQL_INSTANCE" \
  --project="$PROJECT_ID" \
  --user="$DB_USER" \
  --database="$DB_NAME"
```

At the `psql` prompt:

```sql
\i /home/jeevanism/Documents/Projects/AI-Engineering/multi-model-RAG/migrations/001_init.sql
\i /home/jeevanism/Documents/Projects/AI-Engineering/multi-model-RAG/migrations/002_rag_schema.sql
\i /home/jeevanism/Documents/Projects/AI-Engineering/multi-model-RAG/migrations/003_evals_schema.sql
\i /home/jeevanism/Documents/Projects/AI-Engineering/multi-model-RAG/migrations/004_eval_scores.sql
```

Basic verification in `psql`:

```sql
SELECT extname FROM pg_extension WHERE extname='vector';
SELECT COUNT(*) FROM documents;
SELECT COUNT(*) FROM eval_run;
\q
```

## 11. Cloud DB-Backed Endpoint Retest

After Cloud SQL + migrations are done:

```bash
curl -s "$CLOUD_RUN_URL/evals/runs"
```

Ingest cloud document:

```bash
curl -s -X POST "$CLOUD_RUN_URL/ingest/text" \
  -H "Content-Type: application/json" \
  -d '{"title":"Cloud RAG Doc","content":"Paris is the capital of France. Berlin is the capital of Germany."}'
```

RAG chat in cloud:

```bash
curl -s -X POST "$CLOUD_RUN_URL/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"What is the capital of France?","provider":"gemini","rag":true,"top_k":2,"debug":true}'
```

## 12. Local Frontend -> Cloud Backend (CORS Test)

Create `apps/web/.env.local`:

```bash
cat > /home/jeevanism/Documents/Projects/AI-Engineering/multi-model-RAG/apps/web/.env.local <<'EOF'
VITE_API_BASE_URL=https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app
EOF
```

Run frontend:

```bash
cd /home/jeevanism/Documents/Projects/AI-Engineering/multi-model-RAG/apps/web
npm run dev
```

Open `http://localhost:5173` and send a chat request to validate browser CORS against Cloud Run.

## 13. Cleanup / Cost Control (Important for Free Credits)

Pause Cloud SQL instance overnight / when not in use (saves cost):

```bash
gcloud sql instances patch "$SQL_INSTANCE" \
  --project="$PROJECT_ID" \
  --activation-policy=NEVER
```

Start Cloud SQL instance again later (for demo/testing):

```bash
gcloud sql instances patch "$SQL_INSTANCE" \
  --project="$PROJECT_ID" \
  --activation-policy=ALWAYS
```

Tip:
- Cloud Run and Firebase Hosting can stay up.
- DB-backed endpoints (`/ingest/text`, `/evals/*`, `/chat` with `rag=true`) will fail while Cloud SQL is paused.
- Stateless endpoints (`/health`, `/chat`, `/chat/stream` without `rag=true`) can still work.

Delete Cloud SQL instance (destructive):

```bash
gcloud sql instances delete "$SQL_INSTANCE" --project="$PROJECT_ID"
```

Delete Cloud Run service (destructive):

```bash
gcloud run services delete "$CLOUD_RUN_SERVICE" \
  --region="$REGION" \
  --project="$PROJECT_ID"
```

## Notes

- Cloud Run stateless endpoints (`/health`, `/chat`, `/chat/stream`) can work without Cloud SQL.
- DB-backed endpoints (`/ingest/text`, `/evals/*`, `/chat` with `rag=true`) require Cloud SQL + migrations.
- Keep local secrets in `.env.deploy.local` (ignored by git). Do not commit `.env` files.

## 14. Real Gemini + Real Embeddings (Cloud Run)

Create/update the Gemini API key secret:

```bash
printf '%s' "$GEMINI_API_KEY" | gcloud secrets create GEMINI_API_KEY \
  --project="$PROJECT_ID" \
  --replication-policy=automatic \
  --data-file=-
```

If the secret already exists:

```bash
printf '%s' "$GEMINI_API_KEY" | gcloud secrets versions add GEMINI_API_KEY \
  --project="$PROJECT_ID" \
  --data-file=-
```

Important:
- If `Secret Payload cannot be empty`, your `GEMINI_API_KEY` env var is not set in the current shell.
- Cloud Run `--set-secrets ... GEMINI_API_KEY=...:latest` fails if the secret exists but has no versions.

Update Cloud Run for real Gemini generation + real Gemini embeddings:

```bash
gcloud run services update "$CLOUD_RUN_SERVICE" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --add-cloudsql-instances="$INSTANCE_CONN_NAME" \
  --set-secrets="DATABASE_URL=${DB_URL_SECRET}:latest,GEMINI_API_KEY=GEMINI_API_KEY:latest" \
  --update-env-vars="^@^LLM_PROVIDER_MODE=real@EMBEDDING_PROVIDER_MODE=real@EMBEDDING_PROVIDER=gemini@GEMINI_EMBEDDING_MODEL=gemini-embedding-001@CORS_ALLOW_ORIGINS=http://localhost:5173,http://127.0.0.1:5173,https://multi-model-rag-5713b.web.app,https://multi-model-rag-5713b.firebaseapp.com@LOG_LEVEL=info@ENABLE_TRACING=true@DEFAULT_PROVIDER=gemini@DEFAULT_ROUTING_MODE=manual"
```

Rebuild/redeploy note (important):
- `make deploy-cloud-run` can reset env vars back to deploy defaults (stub mode).
- Re-run the `gcloud run services update ... --update-env-vars ...` command after each backend deploy until the deploy script is updated.

`uv.lock` note (important for Docker builds):
- Dockerfile uses `uv sync --frozen --no-dev`, so image deps come from `uv.lock`.
- If you add runtime deps in `pyproject.toml`, run:

```bash
uv lock
```

before rebuilding/deploying, or the image will still miss the new packages.

Cloud proof (real mode):

```bash
curl -s -X POST "$CLOUD_RUN_URL/ingest/text" \
  -H "Content-Type: application/json" \
  -d '{"title":"Real Embedding Cloud Doc 6","content":"Tokyo is the capital of Japan."}'

curl -s -X POST "$CLOUD_RUN_URL/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"What is the capital of Japan?","provider":"gemini","rag":true,"debug":true}'
```

Expected:
- `/ingest/text` returns `"embedding_provider":"gemini"` and `"embedding_model":"gemini-embedding-001"`
- `/chat` returns a real Gemini answer (no `[stub:gemini]` prefix), plus citations and token usage

## 15. Cloud Eval Comparison (Real Mode)

Important:
- `scripts/eval_run.py` defaults to `http://localhost:8000`
- `API_BASE_URL` environment variable is **not** used by the script
- always pass `--api-base-url` explicitly when targeting Cloud Run

Run eval against Cloud Run (save payload JSON):

```bash
uv run python scripts/eval_run.py \
  --api-base-url "https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app" \
  --limit 3 \
  --output .tmp/eval_real.json
```

Run gate against baseline:

```bash
uv run python scripts/eval_gate.py --current .tmp/eval_real.json
```

Persist eval run to DB (shows in Evals dashboard):

```bash
uv run python scripts/eval_run.py \
  --api-base-url "https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app" \
  --limit 3 \
  --persist \
  --output .tmp/eval_real.json
```

Inspect raw eval payload (including per-case errors/results):

```bash
cat .tmp/eval_real.json
```

Useful follow-up:

```bash
curl -s "https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app/evals/runs"
```
