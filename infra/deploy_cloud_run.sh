#!/usr/bin/env bash
set -euo pipefail

# Deploys the API container to Cloud Run using gcloud.
# Required env vars:
#   GCP_PROJECT_ID
#   GCP_REGION
# Optional env vars:
#   CLOUD_RUN_SERVICE (default: multi-model-rag-api)
#   IMAGE_REPO (default: multi-model-rag-api)
#   IMAGE_URI (overrides gcr.io-derived image URI)
#   ALLOW_UNAUTHENTICATED (default: true)
#   CLOUDSQL_INSTANCE (Cloud SQL instance connection name)
#   SECRET_ENV_VARS (comma-separated KEY=secret:version pairs for --set-secrets)
#   SERVICE_ACCOUNT
#   MIN_INSTANCES, MAX_INSTANCES, MEMORY, CPU, INGRESS
#   DATABASE_URL, LOG_LEVEL, ENABLE_TRACING, DEFAULT_PROVIDER, DEFAULT_ROUTING_MODE
#   POST_DEPLOY_SMOKE (default: true) - curl /health after deploy when curl is available

require_var() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    echo "Missing required env var: $name" >&2
    exit 1
  fi
}

require_var "GCP_PROJECT_ID"
require_var "GCP_REGION"

if ! command -v gcloud >/dev/null 2>&1; then
  echo "gcloud CLI is required but not installed." >&2
  exit 1
fi

CLOUD_RUN_SERVICE="${CLOUD_RUN_SERVICE:-multi-model-rag-api}"
IMAGE_REPO="${IMAGE_REPO:-multi-model-rag-api}"
ALLOW_UNAUTHENTICATED="${ALLOW_UNAUTHENTICATED:-true}"
POST_DEPLOY_SMOKE="${POST_DEPLOY_SMOKE:-true}"
IMAGE_URI="${IMAGE_URI:-gcr.io/${GCP_PROJECT_ID}/${IMAGE_REPO}}"

echo "Building image with Cloud Build: ${IMAGE_URI}"
gcloud builds submit --tag "${IMAGE_URI}" .

DEPLOY_ARGS=(
  run deploy "${CLOUD_RUN_SERVICE}"
  --project "${GCP_PROJECT_ID}"
  --region "${GCP_REGION}"
  --image "${IMAGE_URI}"
  --platform managed
  --port 8080
)

if [[ "${ALLOW_UNAUTHENTICATED}" == "true" ]]; then
  DEPLOY_ARGS+=(--allow-unauthenticated)
fi

if [[ -n "${CLOUDSQL_INSTANCE:-}" ]]; then
  DEPLOY_ARGS+=(--add-cloudsql-instances "${CLOUDSQL_INSTANCE}")
fi

if [[ -n "${SERVICE_ACCOUNT:-}" ]]; then
  DEPLOY_ARGS+=(--service-account "${SERVICE_ACCOUNT}")
fi

if [[ -n "${MIN_INSTANCES:-}" ]]; then
  DEPLOY_ARGS+=(--min-instances "${MIN_INSTANCES}")
fi

if [[ -n "${MAX_INSTANCES:-}" ]]; then
  DEPLOY_ARGS+=(--max-instances "${MAX_INSTANCES}")
fi

if [[ -n "${MEMORY:-}" ]]; then
  DEPLOY_ARGS+=(--memory "${MEMORY}")
fi

if [[ -n "${CPU:-}" ]]; then
  DEPLOY_ARGS+=(--cpu "${CPU}")
fi

if [[ -n "${INGRESS:-}" ]]; then
  DEPLOY_ARGS+=(--ingress "${INGRESS}")
fi

ENV_VARS=()
for key in DATABASE_URL LOG_LEVEL ENABLE_TRACING DEFAULT_PROVIDER DEFAULT_ROUTING_MODE; do
  if [[ -n "${!key:-}" ]]; then
    ENV_VARS+=("${key}=${!key}")
  fi
done

if (( ${#ENV_VARS[@]} > 0 )); then
  IFS=,
  DEPLOY_ARGS+=(--set-env-vars "${ENV_VARS[*]}")
  unset IFS
fi

if [[ -n "${SECRET_ENV_VARS:-}" ]]; then
  DEPLOY_ARGS+=(--set-secrets "${SECRET_ENV_VARS}")
fi

echo "Deploying Cloud Run service: ${CLOUD_RUN_SERVICE}"
gcloud "${DEPLOY_ARGS[@]}"

SERVICE_URL="$(
  gcloud run services describe "${CLOUD_RUN_SERVICE}" \
    --project "${GCP_PROJECT_ID}" \
    --region "${GCP_REGION}" \
    --format='value(status.url)'
)"

echo "Deployment complete."
echo "Service URL: ${SERVICE_URL}"

if [[ "${POST_DEPLOY_SMOKE}" == "true" ]]; then
  if command -v curl >/dev/null 2>&1; then
    echo "Running post-deploy smoke check: ${SERVICE_URL}/health"
    curl --fail --silent --show-error "${SERVICE_URL}/health"
    echo
  else
    echo "Skipping post-deploy smoke check (curl not installed)."
  fi
fi
