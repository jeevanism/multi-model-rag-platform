# GCP Troubleshooting (Cloud Run / Cloud SQL / gcloud)

This file documents only the Google Cloud deployment and operations errors we hit for this project, with root causes and fixes.

Scope:
- Cloud Run
- Cloud SQL (Postgres)
- Secret Manager
- `gcloud` command usage
- Cloud SQL Proxy / auth for migrations

Related reference:
- `gcp-commands.md` (copy/paste command catalog)

## 1. `make deploy-cloud-run` failed with missing env vars

### Error

```text
Missing required env var: GCP_PROJECT_ID
```

### Root cause

`infra/deploy_cloud_run.sh` requires deploy env vars (`GCP_PROJECT_ID`, `GCP_REGION`, `CLOUD_RUN_SERVICE`, etc.) and they were not exported in the current shell.

### Fix

Use a local ignored env file and source it before deploy:

```bash
source .env.deploy.local
make deploy-cloud-run
```

## 2. `gcloud` project flag parsed incorrectly

### Error / Mistake

Using:

```bash
--project="$multi-model-rag-platform"
```

This is invalid shell expansion and gets parsed incorrectly.

### Root cause

Used a variable expansion on a literal project ID string. Bash interprets `$multi` as a variable.

### Fix

Use either:

```bash
--project="$PROJECT_ID"
```

or a literal string:

```bash
--project="multi-model-rag-platform"
```

## 3. Cloud Run deploy from local shell failed in restricted environment (OAuth / network)

### Error

```text
There was a problem refreshing your current auth tokens ...
Failed to resolve 'oauth2.googleapis.com'
```

### Root cause

Deployment was attempted from a restricted environment without outbound network/DNS or access to local `gcloud` auth state.

### Fix

Run deploy commands directly from the developer machine shell (with network access and authenticated `gcloud`).

## 4. Cloud Run DB-backed endpoints returned `500` after deploy

### Symptoms

- `GET /evals/runs` -> `500 Internal Server Error`
- `GET /evals/runs/{id}` -> `500`
- `POST /chat` with `"rag": true` -> `500`

### Root cause

Cloud Run service was still using the app default `DATABASE_URL` (localhost) because Cloud SQL and secret-backed `DATABASE_URL` had not been configured.

### Proof from logs

```text
connection to server at "127.0.0.1", port 5432 failed: Connection refused
```

### Fix

Provision Cloud SQL, create DB/user, store `DATABASE_URL` in Secret Manager, attach Cloud SQL + secret to Cloud Run, then run migrations.

## 5. `gcloud run services update --update-env-vars` failed with CORS origins

### Error

```text
Bad syntax for dict arg: [http://127.0.0.1:5173]
```

### Root cause

`gcloud` splits `--update-env-vars` on commas, but `CORS_ALLOW_ORIGINS` also contains commas.

### Fix (used successfully)

Use a custom delimiter (`^@^`) so commas can remain inside `CORS_ALLOW_ORIGINS`:

```bash
gcloud run services update "$CLOUD_RUN_SERVICE" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --add-cloudsql-instances="$INSTANCE_CONN_NAME" \
  --set-secrets="DATABASE_URL=${DB_URL_SECRET}:latest" \
  --update-env-vars="^@^CORS_ALLOW_ORIGINS=http://localhost:5173,http://127.0.0.1:5173@LOG_LEVEL=info@ENABLE_TRACING=true@DEFAULT_PROVIDER=gemini@DEFAULT_ROUTING_MODE=manual"
```

Alternative fixes:
- escape commas (`\,`)
- use `--env-vars-file` YAML (cleaner for repeatable ops)

## 6. Cloud SQL instance create failed due to tier/edition mismatch

### Error

```text
Invalid Tier (db-custom-1-3840) for (ENTERPRISE_PLUS) Edition
```

### Root cause

Cloud SQL request defaulted to `ENTERPRISE_PLUS`, which does not allow custom `db-custom-*` tiers.

### Fix

Explicitly set:

```bash
--edition=ENTERPRISE
```

## 7. Cloud SQL instance create failed due to minimum memory constraint

### Error

```text
The total memory (1024MiB) must be at least 3840MiB.
```

### Root cause

The selected CPU/edition combo required at least `3840MB` RAM.

### Fix

Use:

```bash
--cpu=1 --memory=3840MB
```

Working example used:

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

## 8. Cloud SQL creation looked “stuck” (spinner for several seconds)

### Symptom

`gcloud sql instances create ...` spinner appears to hang.

### Root cause

Cloud SQL instance provisioning is slow compared to Cloud Run and often takes several minutes.

### Fix / Validation

This is normal. Check progress in another terminal:

```bash
gcloud sql instances list --project="$PROJECT_ID"
gcloud sql operations list --instance="$SQL_INSTANCE" --project="$PROJECT_ID"
```

Look for:
- `STATE=RUNNABLE` (instance ready)

## 9. `gcloud sql connect` failed because Cloud SQL Proxy was missing

### Error

```text
Cloud SQL Proxy (v2) couldn't be found in PATH
```

### Root cause

`gcloud sql connect` depends on the Cloud SQL Proxy binary being available.

### Fix

Install `cloud-sql-proxy` manually when the packaged `gcloud components install` path is unavailable:

```bash
curl -o cloud-sql-proxy https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.21.1/cloud-sql-proxy.linux.amd64
chmod +x cloud-sql-proxy
sudo mv cloud-sql-proxy /usr/local/bin/
```

## 10. `gcloud components install cloud-sql-proxy` unavailable

### Error

```text
The cloud-sql-proxy component(s) is unavailable through the packaging system you are currently using.
```

### Root cause

The installed `gcloud` CLI came from a packaging mechanism that does not support that optional component.

### Fix

Install the standalone Cloud SQL Proxy binary manually (see issue #9).

## 11. `gcloud sql connect` failed due to missing ADC (Application Default Credentials)

### Error

```text
failed to create default credentials: credentials: could not find default credentials
```

### Root cause

Cloud SQL Proxy uses ADC. `gcloud auth login` alone is not enough.

### Fix

Run:

```bash
gcloud auth application-default login
gcloud auth application-default set-quota-project "$PROJECT_ID"
```

Then retry:

```bash
gcloud sql connect "$SQL_INSTANCE" \
  --project="$PROJECT_ID" \
  --user="$DB_USER" \
  --database="$DB_NAME"
```

## 12. Cloud Run logs confirm DB misconfiguration quickly

### Symptom

Cloud endpoints return `500`, but root cause is unclear from `curl` alone.

### Fix / Best practice

Check Cloud Run logs immediately:

```bash
gcloud run services logs read "$CLOUD_RUN_SERVICE" \
  --region="$REGION" \
  --project="$PROJECT_ID" \
  --limit=100
```

This revealed the exact `sqlalchemy`/`psycopg` DB connection errors.

## 13. Secret Manager / Cloud Run service account permissions (potential issue)

### Risk

Cloud Run can fail to read `DATABASE_URL` from Secret Manager if the service account lacks `secretAccessor`.

### Fix

Find Cloud Run service account and grant access:

```bash
export RUN_SA="$(gcloud run services describe "$CLOUD_RUN_SERVICE" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --format='value(spec.template.spec.serviceAccountName)')"

gcloud secrets add-iam-policy-binding "$DB_URL_SECRET" \
  --project="$PROJECT_ID" \
  --member="serviceAccount:${RUN_SA}" \
  --role="roles/secretmanager.secretAccessor"
```

## 14. Security note: Secret values were printed in terminal history/output

### What happened

`DB_PASSWORD` and full `DATABASE_URL` were echoed while setting up Cloud SQL and Secret Manager.

### Risk

Credentials may be exposed in shell history, terminal scrollback, or shared logs.

### Recommended fix (do this after setup is stable)

1. Rotate the Cloud SQL password:

```bash
gcloud sql users set-password "$DB_USER" \
  --instance="$SQL_INSTANCE" \
  --project="$PROJECT_ID" \
  --password="NEW_STRONG_PASSWORD"
```

2. Update Secret Manager with a new `DATABASE_URL` version.
3. Ensure Cloud Run uses `:latest` (already configured).
4. Avoid `echo "$DATABASE_URL"` in future.

## 15. `GEMINI_API_KEY` secret create/update failed because payload was empty

### Error

```text
INVALID_ARGUMENT: Secret Payload cannot be empty.
```

### Root cause

The `GEMINI_API_KEY` environment variable was not set in the current shell, so:

```bash
printf '%s' "$GEMINI_API_KEY"
```

produced an empty payload.

### Fix

Set/export the key in the current shell first, then add the secret version:

```bash
export GEMINI_API_KEY='...'
printf '%s' "$GEMINI_API_KEY" | gcloud secrets versions add GEMINI_API_KEY \
  --project="$PROJECT_ID" \
  --data-file=-
```

## 16. Cloud Run update failed because secret existed but had no versions

### Error

```text
Secret .../secrets/GEMINI_API_KEY/versions/latest was not found
```

### Root cause

The secret resource `GEMINI_API_KEY` existed, but no valid version had been created yet.

### Fix

Add a valid secret version first, then rerun `gcloud run services update`:

```bash
printf '%s' "$GEMINI_API_KEY" | gcloud secrets versions add GEMINI_API_KEY \
  --project="$PROJECT_ID" \
  --data-file=-
```

## 17. `make deploy-cloud-run` reverted Cloud Run runtime env vars back to stub mode

### Symptom

After a successful backend image deploy:
- `/ingest/text` still returned `"embedding_provider":"stub"`
- `/chat` still returned `[stub:gemini]`

### Root cause

The deploy script applies its own env vars/defaults during deploy, which reset runtime settings to stub mode.

### Fix / Best practice

After each `make deploy-cloud-run`, re-run the Cloud Run update command that sets:
- `LLM_PROVIDER_MODE=real`
- `EMBEDDING_PROVIDER_MODE=real`
- `EMBEDDING_PROVIDER=gemini`
- `GEMINI_EMBEDDING_MODEL=gemini-embedding-001`
- Firebase + localhost CORS origins

## 18. Cloud Run real mode returned `500` because `google-genai` was missing from the image

### Error (from Cloud Run logs)

```text
ModuleNotFoundError: No module named 'google'
RuntimeError: google-genai package is required for real Gemini embeddings ...
```

### Root cause

The Docker build uses:

```bash
uv sync --frozen --no-dev
```

so image dependencies come from `uv.lock`.

`pyproject.toml` was updated, but `uv.lock` was stale, so the image still did not include `google-genai`.

### Fix

1. Ensure `google-genai` / `openai` are in runtime `[project].dependencies`
2. Regenerate lockfile:

```bash
uv lock
```

3. Rebuild/redeploy backend image:

```bash
source .env.deploy.local
make deploy-cloud-run
```

4. Re-apply Cloud Run real-mode env vars (deploy script reset)

## 19. Cloud Run `500` debugging for real-mode rollout (working sequence)

### Best debugging order

1. Confirm deployed image revision changed (`make deploy-cloud-run` output)
2. Re-apply runtime env vars with `gcloud run services update ... --update-env-vars ...`
3. Run cloud endpoint smoke calls (`/ingest/text`, `/chat` with `rag=true`)
4. Read Cloud Run logs immediately:

```bash
gcloud run services logs read "$CLOUD_RUN_SERVICE" \
  --region="$REGION" \
  --project="$PROJECT_ID" \
  --limit=100
```

This sequence quickly distinguishes:
- stub mode still active
- missing secret/version
- missing SDK in image
- API/runtime errors in real provider/embedding paths

## Quick Cloud Sanity Checklist

Use this order when cloud features fail:

1. `gcloud run services logs read ...`
2. Confirm Cloud SQL instance is `RUNNABLE`
3. Confirm Cloud Run has `--add-cloudsql-instances`
4. Confirm `DATABASE_URL` secret is attached
5. Confirm Secret Manager IAM (`secretAccessor`) for Cloud Run service account
6. Run migrations on Cloud SQL
7. Retest `/evals/runs`, `/ingest/text`, and `rag=true`
