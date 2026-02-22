Below is a **step-by-step iteration guide** you can save as `PROJECT_PLAN.md`. It includes **exact installs** for a Linux dev machine and a **build order** that matches the 5-week roadmap (without overbuilding).

````markdown
# PROJECT_PLAN.md — GenAI Planning Assistant (Multi-model RAG on GCP)

## Goal
Build a production-oriented GenAI system that demonstrates:
- FastAPI backend (streaming chat)
- Multi-model provider abstraction (Gemini + OpenAI)
- RAG with Postgres + pgvector
- Evaluation harness (correctness/groundedness/hallucination/latency/cost)
- Observability (structured logs + tracing)
- React + TypeScript UI
- Cloud deployment (Cloud Run + Cloud SQL + GCS + Secret Manager)

---

## 0) Local prerequisites (Linux)

### OS assumptions
- Ubuntu/Debian-based (commands may differ slightly on Fedora/Arch)

### 0.1 System packages
```bash
sudo apt update
sudo apt install -y \
  git curl wget unzip ca-certificates gnupg lsb-release \
  build-essential make \
  jq ripgrep \
  python3 python3-venv \
  postgresql-client \
  nodejs npm
````

> Note: Ubuntu’s Node may be old. Prefer Node via `nvm` below.

### 0.2 Install Docker + Docker Compose

```bash
# Docker official repo
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# allow running docker without sudo
sudo usermod -aG docker $USER
newgrp docker
docker --version
docker compose version
```

### 0.3 Install Node.js (recommended: nvm + Node 20)

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
# restart shell, then:
nvm install 20
nvm use 20
node -v
npm -v
```

### 0.4 Install Python tooling (required: uv)

`uv` is the standard tool for Python venv + dependency management in this project.

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
# restart shell
uv --version
python3 --version
```

### 0.5 Install Google Cloud SDK (for deployment later)

```bash
curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg \
  | sudo gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg

echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" \
  | sudo tee /etc/apt/sources.list.d/google-cloud-sdk.list

sudo apt update
sudo apt install -y google-cloud-cli
gcloud version
```

### 0.6 Optional but recommended

* `direnv` (auto-load env vars per project)

```bash
sudo apt install -y direnv
echo 'eval "$(direnv hook bash)"' >> ~/.bashrc
# restart shell
```

---

## 1) Repo scaffold (Day 1)

### 1.1 Create repo structure

```bash
mkdir -p genai-planning-assistant/{apps/{api,web},packages/{llm,rag,evals,observability},infra, datasets, docs}
cd genai-planning-assistant
git init
```

### 1.2 Add root files

Create:

* `README.md`
* `PROJECT_PLAN.md` (this file)
* `.gitignore`
* `.env.example`

---

## 2) Local development environment (Day 1–2)

### 2.1 Docker compose for Postgres + pgvector

Create `docker-compose.yml` (minimal):

* Postgres 16
* pgvector enabled
* expose 5432
* persistent volume

You’ll use this DB locally for:

* embeddings + chunks
* eval results
* runs + metrics

### 2.2 Start DB

```bash
docker compose up -d
psql "postgresql://postgres:postgres@localhost:5432/postgres" -c "SELECT 1;"
```

---

## 3) Backend iteration 1: FastAPI skeleton + health (Day 2)

### 3.1 Create Python project (apps/api)

```bash
cd apps/api
uv venv
source .venv/bin/activate

uv pip install \
  fastapi uvicorn[standard] pydantic \
  sqlalchemy psycopg[binary] \
  python-dotenv \
  httpx \
  tenacity \
  orjson
```

Recommended dev deps:

```bash
uv pip install pytest pytest-asyncio ruff black mypy
```

### 3.2 Minimal app

* `main.py` with:

  * `GET /health`
  * structured JSON logging (basic)
* run:

```bash
uvicorn main:app --reload --port 8000
curl http://localhost:8000/health
```

---

## 4) Backend iteration 2: LLM provider abstraction (Week 1)

### 4.1 Install provider SDKs

```bash
uv pip install openai google-genai
```

### 4.2 Implement `packages/llm`

Create:

* `llm/base.py` — `LLMProvider` interface
* `llm/providers/openai.py`
* `llm/providers/gemini.py`
* `llm/router.py` — manual vs auto routing (start manual)
* `llm/types.py` — unified response schema

### 4.3 Add streaming endpoint (SSE)

Implement:

* `POST /chat` that streams tokens via Server-Sent Events

---

## 5) RAG iteration 1: ingestion + chunk store (Week 2)

### 5.1 Add libraries

```bash
uv pip install tiktoken
```

(If you want a simple chunker, tiktoken helps token-based chunking. Keep it minimal.)

### 5.2 DB tables for RAG

Create migrations (simple SQL files are fine initially):

* `documents`
* `chunks`
* `embeddings` (vector column via pgvector)

### 5.3 Ingestion pipeline

Implement:

* upload doc (local file or simple endpoint)
* chunk text
* store chunks
* generate embeddings (use Gemini or OpenAI embeddings)
* store vectors

### 5.4 Retrieval

Implement:

* similarity search top-k with pgvector
* return chunk text + metadata
* feed into prompt template with citations

---

## 6) Eval framework iteration (Week 3 — critical)

### 6.1 Create dataset

Create `datasets/eval_set.jsonl` with:

* 30–50 cases
* mix of: answerable RAG + refusal cases

### 6.2 Implement eval runner

Create `packages/evals`:

* `run.py` loads dataset, calls `/chat`, captures:

  * answer, citations, retrieved chunks
  * latency breakdown
  * tokens & cost
* `judge.py` runs judge prompts:

  * groundedness
  * fact match
  * hallucination
* `aggregate.py` produces summary metrics
* `gate.py` compares summary vs baseline

### 6.3 Store eval results in DB

Create DB tables:

* `eval_run`
* `eval_run_case`
* `eval_baseline` (optional)

---

## 7) Observability iteration (Week 4)

### 7.1 Structured logging

Log per request:

* request_id
* model provider/name
* retrieval_ms, llm_ms, total_ms
* tokens, cost
* error type

### 7.2 Tracing (OpenTelemetry)

Add:

```bash
uv pip install opentelemetry-api opentelemetry-sdk opentelemetry-instrumentation-fastapi
```

Instrument:

* request span
* retrieval span
* llm span

(Keep it minimal. You can expand later.)

---

## 8) Frontend iteration (Week 5)

### 8.1 Create React app (apps/web)

```bash
cd ../web
npm create vite@latest . -- --template react-ts
npm install
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

### 8.2 Build UI

Pages:

* Chat (streaming)
* Eval dashboard (table)
* Ingest page (upload)

Features:

* model selector (manual)
* show citations
* show latency + cost

---

## 9) Cloud deployment (Week 4–5)

### 9.1 GCP resources (minimum)

* Cloud Run (API)
* Cloud SQL Postgres (with pgvector)
* GCS bucket for docs
* Secret Manager for keys
* Cloud Logging + Monitoring

### 9.2 Deploy backend

* Dockerfile
* Cloud Run deploy command
* set env vars for DB + secrets

### 9.3 Deploy frontend

* Cloud Run static or Firebase Hosting (either is fine)
* Keep it simple: serve built static files

---

## 10) Definition of Done (DoD)

Project is “portfolio-ready” when:

* RAG works with citations
* Gemini + OpenAI both work behind a provider abstraction
* `/chat` streams responses
* Evaluation suite runs and outputs summary table
* Regression gating exists (baseline + thresholds)
* Logs include latency + token/cost
* Deployed URL exists (even if access-controlled)
* README contains architecture + trade-offs + screenshots

---

## 11) What to cut (do NOT build)

* Kubernetes / GKE
* Terraform (optional later)
* 4+ model providers (2 is enough)
* Full auth platform (simple JWT/env token is enough)
* Multi-agent orchestration engine
* Fine-tuning / training pipelines
* Complex caching & queues (unless absolutely needed)

---

## 12) Environment variables (.env.example)

* DATABASE_URL=
* GEMINI_API_KEY= (or Vertex auth config)
* OPENAI_API_KEY=
* DEFAULT_ROUTING_MODE=manual
* DEFAULT_PROVIDER=gemini
* LOG_LEVEL=info

```

---

## What you should do first (today)

1) Install: **Docker + uv + nvm Node 20 + gcloud**  
2) Scaffold repo structure + `docker-compose.yml` for Postgres/pgvector  
3) Bring up DB, create FastAPI `/health`  
4) Implement **LLM provider interface** + Gemini/OpenAI “hello world” generate  
5) Only then move to RAG

If you want, paste your Linux distro (Ubuntu version, etc.) and I’ll adjust the install commands to match exactly.
```
