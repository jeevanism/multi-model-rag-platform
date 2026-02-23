# GCP Setup Steps (Chronological Runbook)

This file is a chronological, copy/paste-friendly record of the Google Cloud setup steps we performed for this project.

Purpose:
- Track exactly what was done
- Reproduce setup later
- Continue appending as we complete more cloud work

Related docs:
- `gcp-commands.md` (command reference)
- `troubleshooting.md` (GCP-only errors and fixes)

## Current Status (As Of Latest Update)

- Cloud Run backend deployed ✅
- Cloud SQL (Postgres 16) instance created ✅
- Cloud SQL database + user password configured ✅
- Secret Manager `DATABASE_URL` created ✅
- Cloud Run updated with Cloud SQL + secret + env vars ✅
- Cloud SQL migrations (`001`–`004`) applied ✅
- `pgvector` extension enabled in Cloud SQL ✅
- DB-backed cloud endpoint retest: pending (next step) ⏳

## 0. Set Active Project

```bash
gcloud config set project multi-model-rag-platform
```

Note:
- The “environment tag” message shown by org policy was informational and did not block commands.

## 1. Define Shell Variables (Current Session)

```bash
export PROJECT_ID="multi-model-rag-platform"
export REGION="us-central1"
export SQL_INSTANCE="multi-model-rag-pg"
export DB_NAME="multimodel_rag"
export DB_USER="postgres"
export DB_PASSWORD="REPLACE_WITH_A_REAL_STRONG_PASSWORD"
export CLOUD_RUN_SERVICE="multi-model-rag-api"
export DB_URL_SECRET="multi-model-rag-database-url"
```

Important:
- If you `exit` the shell, these `export`s are lost and must be set again.

## 2. Enable Required GCP APIs

```bash
gcloud services enable \
  run.googleapis.com \
  sqladmin.googleapis.com \
  secretmanager.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  --project="$PROJECT_ID"
```

Observed result:
- Operation completed successfully ✅

## 3. Backend Cloud Run Deploy (Using Local Ignored Env File)

Create local deploy env file (git-ignored):

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

Deploy:

```bash
source .env.deploy.local
make deploy-cloud-run
```

Observed result:
- Cloud Build succeeded ✅
- Cloud Run deployed ✅
- Service URL available ✅
- `/health` post-deploy smoke passed ✅

Deployed backend URL:

```text
https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app
```

## 4. Initial Cloud Run Verification (Before Cloud SQL)

Worked:

```bash
curl -s https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app/health

curl -s -X POST https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"hello from cloud","provider":"gemini"}'

curl -N -X POST https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message":"hello stream","provider":"gemini"}'
```

Failed (expected before DB setup):

```bash
curl -s https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app/evals/runs
curl -s https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app/evals/runs/2
curl -s -X POST https://multi-model-rag-api-ozzmnn5qja-uc.a.run.app/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"What is the capital of France?","provider":"gemini","rag":true,"top_k":2,"debug":true}'
```

Observed root cause:
- Cloud Run logs showed DB connection to `127.0.0.1:5432` (`localhost`) failing.

## 5. Cloud Run Logs Check (Used for Diagnosis)

```bash
gcloud run services logs read multi-model-rag-api \
  --region=us-central1 \
  --project=multi-model-rag-platform \
  --limit=100
```

Observed result:
- `sqlalchemy` / `psycopg` `OperationalError`
- connection to `127.0.0.1:5432` refused ✅ (confirmed DB config missing in cloud)

## 6. Create Cloud SQL Postgres Instance

Attempt that failed (defaulted to `ENTERPRISE_PLUS` + invalid custom tier):

```bash
gcloud sql instances create "$SQL_INSTANCE" \
  --project="$PROJECT_ID" \
  --database-version=POSTGRES_16 \
  --region="$REGION" \
  --cpu=1 \
  --memory=3840MB \
  --storage-size=20GB
```

Working command (explicit `ENTERPRISE`):

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

Observed result:
- Instance created ✅
- Status `RUNNABLE` ✅

## 7. Set Cloud SQL DB Password and Create Database

Set a real password in shell:

```bash
export DB_PASSWORD='REPLACE_WITH_REAL_STRONG_PASSWORD'
echo "${#DB_PASSWORD}"
```

Set `postgres` password:

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

Observed result:
- User password updated ✅
- Database `multimodel_rag` created ✅

## 8. Get Cloud SQL Connection Name

```bash
export INSTANCE_CONN_NAME="$(gcloud sql instances describe "$SQL_INSTANCE" \
  --project="$PROJECT_ID" \
  --format='value(connectionName)')"
echo "$INSTANCE_CONN_NAME"
```

Observed result:

```text
multi-model-rag-platform:us-central1:multi-model-rag-pg
```

## 9. Create `DATABASE_URL` Secret in Secret Manager

Build Cloud SQL socket-based SQLAlchemy URL:

```bash
export DATABASE_URL="postgresql+psycopg://${DB_USER}:${DB_PASSWORD}@/${DB_NAME}?host=/cloudsql/${INSTANCE_CONN_NAME}"
```

Create secret:

```bash
printf '%s' "$DATABASE_URL" | gcloud secrets create "$DB_URL_SECRET" \
  --project="$PROJECT_ID" \
  --replication-policy=automatic \
  --data-file=-
```

Observed result:
- Secret created ✅
- Version `1` created ✅

Security note:
- Avoid echoing the full `DATABASE_URL` in future (contains password).

## 10. Grant Cloud Run Service Account Access to Secret

Get Cloud Run service account:

```bash
export RUN_SA="$(gcloud run services describe "$CLOUD_RUN_SERVICE" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --format='value(spec.template.spec.serviceAccountName)')"
echo "$RUN_SA"
```

Observed result:

```text
333316966566-compute@developer.gserviceaccount.com
```

Grant secret access:

```bash
gcloud secrets add-iam-policy-binding "$DB_URL_SECRET" \
  --project="$PROJECT_ID" \
  --member="serviceAccount:${RUN_SA}" \
  --role="roles/secretmanager.secretAccessor"
```

Observed result:
- IAM binding updated ✅

## 11. Update Cloud Run With Cloud SQL + Secret + Env Vars

First attempt failed because `CORS_ALLOW_ORIGINS` contains commas and `gcloud` parses commas as env-var separators.

Working command (custom delimiter `^@^`):

```bash
gcloud run services update "$CLOUD_RUN_SERVICE" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --add-cloudsql-instances="$INSTANCE_CONN_NAME" \
  --set-secrets="DATABASE_URL=${DB_URL_SECRET}:latest" \
  --update-env-vars="^@^CORS_ALLOW_ORIGINS=http://localhost:5173,http://127.0.0.1:5173@LOG_LEVEL=info@ENABLE_TRACING=true@DEFAULT_PROVIDER=gemini@DEFAULT_ROUTING_MODE=manual"
```

Observed result:
- New revision deployed ✅
- Traffic routed to new revision ✅

## 12. Install Cloud SQL Proxy Binary (Required for `gcloud sql connect`)

`gcloud components install cloud-sql-proxy` was unavailable in the packaged CLI, so we installed manually:

```bash
curl -o cloud-sql-proxy https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.21.1/cloud-sql-proxy.linux.amd64
chmod +x cloud-sql-proxy
sudo mv cloud-sql-proxy /usr/local/bin/
```

Observed result:
- `gcloud sql connect` could find the proxy ✅

## 13. Configure ADC (Application Default Credentials) for Cloud SQL Proxy

`gcloud sql connect` initially failed because ADC was missing.

Fix:

```bash
gcloud auth application-default login
gcloud auth application-default set-quota-project "$PROJECT_ID"
```

Observed result:
- ADC configured ✅

## 14. Connect to Cloud SQL and Run Migrations

Connect:

```bash
gcloud sql connect "$SQL_INSTANCE" \
  --project="$PROJECT_ID" \
  --user="$DB_USER" \
  --database="$DB_NAME"
```

At `psql` prompt, run:

```sql
\i /home/jeevanism/Documents/Projects/AI-Engineering/multi-model-RAG/migrations/001_init.sql
\i /home/jeevanism/Documents/Projects/AI-Engineering/multi-model-RAG/migrations/002_rag_schema.sql
\i /home/jeevanism/Documents/Projects/AI-Engineering/multi-model-RAG/migrations/003_evals_schema.sql
\i /home/jeevanism/Documents/Projects/AI-Engineering/multi-model-RAG/migrations/004_eval_scores.sql
```

Verification queries:

```sql
SELECT extname FROM pg_extension WHERE extname='vector';
SELECT COUNT(*) FROM documents;
SELECT COUNT(*) FROM eval_run;
```

Observed result:
- `vector` extension exists ✅
- tables created ✅
- counts start at `0` ✅

## 15. Next Steps (To Append After Completion)

Pending cloud validation after migrations:

1. Re-test DB-backed endpoints on Cloud Run:
   - `GET /evals/runs`
   - `POST /ingest/text`
   - `POST /chat` with `rag=true`
2. Optional: local frontend -> Cloud Run backend CORS browser test
3. Optional: run cloud evals and confirm `/evals/runs` shows persisted rows

## 16. Security Follow-Up (Recommended)

The DB password and full `DATABASE_URL` were exposed in terminal output while setting up cloud DB.

After cloud validation is complete:

1. Rotate Cloud SQL password
2. Add new Secret Manager version for `DATABASE_URL`
3. Ensure Cloud Run uses `:latest` (already configured)

## 17. Update Policy For This File

When we complete a new cloud step:
- add the exact command(s)
- add the observed result
- add any errors/fixes encountered

This file should remain chronological and operational (not a general reference).
