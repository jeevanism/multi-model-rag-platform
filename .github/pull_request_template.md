## Summary
- What changed?
- Why was this change needed?

## Scope
- [ ] Small/isolated change
- [ ] Feature change
- [ ] Refactor
- [ ] Bug fix
- [ ] Docs only

## Validation
- [ ] Local proof command(s) run (curl / script / smoke test)
- [ ] Behavior verified manually where applicable
- [ ] Added or updated tests for changed behavior

## Quality Gates (from `python-practice.md`)
- [ ] `uv run ruff check .`
- [ ] `uv run ruff format --check .` (or formatted with `uv run ruff format .`)
- [ ] `uv run mypy apps/api tests`
- [ ] `uv run pytest -q`

## Database / API Impact
- [ ] No schema changes
- [ ] Schema changes included with migration
- [ ] API contract changed (request/response)
- [ ] No API contract changes

## Risk Review
- [ ] No secrets added to code/logs
- [ ] Error handling updated where needed
- [ ] Logging is sufficient for debugging changed behavior

## Notes
- Follow-up work / limitations / known tradeoffs
