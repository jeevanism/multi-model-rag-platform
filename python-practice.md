# Python Practices and Standards (Project-Aligned, Production-Ready)

## Role and Objective
Build and maintain a modern, production-ready Python backend that is:
- correct
- observable
- testable
- secure by default
- easy to change in small iterations

This document is the Python engineering standard for this repository and should align with the actual project setup.

## Core Technology Standard (for this project)
- **Package Manager & Environment:** `uv` (Astral) for venv + dependency management
- **Framework:** FastAPI
- **Validation & Serialization:** Pydantic v2 (add `pydantic-settings` when config expands)
- **Database ORM:** SQLAlchemy 2.0
- **Migrations:** Start with SQL migration files (current repo), move to Alembic when schema complexity grows
- **Linting & Formatting:** `ruff` (lint + format)
- **Type Checking:** `mypy` (strict enough to catch interface mistakes)
- **Testing:** `pytest`

## 1. Package and Dependency Management (`uv` Required)
- Use `uv` for all Python package installation and environment setup.
- Prefer:
  - `uv venv`
  - `source .venv/bin/activate`
  - `uv pip install -e ".[dev]"`
- For adding dependencies in mature iterations, prefer:
  - `uv add <pkg>`
  - `uv add --dev <pkg>`
- Use `uv run <command>` for local commands when not activating the venv.
- Keep dependency definitions in `pyproject.toml`.

Rules:
- Do not use ad hoc `pip install ...` commands in docs/scripts/CI unless there is a specific exception.
- Do not maintain duplicate dependency lists (`requirements.txt`) unless required for deployment tooling.

## 2. Project Structure (Pragmatic, Evolves with Size)
Early-stage projects can start simple. As features grow, move toward feature-based modules.

Current repo direction:
- `apps/api/`: FastAPI app and API-specific wiring
- `packages/`: shared/domain packages (`llm`, `rag`, `evals`, `observability`)
- `tests/`: unit and integration tests
- `migrations/`: SQL migration files (initially)

Recommended API structure as complexity grows:
- `apps/api/main.py`: app factory / FastAPI app setup
- `apps/api/settings.py`: env-backed settings
- `apps/api/db.py`: engine/session lifecycle
- `apps/api/routes/`: API route modules
- `apps/api/dependencies.py`: FastAPI dependencies (`Depends`)
- `apps/api/services/`: business orchestration logic
- `apps/api/schemas/`: request/response models

Guideline:
- Prefer feature/domain grouping once multiple endpoints/features exist.
- Do not over-engineer module layout before there are real features.

## 3. Architecture and API Rules
### Service Layer Boundary
- Keep routers thin.
- Routers should:
  - validate input
  - call service/domain functions
  - map exceptions to HTTP responses
- Business logic should live outside route handlers.

### Pydantic Boundaries
- Do not return ORM models directly from API endpoints.
- Define request/response schemas explicitly.
- Keep transport schemas separate from persistence models.

### Dependency Injection
- Use FastAPI `Depends()` for framework-level dependencies:
  - DB sessions
  - auth/context
  - request-scoped utilities
- Use normal Python function parameters for internal service calls.

### Async vs Sync (Project-Realistic Standard)
- Prefer `async def` endpoints for new API routes.
- Avoid blocking I/O in request handlers.
- SQLAlchemy may be sync or async depending on the current iteration and complexity.
- If sync DB access is used in API paths, keep usage simple and be explicit about the tradeoff.
- Migrations, scripts, and tests may use sync code even if the API is async.

## 4. Configuration and Secrets
- Read configuration from environment variables.
- Centralize config in a single settings module.
- Provide `.env.example` with non-secret placeholders only.
- Never hardcode API keys, tokens, or DB credentials in code.
- Define safe defaults only for local development.

## 5. Database and Migration Practices
- Use SQLAlchemy 2.0 style APIs.
- Keep session lifecycle explicit and short-lived.
- Add migrations for every schema change (even early SQL files).
- Migration files must be idempotent where possible (e.g., `IF NOT EXISTS`).
- Do not mix schema changes with unrelated feature logic in one commit.

When to adopt Alembic:
- Multiple tables + repeated schema changes
- Team collaboration on concurrent branches
- Need for upgrade/downgrade traceability

## 6. Testing Standards
Test strategy must include:
- **Unit tests** for pure logic and adapters (fast)
- **Integration tests** for DB/API wiring (real services where needed)
- **Smoke tests** for proof commands (curl / scripts)

Rules:
- Every bug fix should include a regression test when practical.
- Tests should be deterministic and isolated.
- Integration tests should skip gracefully when required local services are unavailable (unless CI guarantees them).
- Name tests for behavior, not implementation details.

## 7. Typing Standards
- Type hints are required for public functions, methods, and module interfaces.
- Prefer precise return types over `Any`.
- Avoid silent `Any` propagation in service boundaries.
- Use typed DTOs/schemas for cross-module interfaces.

Practical rule:
- It is acceptable to temporarily relax typing in exploratory code, but tighten it before merge.

## 8. Error Handling and Logging
- Fail with clear, actionable errors.
- Do not swallow exceptions silently.
- Add structured logs at system boundaries:
  - incoming request
  - DB call
  - external API call (LLM/provider)
  - error path
- Never log secrets or full sensitive payloads.

For this project, log at least:
- request id (when available)
- provider/model
- latency
- error type

## 9. Performance and Reliability
- Set timeouts for all outbound HTTP requests.
- Add retries only for transient failures and only where safe.
- Measure latency before optimizing.
- Avoid premature caching/queues.
- Keep critical paths observable (timings, error rates).

## 10. CI and Local Quality Gates
Required local checks before push:
- `uv run ruff check .`
- `uv run ruff format --check .` (or run `uv run ruff format .` before)
- `uv run mypy apps/api tests`
- `uv run pytest -q`

CI should enforce the same checks and use `uv`.

## 11. Docker and Runtime Packaging
- Prefer multi-stage Dockerfiles for production images.
- Use `uv` in build steps for dependency installation.
- Keep runtime image minimal.
- Run the app as a non-root user when practical.
- Configure health checks and graceful shutdown behavior.

## 12. Repo-Specific Notes (Multi-Model RAG Project)
- Keep provider integrations behind a unified interface (`packages/llm`)
- Keep RAG retrieval and prompt construction testable without HTTP endpoints
- Track eval logic and metric aggregation as first-class code (`packages/evals`)
- Add observability early enough to diagnose retrieval and LLM latency issues

## 13. Anti-Patterns to Avoid
- Fat route handlers with embedded business logic
- Returning ORM objects directly in API responses
- Hidden global state for request-scoped data
- Unbounded retries/timeouts on external calls
- Mixing refactors with behavior changes in one commit
- Introducing frameworks/tools not justified by current project needs
