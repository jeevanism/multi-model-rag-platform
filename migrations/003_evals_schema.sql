CREATE TABLE IF NOT EXISTS eval_run (
    id BIGSERIAL PRIMARY KEY,
    dataset_name TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT,
    api_base_url TEXT NOT NULL,
    total_cases INTEGER NOT NULL,
    passed_cases INTEGER NOT NULL DEFAULT 0,
    avg_latency_ms DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS eval_run_case (
    id BIGSERIAL PRIMARY KEY,
    eval_run_id BIGINT NOT NULL REFERENCES eval_run(id) ON DELETE CASCADE,
    case_id TEXT NOT NULL,
    question TEXT NOT NULL,
    expected_contains TEXT[] NOT NULL DEFAULT '{}',
    answer TEXT NOT NULL,
    citations TEXT[] NOT NULL DEFAULT '{}',
    rag_used BOOLEAN NOT NULL DEFAULT FALSE,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    passed BOOLEAN NOT NULL DEFAULT FALSE,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_eval_run_case_eval_run_id ON eval_run_case(eval_run_id);
CREATE INDEX IF NOT EXISTS idx_eval_run_case_case_id ON eval_run_case(case_id);
